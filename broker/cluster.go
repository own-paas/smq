package broker

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/sestack/smq/global"
	"github.com/sestack/smq/model"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
	"go.uber.org/zap"
	"strings"
	"time"
)

var (
	// TICKER ticket
	TICKER = time.Second * 3
	// TTL timeout
	TTL = int64(5)
)

// EvtType event type
type EvtType int

// EvtSrc event src
type EvtSrc int

const (
	// EventTypeNew event type new
	EventTypeNew = EvtType(0)
	// EventTypeUpdate event type update
	EventTypeUpdate = EvtType(1)
	// EventTypeDelete event type delete
	EventTypeDelete = EvtType(2)
)

const (
	// EventSrcNode node event
	EventSrcNode = EvtSrc(0)
	// EventSrcClient client event
	EventSrcClient = EvtSrc(1)
	// EventSrcSubscribe subscribe event
	EventSrcSubscription = EvtSrc(2)
	EventSrcBridge       = EvtSrc(3)
	EventSrcUser         = EvtSrc(4)
	EventSrcAcl          = EvtSrc(5)
)

// Evt event
type Evt struct {
	Src   EvtSrc
	Type  EvtType
	Key   string
	Value interface{}
}

type Cluster struct {
	sysDir          string
	clusterDir      string
	nodeDir         string
	clientDir       string
	subscriptionDir string
	bridgeDir       string
	userDir         string
	aclDir          string

	broker   *Broker
	leader   bool
	evtSysCh chan *Evt
}

func (c *Cluster) campaign() {
	for {
		s, err := concurrency.NewSession(c.broker.store.Raw().(*clientv3.Client), concurrency.WithTTL(15))
		if err != nil {
			global.LOGGER.Error("elect: new session", zap.Error(err))
			time.Sleep(time.Second * 2)
			continue
		}
		e := concurrency.NewElection(s, c.clusterDir+"/manager")
		ctx := context.TODO()

		if err = e.Campaign(ctx, c.broker.id); err != nil {
			global.LOGGER.Error("elect: campaign", zap.Error(err))
			s.Close()
			time.Sleep(1 * time.Second)
			continue
		}

		global.LOGGER.Info("elect: success", zap.String("broker", c.broker.id))
		c.leader = true

		select {
		case <-s.Done():
			c.leader = false
			global.LOGGER.Error("elect: expired", zap.String("broker", c.broker.id))
		}
	}
}

func (c *Cluster) watch() {
	etcdClient := c.broker.store.Raw().(*clientv3.Client)
	watcher := clientv3.NewWatcher(etcdClient)
	defer watcher.Close()

	ctx := etcdClient.Ctx()
	for {
		rch := watcher.Watch(ctx, c.sysDir, clientv3.WithPrefix())
		for wresp := range rch {
			if wresp.Canceled {
				return
			}

			for _, ev := range wresp.Events {
				var evtType EvtType

				switch ev.Type {
				case mvccpb.DELETE:
					evtType = EventTypeDelete
				case mvccpb.PUT:
					if ev.IsCreate() {
						evtType = EventTypeNew
					} else if ev.IsModify() {
						evtType = EventTypeUpdate
					}
				}

				var value *Evt
				key := string(ev.Kv.Key)

				if strings.HasPrefix(key, c.clientDir) {
					value = c.doWatchWithClient(evtType, ev.Kv)
				} else if strings.HasPrefix(key, c.subscriptionDir) {
					value = c.doWatchWithSubscription(evtType, ev.Kv)
				} else if strings.HasPrefix(key, c.bridgeDir) {
					value = c.doWatchWithBridge(evtType, ev.Kv)
				} else if strings.HasPrefix(key, c.userDir) {
					value = c.doWatchWithUser(evtType, ev.Kv)
				} else if strings.HasPrefix(key, c.aclDir) {
					value = c.doWatchWithAcl(evtType, ev.Kv)
				} else {
					continue
				}

				global.LOGGER.Debug("watch sys event",
					zap.String("key", key),
					zap.Int("type", int(evtType)))

				if value != nil {
					c.evtSysCh <- value
				}
			}
		}

		select {
		case <-ctx.Done():
			// server closed, return
			return
		default:
		}
	}
}

func (c *Cluster) readyToReceiveMethodWatchEvent() {
	for {
		evt := <-c.evtSysCh
		if evt.Src == EventSrcClient {
			c.doClientEvent(evt)
		} else if evt.Src == EventSrcSubscription {
			c.doSubscriptionEvent(evt)
		} else if evt.Src == EventSrcBridge {
			c.doBridgeEvent(evt)
		} else if evt.Src == EventSrcUser {
			c.doUserEvent(evt)
		} else if evt.Src == EventSrcAcl {
			c.doAclEvent(evt)
		} else {
			global.LOGGER.Warn("unknown event")
		}
	}
}

func (c *Cluster) manager() {
	for {
		time.Sleep(5 * time.Minute)
		if c.leader {
			global.LOGGER.Debug("cluster manager start")
			clientChange := false

			nodes, err := c.GetNodes()
			if err != nil {
				continue
			}

			clients, err := c.GetClients()
			if err != nil {
				continue
			}

			for cid, client := range clients {
				if _, ok := nodes[client.NodeID]; !ok {
					global.LOGGER.Debug("no node detected client", zap.String("ClientID", cid), zap.String("NodeID", client.NodeID))
					err := c.StoreRemoveClient(cid)
					if err != nil {
						global.LOGGER.Error("remove client error", zap.String("client", cid), zap.Error(err))
					} else {
						global.LOGGER.Info("remove client", zap.String("client", cid))
						clientChange = true
					}
				}
			}

			if clientChange {
				clients, err = c.GetClients()
				if err != nil {
					continue
				}
			}

			subscriptions, err := c.GetSubscriptions()
			if err != nil {
				continue
			}

			for _, v := range subscriptions {
				if _, ok := clients[v.ClientID]; !ok {
					global.LOGGER.Debug("no client detected subscription", zap.String("ClientID", v.ClientID), zap.String("Topic", v.Topic))
					c.StoreRemoveSubscription(v.Topic, v.ClientID)
				}
			}
		}
	}
}

func (c *Cluster) Start() {
	if c.broker.config.Bridge.Enable {
		c.LoadBridges()
	}

	c.LoadUsers()

	if c.broker.config.Acl.Enable {
		c.LoadAcls()
	}

	c.LoadClients()
	c.LoadSubscriptions()

	go c.readyToReceiveMethodWatchEvent()
	go c.campaign()
	go c.manager()

	c.watch()
	global.LOGGER.Error("cluster watch close or failed")
}

func (r *Cluster) doClientEvent(evt *Evt) {
	client, ok := evt.Value.(*model.Client)
	if ok {
		if evt.Type == EventTypeNew || evt.Type == EventTypeUpdate {
			r.AddORUpdateClient(client)
		} else if evt.Type == EventTypeDelete {
			r.RemoveClient(evt.Key)
		}
	}
}

func (r *Cluster) doSubscriptionEvent(evt *Evt) {
	subscription, ok := evt.Value.(*model.Subscription)
	if ok {
		if evt.Type == EventTypeNew || evt.Type == EventTypeUpdate {
			r.AddORUpdateSubscription(subscription)
		} else if evt.Type == EventTypeDelete {
			keys := strings.Split(evt.Key, ":")
			if len(keys) == 2 {
				subscription.ID = evt.Key
				subscription.Topic = keys[0]
				subscription.ClientID = keys[1]
			}
			r.RemoveSubscription(subscription)
		}
	}
}

func (r *Cluster) doBridgeEvent(evt *Evt) {
	bridgeValue, ok := evt.Value.(string)
	if ok {
		if evt.Type == EventTypeNew || evt.Type == EventTypeUpdate {
			r.AddORUpdateBridge(evt.Key, bridgeValue)
		} else if evt.Type == EventTypeDelete {
			r.RemoveBridge(evt.Key)
		}
	}
}

func (r *Cluster) doUserEvent(evt *Evt) {
	user, ok := evt.Value.(*model.User)
	if ok {
		if evt.Type == EventTypeNew || evt.Type == EventTypeUpdate {
			r.AddORUpdateUser(user)
		} else if evt.Type == EventTypeDelete {
			user.Name = evt.Key
			r.RemoveUser(user)
		}
	}
}

func (r *Cluster) doAclEvent(evt *Evt) {
	acl, ok := evt.Value.(*model.Acl)
	if ok {
		if evt.Type == EventTypeNew || evt.Type == EventTypeUpdate {
			r.AddORUpdateAcl(acl)
		} else if evt.Type == EventTypeDelete {
			r.RemoveAcl(evt.Key)
		}
	}
}

func (c *Cluster) doWatchWithClient(evtType EvtType, kv *mvccpb.KeyValue) *Evt {
	value := &model.Client{}
	if len(kv.Value) > 0 {
		if err := json.Unmarshal([]byte(kv.Value), value); err != nil {
			//log.Fatalf("client unmarshal failed, data=%s errors: %+v", string(kv.Value), err)
			return nil
		}
	}

	if value.NodeID == c.broker.id {
		return nil
	}

	return &Evt{
		Src:   EventSrcClient,
		Type:  evtType,
		Key:   strings.Replace(string(kv.Key), fmt.Sprintf("%s/", c.clientDir), "", 1),
		Value: value,
	}
}

func (c *Cluster) doWatchWithBridge(evtType EvtType, kv *mvccpb.KeyValue) *Evt {
	return &Evt{
		Src:   EventSrcBridge,
		Type:  evtType,
		Key:   strings.Replace(string(kv.Key), fmt.Sprintf("%s/", c.bridgeDir), "", 1),
		Value: string(kv.Value),
	}
}

func (c *Cluster) doWatchWithUser(evtType EvtType, kv *mvccpb.KeyValue) *Evt {
	value := &model.User{}
	if len(kv.Value) > 0 {
		if err := json.Unmarshal([]byte(kv.Value), value); err != nil {
			return nil
		}
	}

	return &Evt{
		Src:   EventSrcUser,
		Type:  evtType,
		Key:   strings.Replace(string(kv.Key), fmt.Sprintf("%s/", c.userDir), "", 1),
		Value: value,
	}
}

func (c *Cluster) doWatchWithAcl(evtType EvtType, kv *mvccpb.KeyValue) *Evt {
	value := &model.Acl{}
	if len(kv.Value) > 0 {
		if err := json.Unmarshal([]byte(kv.Value), value); err != nil {
			return nil
		}
	}

	return &Evt{
		Src:   EventSrcAcl,
		Type:  evtType,
		Key:   strings.Replace(string(kv.Key), fmt.Sprintf("%s/", c.aclDir), "", 1),
		Value: value,
	}
}

func (c *Cluster) doWatchWithSubscription(evtType EvtType, kv *mvccpb.KeyValue) *Evt {
	value := &model.Subscription{}
	if len(kv.Value) > 0 {
		if err := json.Unmarshal([]byte(kv.Value), value); err != nil {
			//log.Fatalf("subscribe unmarshal failed, data=%s errors: %+v", string(kv.Value), err)
			return nil
		}
	}

	if value.NodeID == c.broker.id {
		return nil
	}

	return &Evt{
		Src:   EventSrcSubscription,
		Type:  evtType,
		Key:   strings.Replace(string(kv.Key), fmt.Sprintf("%s/", c.subscriptionDir), "", 1),
		Value: value,
	}
}

func (c *Cluster) RegistryNode(node *model.Node, ttl int64) {
	key := fmt.Sprintf("%s/%s", c.nodeDir, node.ID)
	etcdClient, _ := c.broker.store.Raw().(*clientv3.Client)
	etcdTxn, _ := c.broker.store.Txn().(clientv3.Txn)
	data, err := node.Marshal()
	if err != nil {
		return
	}

	lessor := clientv3.NewLease(etcdClient)
	defer lessor.Close()

	ctx, cancel := context.WithTimeout(etcdClient.Ctx(), DefaultRequestTimeout)
	leaseResp, err := lessor.Grant(ctx, ttl)
	cancel()
	if err != nil {
		fmt.Println(err)
		return
	}

	keepRespChan, err := etcdClient.KeepAlive(etcdClient.Ctx(), leaseResp.ID)
	if err != nil {
		fmt.Println(err)
		return
	}

	if _, err := etcdTxn.Then(clientv3.OpPut(key, string(data), clientv3.WithLease(leaseResp.ID))).Commit(); err != nil {
		fmt.Println(err)
		return
	}

	for {
		select {
		case <-keepRespChan:
			if keepRespChan == nil {
				break
			}
		}
	}
}
