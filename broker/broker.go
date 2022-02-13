package broker

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/sestack/smq/acl"
	"github.com/sestack/smq/auth"
	"github.com/sestack/smq/config"
	"github.com/sestack/smq/global"
	"github.com/sestack/smq/model"
	"github.com/sestack/smq/notify"
	clientv3 "go.etcd.io/etcd/client/v3"
	"net"
	"os"
	"sync"
	"time"

	"github.com/sestack/smq/broker/lib/sessions"
	"github.com/sestack/smq/broker/lib/topics"

	"github.com/sestack/smq/bridge"
	"github.com/sestack/smq/mqtt/packets"
	"github.com/sestack/smq/pool"
	"go.uber.org/zap"
)

var BROKER *Broker

const (
	Version = "v0.1"
)

type Message struct {
	client  *client
	packet  packets.ControlPacket
	forward bool
}

type Broker struct {
	id         string
	mu         sync.Mutex
	config     *config.GlobalConfig
	wpool      *pool.WorkerPool
	clients    sync.Map
	topicsMgr  *topics.Manager
	sessionMgr *sessions.Manager
	auth       auth.Auth
	acl        acl.Acl
	notify     notify.Notification
	bridgeMQ   bridge.BridgeMQ
	store      Store
	router     Router
	cluster    *Cluster
}

func NewBroker(config *config.GlobalConfig) (*Broker, error) {
	rootPrefix := fmt.Sprintf("/%s", config.Cluster)
	sysPrefix := fmt.Sprintf("%s/sys", rootPrefix)
	forwardPrefix := fmt.Sprintf("%s/forward", rootPrefix)

	b := &Broker{
		id:     config.Broker,
		config: config,
		wpool:  pool.New(config.Worker),
	}

	etcdConfig := &clientv3.Config{
		Endpoints:   config.Etcd.Endpoints,
		DialTimeout: DefaultTimeout,
	}

	if config.Etcd.User != "" {
		etcdConfig.Username = config.Etcd.User
	}

	if config.Etcd.Password != "" {
		etcdConfig.Password = config.Etcd.Password
	}

	cli, err := clientv3.New(*etcdConfig)
	if err != nil {
		return nil, err
	}

	b.store = &etcdStore{
		rootPrefix: rootPrefix,
		rawClient:  cli,
	}

	b.cluster = &Cluster{
		sysDir:          sysPrefix,
		clusterDir:      fmt.Sprintf("%s/cluster", rootPrefix),
		nodeDir:         fmt.Sprintf("%s/nodes", sysPrefix),
		clientDir:       fmt.Sprintf("%s/clients", sysPrefix),
		subscriptionDir: fmt.Sprintf("%s/subscriptions", sysPrefix),
		bridgeDir:       fmt.Sprintf("%s/bridges", sysPrefix),
		userDir:         fmt.Sprintf("%s/users", sysPrefix),
		aclDir:          fmt.Sprintf("%s/acls", sysPrefix),
		broker:          b,
		evtSysCh:        make(chan *Evt),
	}

	b.router = &etcdRouter{
		forwardDir:      forwardPrefix,
		forwardWatchDir: fmt.Sprintf("%s/%s", forwardPrefix, b.id),
		broker:          b,
	}

	b.topicsMgr, err = topics.NewManager("mem")
	if err != nil {
		global.LOGGER.Error("new topic manager error", zap.Error(err))
		return nil, err
	}

	b.sessionMgr, err = sessions.NewManager("mem")
	if err != nil {
		global.LOGGER.Error("new session manager error", zap.Error(err))
		return nil, err
	}

	b.auth = auth.NewAuth(b.config.Auth.Type)

	if b.config.Bridge.Enable {
		b.bridgeMQ = bridge.NewBridgeMQ(b.config.Bridge.Type)
	}

	if b.config.Notify.Enable {
		b.notify = notify.NewNotify(b.config.Notify.Type)
	}

	if b.config.Acl.Enable {
		b.acl = acl.NewAcl("etcd")
	}

	return b, nil
}

func (b *Broker) SubmitWork(clientId string, msg *Message) {
	if b.wpool == nil {
		b.wpool = pool.New(b.config.Worker)
	}

	b.wpool.Submit(clientId, func() {
		ProcessMessage(msg)
	})
}

func (b *Broker) Start() {
	if b == nil {
		global.LOGGER.Error("broker is null")
	}

	global.LOGGER.Debug("start cluster server")
	go b.cluster.Start()

	global.LOGGER.Debug("start router server")
	go b.router.Receive()

	if b.config.Http.Enable {
		global.LOGGER.Debug("start http server")
		go InitHTTPMoniter(b)
	}

	if b.config.Port != "" {
		global.LOGGER.Debug("start broker listen client over tcp")
		if b.config.Tls.Enable {
			go b.StartClientListening(true)
		} else {
			go b.StartClientListening(false)
		}
	}

	global.LOGGER.Debug("register node")
	time.Sleep(time.Second * 2)
	go b.cluster.RegistryNode(&model.Node{
		ID:        b.config.Broker,
		Name:      b.config.Name,
		Version:   Version,
		Timestamp: time.Now().Unix(),
	}, 30)
}

func (b *Broker) StartClientListening(Tls bool) {
	var err error
	var l net.Listener
	for {
		hp := b.config.Host + ":" + b.config.Port
		if Tls {
			tlsConfig, err := b.config.Tls.NewTLSConfig()
			if err != nil {
				global.LOGGER.Error("start tls listening client on ", zap.String("hp", hp))
				os.Exit(1)
			}
			l, err = tls.Listen("tcp", hp, tlsConfig)
			global.LOGGER.Info("start tls listening client on ", zap.String("hp", hp))
		} else {
			l, err = net.Listen("tcp", hp)
			global.LOGGER.Info("start listening client on ", zap.String("hp", hp))
		}
		if err != nil {
			global.LOGGER.Error("error listening on ", zap.Error(err))
			time.Sleep(1 * time.Second)
		} else {
			// successfully listening
			break
		}
	}
	tmpDelay := 10 * ACCEPT_MIN_SLEEP
	for {
		conn, err := l.Accept()
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				global.LOGGER.Error("temporary client accept error, sleeping",
					zap.Error(ne), zap.Duration("sleeping", tmpDelay/time.Millisecond))
				time.Sleep(tmpDelay)
				tmpDelay *= 2
				if tmpDelay > ACCEPT_MAX_SLEEP {
					tmpDelay = ACCEPT_MAX_SLEEP
				}
			} else {
				global.LOGGER.Error("accept error: %v", zap.Error(err))
			}
			continue
		}
		tmpDelay = ACCEPT_MIN_SLEEP
		go b.handleConnection(LOCAL, conn)

	}
}

func (b *Broker) DisConnClientByClientId(clientId string) {
	cli, loaded := b.clients.LoadAndDelete(clientId)
	if !loaded {
		return
	}
	conn, success := cli.(*client)
	if !success {
		return
	}
	conn.Close()
}

func (b *Broker) handleConnection(typ int, conn net.Conn) {
	//process connect packet
	packet, err := packets.ReadPacket(conn)
	if err != nil {
		global.LOGGER.Error("read connect packet error: ", zap.Error(err))
		return
	}
	if packet == nil {
		global.LOGGER.Error("received nil packet")
		return
	}
	msg, ok := packet.(*packets.ConnectPacket)
	if !ok {
		global.LOGGER.Error("received msg that was not Connect")
		return
	}

	global.LOGGER.Info("read connect from ", zap.String("clientID", msg.ClientIdentifier))

	connack := packets.NewControlPacket(packets.Connack).(*packets.ConnackPacket)
	connack.SessionPresent = msg.CleanSession
	connack.ReturnCode = msg.Validate()

	if connack.ReturnCode != packets.Accepted {
		func() {
			defer conn.Close()
			err = connack.Write(conn)
			if err != nil {
				global.LOGGER.Error("client connection error, ", zap.Error(err), zap.String("clientID", msg.ClientIdentifier))
			}
		}()
		return
	}

	if typ == LOCAL && global.CONFIG.Auth.Enable {
		if !b.CheckConnectAuth(msg.ClientIdentifier, msg.Username, string(msg.Password)) {
			connack.ReturnCode = packets.ErrRefusedNotAuthorised
			func() {
				defer conn.Close()
				err = connack.Write(conn)
				if err != nil {
					global.LOGGER.Error("client authentication failed connection closed, ", zap.Error(err), zap.String("clientID", msg.ClientIdentifier))
				}
			}()
			return
		}
	}

	err = connack.Write(conn)
	if err != nil {
		global.LOGGER.Error("send connack error, ", zap.Error(err), zap.String("clientID", msg.ClientIdentifier))
		return
	}

	willmsg := packets.NewControlPacket(packets.Publish).(*packets.PublishPacket)
	if msg.WillFlag {
		willmsg.Qos = msg.WillQos
		willmsg.TopicName = msg.WillTopic
		willmsg.Retain = msg.WillRetain
		willmsg.Payload = msg.WillMessage
		willmsg.Dup = msg.Dup
	} else {
		willmsg = nil
	}

	info := &model.Client{
		ID:         msg.ClientIdentifier,
		NodeID:     b.id,
		UserName:   msg.Username,
		KeepAlive:  msg.Keepalive,
		ProtoName:  msg.ProtocolName,
		ProtoVer:   uint(msg.ProtocolVersion),
		CleanStart: msg.CleanSession,
	}

	if msg.Other != nil {
		if err := json.Unmarshal(msg.Other, &info); err != nil {
			global.LOGGER.Error("read client info error", zap.Error(err))
		}
	}

	c := &client{
		typ:     typ,
		broker:  b,
		conn:    conn,
		info:    info,
		willMsg: willmsg,
	}

	c.init()

	err = b.getSession(c, msg, connack)
	if err != nil {
		global.LOGGER.Error("get session error: ", zap.String("clientID", c.info.ID))
		return
	}

	cid := c.info.ID

	var exist bool
	var old interface{}

	switch typ {
	case LOCAL:
		old, exist = b.clients.Load(cid)
		if exist {
			global.LOGGER.Warn("client exist, close old...", zap.String("clientID", c.info.ID))
			ol, ok := old.(*client)
			if ok {
				ol.Close()
			}
		}
		b.clients.Store(cid, c)
		b.cluster.StoreAddORUpdateClient(c.info)

		if global.CONFIG.Notify.Enable && b.notify != nil {
			b.notify.OnClientConnect(c.info)
		}
	}

	c.readLoop()
}

func (b *Broker) PublishMessage(packet *packets.PublishPacket) {
	var subs []interface{}
	var qoss []byte
	b.mu.Lock()
	err := b.topicsMgr.Subscribers([]byte(packet.TopicName), packet.Qos, &subs, &qoss)
	b.mu.Unlock()
	if err != nil {
		global.LOGGER.Error("search sub client error,  ", zap.Error(err))
		return
	}

	for _, sub := range subs {
		s, ok := sub.(*subscription)
		if ok {
			err := s.client.WriterPacket(packet)
			if err != nil {
				global.LOGGER.Error("write message error,  ", zap.Error(err))
			}
		}
	}
}
