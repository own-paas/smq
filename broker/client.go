package broker

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/sestack/smq/bridge"
	"github.com/sestack/smq/global"
	"github.com/sestack/smq/model"
	"math/rand"
	"net"
	"reflect"
	"regexp"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/sestack/smq/broker/lib/sessions"
	"github.com/sestack/smq/broker/lib/topics"
	"github.com/sestack/smq/mqtt/packets"
	"go.uber.org/zap"
)

const (
	LOCAL  = 0
	REMOTE = 1
)

const (
	_GroupTopicRegexp = `^\$share/([0-9a-zA-Z_-]+)/(.*)$`
	_QueueTopicRegexp = `^\$queue/(.*)$`
)

const (
	Connected    = 1
	Disconnected = 2
)

const (
	awaitRelTimeout int64 = 20
	retryInterval   int64 = 20
)

var (
	groupCompile = regexp.MustCompile(_GroupTopicRegexp)
	queueCompile = regexp.MustCompile(_QueueTopicRegexp)
)

type client struct {
	typ            int
	mu             sync.Mutex
	broker         *Broker
	conn           net.Conn
	info           *model.Client
	willMsg        *packets.PublishPacket
	status         int
	ctx            context.Context
	cancelFunc     context.CancelFunc
	session        *sessions.Session
	subMap         map[string]*subscription
	subMapMu       sync.RWMutex
	topicsMgr      *topics.Manager
	subs           []interface{}
	qoss           []byte
	rmsgs          []*packets.PublishPacket
	awaitingRel    map[uint16]int64
	awaitingRelMu  sync.RWMutex
	maxAwaitingRel int
	inflight       map[uint16]*inflightElem
	inflightMu     sync.RWMutex
	retryTimer     *time.Timer
	retryTimerLock sync.Mutex
}

type InflightStatus uint8

const (
	Publish InflightStatus = 0
	Pubrel  InflightStatus = 1
)

type inflightElem struct {
	status    InflightStatus
	packet    *packets.PublishPacket
	timestamp int64
}

type subscription struct {
	client    *client
	topic     string
	qos       byte
	share     bool
	groupName string
}

var (
	DisconnectedPacket = packets.NewControlPacket(packets.Disconnect).(*packets.DisconnectPacket)
	r                  = rand.New(rand.NewSource(time.Now().UnixNano()))
)

func (c *client) init() {
	c.status = Connected
	if c.typ == LOCAL && c.conn != nil {
		c.info.Status = Connected
		c.info.Timestamp = time.Now().Unix()
		remoteAddr := c.conn.RemoteAddr()
		c.info.ConnectIP = ""
		c.info.ConnectIP, _, _ = net.SplitHostPort(remoteAddr.String())
	}
	c.ctx, c.cancelFunc = context.WithCancel(context.Background())
	c.subMap = make(map[string]*subscription)
	c.topicsMgr = c.broker.topicsMgr
	c.awaitingRel = make(map[uint16]int64)
	c.inflight = make(map[uint16]*inflightElem)
}

func (c *client) readLoop() {
	nc := c.conn
	b := c.broker
	if nc == nil || b == nil {
		return
	}

	keepAlive := time.Second * time.Duration(c.info.KeepAlive)
	timeOut := keepAlive + (keepAlive / 2)

	for {
		select {
		case <-c.ctx.Done():
			return
		default:
			//add read timeout
			if keepAlive > 0 {
				if err := nc.SetReadDeadline(time.Now().Add(timeOut)); err != nil {
					global.LOGGER.Error("set read timeout error: ", zap.Error(err), zap.String("ClientID", c.info.ID))
					msg := &Message{
						client: c,
						packet: DisconnectedPacket,
					}
					b.SubmitWork(c.info.ID, msg)
					return
				}
			}

			packet, err := packets.ReadPacket(nc)
			if err != nil {
				global.LOGGER.Error("read packet error: ", zap.Error(err), zap.String("ClientID", c.info.ID))
				msg := &Message{
					client: c,
					packet: DisconnectedPacket,
				}
				b.SubmitWork(c.info.ID, msg)
				return
			}

			// if packet is disconnect from client, then need to break the read packet loop and clear will msg.
			if _, isDisconnect := packet.(*packets.DisconnectPacket); isDisconnect {
				c.willMsg = nil
				c.cancelFunc()
			}

			msg := &Message{
				client: c,
				packet: packet,
			}
			b.SubmitWork(c.info.ID, msg)
		}
	}

}

// extractPacketFields function reads a control packet and extracts only the fields
// that needs to pass on UTF-8 validation
func extractPacketFields(msgPacket packets.ControlPacket) []string {
	var fields []string

	// Get packet type
	switch msgPacket.(type) {
	case *packets.ConnackPacket:
	case *packets.ConnectPacket:
	case *packets.PublishPacket:
		packet := msgPacket.(*packets.PublishPacket)
		fields = append(fields, packet.TopicName)
		break

	case *packets.SubscribePacket:
	case *packets.SubackPacket:
	case *packets.UnsubscribePacket:
		packet := msgPacket.(*packets.UnsubscribePacket)
		fields = append(fields, packet.Topics...)
		break
	}

	return fields
}

// validatePacketFields function checks if any of control packets fields has ill-formed
// UTF-8 string
func validatePacketFields(msgPacket packets.ControlPacket) (validFields bool) {

	// Extract just fields that needs validation
	fields := extractPacketFields(msgPacket)

	for _, field := range fields {

		// Perform the basic UTF-8 validation
		if !utf8.ValidString(field) {
			validFields = false
			return
		}

		// A UTF-8 encoded string MUST NOT include an encoding of the null
		// character U+0000
		// If a receiver (Server or Client) receives a Control Packet containing U+0000
		// it MUST close the Network Connection
		// http://docs.oasis-open.org/mqtt/mqtt/v3.1.1/os/mqtt-v3.1.1-os.pdf page 14
		if bytes.ContainsAny([]byte(field), "\u0000") {
			validFields = false
			return
		}
	}

	// All fields has been validated successfully
	validFields = true

	return
}

func ProcessMessage(msg *Message) {
	c := msg.client
	ca := msg.packet
	if ca == nil {
		return
	}

	global.LOGGER.Debug("Recv message:", zap.String("message type", reflect.TypeOf(msg.packet).String()[9:]), zap.String("ClientID", c.info.ID))

	// Perform field validation
	if !validatePacketFields(ca) {

		// http://docs.oasis-open.org/mqtt/mqtt/v3.1.1/os/mqtt-v3.1.1-os.pdf
		// Page 14
		//
		// If a Server or Client receives a Control Packet
		// containing ill-formed UTF-8 it MUST close the Network Connection

		_ = c.conn.Close()

		// Update client status
		c.status = Disconnected
		c.info.Status = Disconnected

		global.LOGGER.Error("Client disconnected due to malformed packet", zap.String("ClientID", c.info.ID))

		return
	}

	switch ca.(type) {
	case *packets.ConnackPacket:
	case *packets.ConnectPacket:
	case *packets.PublishPacket:
		packet := ca.(*packets.PublishPacket)

		c.ProcessPublish(packet, msg.forward)
	case *packets.PubackPacket:
		packet := ca.(*packets.PubackPacket)
		c.inflightMu.Lock()
		if _, found := c.inflight[packet.MessageID]; found {
			delete(c.inflight, packet.MessageID)
		} else {
			global.LOGGER.Error("Duplicated PUBACK PacketId", zap.Uint16("MessageID", packet.MessageID))
		}
		c.inflightMu.Unlock()
	case *packets.PubrecPacket:
		packet := ca.(*packets.PubrecPacket)
		c.inflightMu.RLock()
		ielem, found := c.inflight[packet.MessageID]
		c.inflightMu.RUnlock()
		if found {
			if ielem.status == Publish {
				ielem.status = Pubrel
				ielem.timestamp = time.Now().Unix()
			} else if ielem.status == Pubrel {
				global.LOGGER.Error("Duplicated PUBREC PacketId", zap.Uint16("MessageID", packet.MessageID))
			}
		} else {
			global.LOGGER.Error("The PUBREC PacketId is not found.", zap.Uint16("MessageID", packet.MessageID))
		}

		pubrel := packets.NewControlPacket(packets.Pubrel).(*packets.PubrelPacket)
		pubrel.MessageID = packet.MessageID
		if err := c.WriterPacket(pubrel); err != nil {
			global.LOGGER.Error("send pubrel error, ", zap.Error(err), zap.String("ClientID", c.info.ID))
			return
		}
	case *packets.PubrelPacket:
		packet := ca.(*packets.PubrelPacket)
		_ = c.pubRel(packet.MessageID)
		pubcomp := packets.NewControlPacket(packets.Pubcomp).(*packets.PubcompPacket)
		pubcomp.MessageID = packet.MessageID
		if err := c.WriterPacket(pubcomp); err != nil {
			global.LOGGER.Error("send pubcomp error, ", zap.Error(err), zap.String("ClientID", c.info.ID))
			return
		}
	case *packets.PubcompPacket:
		packet := ca.(*packets.PubcompPacket)
		c.inflightMu.Lock()
		delete(c.inflight, packet.MessageID)
		c.inflightMu.Unlock()
	case *packets.SubscribePacket:
		packet := ca.(*packets.SubscribePacket)
		c.ProcessSubscribe(packet)
	case *packets.SubackPacket:
	case *packets.UnsubscribePacket:
		packet := ca.(*packets.UnsubscribePacket)
		c.ProcessUnSubscribe(packet)
	case *packets.UnsubackPacket:
	case *packets.PingreqPacket:
		c.ProcessPing()
	case *packets.PingrespPacket:
	case *packets.DisconnectPacket:
		c.Close()
	default:
		global.LOGGER.Info("Recv Unknow message.......", zap.String("ClientID", c.info.ID))
	}
}

func (c *client) ProcessPublish(packet *packets.PublishPacket, forward bool) {

	topic := packet.TopicName

	if global.CONFIG.Acl.Enable && c.broker.acl != nil {
		if !c.broker.acl.CheckAcl("pub", c.info.ID, c.info.UserName, c.info.ConnectIP, topic) {
			global.LOGGER.Error("Pub topic Auth failed: ", zap.String("topic", topic), zap.String("ClientID", c.info.ID))
			return
		}
	}

	switch packet.Qos {
	case QosAtMostOnce:
		c.ProcessPublishMessage(packet, forward)
	case QosAtLeastOnce:
		puback := packets.NewControlPacket(packets.Puback).(*packets.PubackPacket)
		puback.MessageID = packet.MessageID
		if err := c.WriterPacket(puback); err != nil {
			global.LOGGER.Error("send puback error, ", zap.Error(err), zap.String("ClientID", c.info.ID))
			return
		}
		c.ProcessPublishMessage(packet, forward)
	case QosExactlyOnce:
		if err := c.registerPublishPacketId(packet.MessageID); err != nil {
			return
		} else {
			pubrec := packets.NewControlPacket(packets.Pubrec).(*packets.PubrecPacket)
			pubrec.MessageID = packet.MessageID
			if err := c.WriterPacket(pubrec); err != nil {
				global.LOGGER.Error("send pubrec error, ", zap.Error(err), zap.String("ClientID", c.info.ID))
				return
			}
			c.ProcessPublishMessage(packet, forward)
		}
		return
	default:
		global.LOGGER.Error("publish with unknown qos", zap.String("ClientID", c.info.ID))
		return
	}

}

func (c *client) ProcessPublishMessage(packet *packets.PublishPacket, forward bool) {
	b := c.broker
	if b == nil {
		return
	}

	if b.config.Bridge.Enable && strings.HasPrefix(packet.TopicName, "$queue/") {
		substr := queueCompile.FindStringSubmatch(packet.TopicName)
		if len(substr) != 2 {
			return
		}

		c.broker.Bridge(&bridge.Elements{
			ClientID:  c.info.ID,
			Username:  c.info.UserName,
			Timestamp: time.Now().Unix(),
			Payload:   packet.Payload,
			Topic:     substr[1],
		})

		return
	}

	if packet.Retain {
		if err := c.topicsMgr.Retain(packet); err != nil {
			global.LOGGER.Error("Error retaining message: ", zap.Error(err), zap.String("ClientID", c.info.ID))
		}
	}

	if forward {
		publishFoward(c, packet)
		return
	}

	err := c.topicsMgr.Subscribers([]byte(packet.TopicName), packet.Qos, &c.subs, &c.qoss)
	if err != nil {
		global.LOGGER.Error("Error retrieving subscribers list: ", zap.String("ClientID", c.info.ID))
		return
	}

	// fmt.Println("psubs num: ", len(c.subs))
	if len(c.subs) == 0 {
		return
	}

	qsub := map[string][]int{}
	for i, sub := range c.subs {
		s, ok := sub.(*subscription)
		if ok {
			if s.share {
				_, ok := qsub[s.groupName]
				if ok {
					qsub[s.groupName] = append(qsub[s.groupName], i)
				} else {
					qsub[s.groupName] = []int{i}
				}
			} else {
				publish(s, packet)
			}
		}
	}

	for group, objs := range qsub {
		if len(objs) > 0 {
			idx := r.Intn(len(objs))
			sub := c.subs[qsub[group][idx]].(*subscription)
			publish(sub, packet)
		}
	}
}

func (c *client) processRemoteSubscribe(obj *model.Subscription) {
	if c.status == Disconnected {
		return
	}

	t := obj.Topic

	sub := &subscription{
		client: c,
		qos:    byte(obj.Qos),
		topic:  obj.Topic,
	}

	sub.groupName = ""
	sub.share = false
	if strings.HasPrefix(sub.topic, "$share/") {
		substr := groupCompile.FindStringSubmatch(sub.topic)
		if len(substr) != 3 {
			return
		}
		sub.share = true
		sub.groupName = substr[1]
		sub.topic = substr[2]
	}

	c.subMapMu.Lock()
	if oldSub, exist := c.subMap[t]; exist {
		_ = c.topicsMgr.Unsubscribe([]byte(oldSub.topic), oldSub)
		delete(c.subMap, t)
	}
	c.subMapMu.Unlock()

	_, err := c.topicsMgr.Subscribe([]byte(sub.topic), sub.qos, sub)
	if err != nil {
		global.LOGGER.Error("subscribe error, ", zap.Error(err), zap.String("ClientID", c.info.ID))
		return
	}

	c.subMapMu.Lock()
	c.subMap[t] = sub
	c.subMapMu.Unlock()
}

func (c *client) ProcessSubscribe(packet *packets.SubscribePacket) {
	if c.status == Disconnected {
		return
	}

	b := c.broker
	if b == nil {
		return
	}

	subTopics := packet.Topics
	qoss := packet.Qoss

	suback := packets.NewControlPacket(packets.Suback).(*packets.SubackPacket)
	suback.MessageID = packet.MessageID
	var retcodes []byte

	for i, topic := range subTopics {
		t := topic
		//check topic auth for client
		if global.CONFIG.Acl.Enable && b.acl != nil {
			if !b.acl.CheckAcl("sub", c.info.ID, c.info.UserName, c.info.ConnectIP, topic) {
				global.LOGGER.Error("Sub topic Auth failed: ", zap.String("topic", topic), zap.String("ClientID", c.info.ID))
				retcodes = append(retcodes, QosFailure)
				continue
			}
		}

		groupName := ""
		share := false
		if strings.HasPrefix(topic, "$share/") {
			substr := groupCompile.FindStringSubmatch(topic)
			if len(substr) != 3 {
				retcodes = append(retcodes, QosFailure)
				continue
			}
			share = true
			groupName = substr[1]
			topic = substr[2]
		}
		c.subMapMu.Lock()
		if oldSub, exist := c.subMap[t]; exist {
			_ = c.topicsMgr.Unsubscribe([]byte(oldSub.topic), oldSub)
			delete(c.subMap, t)
		}
		c.subMapMu.Unlock()

		sub := &subscription{
			topic:     topic,
			qos:       qoss[i],
			client:    c,
			share:     share,
			groupName: groupName,
		}

		rqos, err := c.topicsMgr.Subscribe([]byte(topic), qoss[i], sub)
		if err != nil {
			global.LOGGER.Error("subscribe error, ", zap.Error(err), zap.String("ClientID", c.info.ID))
			retcodes = append(retcodes, QosFailure)
			continue
		}

		c.subMapMu.Lock()
		c.subMap[t] = sub
		c.subMapMu.Unlock()

		_ = c.session.AddTopic(t, qoss[i])
		retcodes = append(retcodes, rqos)
		_ = c.topicsMgr.Retained([]byte(topic), &c.rmsgs)

		subData := &model.Subscription{
			NodeID:    c.broker.id,
			Topic:     t,
			ClientID:  c.info.ID,
			Qos:       uint(qoss[i]),
			Timestamp: time.Now().Unix(),
		}
		c.broker.cluster.StoreAddORUpdateSubscription(subData)

		if global.CONFIG.Notify.Enable && b.notify != nil {
			b.notify.OnSubscribe(subData)
			if subData.Topic == fmt.Sprintf("agents/%s", subData.ClientID) {
				b.notify.OnAgentConnect(c.info)
			}
		}

	}

	suback.ReturnCodes = retcodes

	err := c.WriterPacket(suback)
	if err != nil {
		global.LOGGER.Error("send suback error, ", zap.Error(err), zap.String("ClientID", c.info.ID))
		return
	}

	//process retain message
	for _, rm := range c.rmsgs {
		if err := c.WriterPacket(rm); err != nil {
			global.LOGGER.Error("Error publishing retained message:", zap.Any("err", err), zap.String("ClientID", c.info.ID))
		} else {
			global.LOGGER.Info("process retain  message: ", zap.Any("packet", packet), zap.String("ClientID", c.info.ID))
		}
	}
}

func (c *client) ProcessUnSubscribe(packet *packets.UnsubscribePacket) {
	if c.status == Disconnected {
		return
	}
	b := c.broker
	if b == nil {
		return
	}

	unSubTopics := packet.Topics

	for _, topic := range unSubTopics {
		//{
		//	//publish kafka
		//
		//	b.Publish(&bridge.Elements{
		//		ClientID:  c.info.ID,
		//		Username:  c.info.username,
		//		Action:    bridge.Unsubscribe,
		//		Timestamp: time.Now().Unix(),
		//		Topic:     topic,
		//	})
		//
		//}

		c.subMapMu.Lock()
		sub, exist := c.subMap[topic]
		if exist {
			_ = c.topicsMgr.Unsubscribe([]byte(sub.topic), sub)
			_ = c.session.RemoveTopic(topic)
			delete(c.subMap, topic)
		}
		c.subMapMu.Unlock()
		c.broker.cluster.StoreRemoveSubscription(topic, c.info.ID)

		if global.CONFIG.Notify.Enable && b.notify != nil {
			b.notify.OnUnSubscribe(map[string]interface{}{"client_id": c.info.ID, "node_id": c.info.NodeID, "topic": topic, "qos": int(sub.qos), "timestamp": time.Now().Unix()})

			if topic == fmt.Sprintf("agents/%s", c.info.ID) {
				c.info.Timestamp = time.Now().Unix()
				b.notify.OnAgentDisconnected(c.info)
			}
		}
	}

	unsuback := packets.NewControlPacket(packets.Unsuback).(*packets.UnsubackPacket)
	unsuback.MessageID = packet.MessageID

	err := c.WriterPacket(unsuback)
	if err != nil {
		global.LOGGER.Error("send unsuback error, ", zap.Error(err), zap.String("ClientID", c.info.ID))
		return
	}
}

func (c *client) ProcessPing() {
	if c.status == Disconnected {
		return
	}
	resp := packets.NewControlPacket(packets.Pingresp).(*packets.PingrespPacket)
	err := c.WriterPacket(resp)
	if err != nil {
		global.LOGGER.Error("send PingResponse error, ", zap.Error(err), zap.String("ClientID", c.info.ID))
		return
	}
}

func (c *client) Close() {
	if c.status == Disconnected {
		return
	}

	c.cancelFunc()

	c.status = Disconnected
	c.info.Status = Disconnected
	//wait for message complete
	// time.Sleep(1 * time.Second)
	// c.status = Disconnected

	b := c.broker
	//b.Publish(&bridge.Elements{
	//	ClientID:  c.info.ID,
	//	Username:  c.info.username,
	//	Action:    bridge.Disconnect,
	//	Timestamp: time.Now().Unix(),
	//})

	if c.conn != nil {
		_ = c.conn.Close()
		c.conn = nil
	}

	if b == nil {
		return
	}

	b.cluster.RemoveClient(c.info.ID)
	b.cluster.StoreRemoveClient(c.info.ID)

	c.subMapMu.RLock()
	defer c.subMapMu.RUnlock()

	unSubTopics := make([]string, 0)
	for topic, sub := range c.subMap {
		unSubTopics = append(unSubTopics, topic)

		// guard against race condition where a client gets Close() but wasn't initialized yet fully
		if sub == nil || b.topicsMgr == nil {
			continue
		}

		if err := b.topicsMgr.Unsubscribe([]byte(sub.topic), sub); err != nil {
			global.LOGGER.Error("unsubscribe error, ", zap.Error(err), zap.String("ClientID", c.info.ID))
		}

		b.cluster.StoreRemoveSubscription(topic, c.info.ID)
		if global.CONFIG.Notify.Enable && b.notify != nil {
			b.notify.OnUnSubscribe(map[string]interface{}{"client_id": c.info.ID, "node_id": c.info.NodeID, "topic": topic, "qos": int(sub.qos), "timestamp": time.Now().Unix()})
			if topic == fmt.Sprintf("agents/%s", c.info.ID) {
				c.info.Timestamp = time.Now().Unix()
				b.notify.OnAgentDisconnected(c.info)
			}
		}
	}

	//offline notification
	if global.CONFIG.Notify.Enable && b.notify != nil {
		b.notify.OnClientDisconnected(c.info)
	}

	if c.willMsg != nil {
		b.PublishMessage(c.willMsg)
	}
}

func (c *client) WriterPacket(packet packets.ControlPacket) error {
	defer func() {
		if err := recover(); err != nil {
			global.LOGGER.Error("recover error, ", zap.Any("recover", r))
		}
	}()
	if c.status == Disconnected {
		return nil
	}

	if packet == nil {
		return nil
	}
	if c.conn == nil {
		c.Close()
		return errors.New("connect lost ....")
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	return packet.Write(c.conn)
}

func (c *client) registerPublishPacketId(packetId uint16) error {
	if c.isAwaitingFull() {
		global.LOGGER.Error("Dropped qos2 packet for too many awaiting_rel", zap.Uint16("id", packetId))
		return errors.New("DROPPED_QOS2_PACKET_FOR_TOO_MANY_AWAITING_REL")
	}

	c.awaitingRelMu.Lock()
	defer c.awaitingRelMu.Unlock()
	if _, found := c.awaitingRel[packetId]; found {
		return errors.New("RC_PACKET_IDENTIFIER_IN_USE")
	}
	c.awaitingRel[packetId] = time.Now().Unix()
	time.AfterFunc(time.Duration(awaitRelTimeout)*time.Second, c.expireAwaitingRel)
	return nil
}

func (c *client) isAwaitingFull() bool {
	c.awaitingRelMu.RLock()
	defer c.awaitingRelMu.RUnlock()
	if c.maxAwaitingRel == 0 {
		return false
	}
	if len(c.awaitingRel) < c.maxAwaitingRel {
		return false
	}
	return true
}

func (c *client) expireAwaitingRel() {
	c.awaitingRelMu.Lock()
	defer c.awaitingRelMu.Unlock()
	if len(c.awaitingRel) == 0 {
		return
	}
	now := time.Now().Unix()
	for packetId, Timestamp := range c.awaitingRel {
		if now-Timestamp >= awaitRelTimeout {
			global.LOGGER.Error("Dropped qos2 packet for await_rel_timeout", zap.Uint16("id", packetId))
			delete(c.awaitingRel, packetId)
		}
	}
}

func (c *client) pubRel(packetId uint16) error {
	c.awaitingRelMu.Lock()
	defer c.awaitingRelMu.Unlock()
	if _, found := c.awaitingRel[packetId]; found {
		delete(c.awaitingRel, packetId)
	} else {
		global.LOGGER.Error("The PUBREL PacketId is not found", zap.Uint16("id", packetId))
		return errors.New("RC_PACKET_IDENTIFIER_NOT_FOUND")
	}
	return nil
}
