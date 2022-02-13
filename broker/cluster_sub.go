package broker

import (
	"crypto/md5"
	"fmt"
	"github.com/sestack/smq/global"
	"github.com/sestack/smq/model"
	"go.etcd.io/etcd/api/v3/mvccpb"
	"go.uber.org/zap"
)

func (c *Cluster) LoadSubscriptions() {
	global.LOGGER.Debug("load Subscriptions")

	subscriptions, err := c.GetSubscriptions()
	if err != nil {
		global.LOGGER.Error("load subscriptions failed, errors:", zap.Error(err))
		return
	}

	for _, obj := range subscriptions {
		cli, exist := c.broker.clients.Load(obj.ClientID)
		if !exist {
			return
		}

		cliObj, ok := cli.(*client)
		if !ok {
			return
		}

		cliObj.processRemoteSubscribe(obj)
	}

}

func (c *Cluster) StoreAddORUpdateSubscription(subscription *model.Subscription) error {
	global.LOGGER.Debug("add or update subscription", zap.String("topic", subscription.Topic), zap.String("client", subscription.ClientID))
	subscription.ID = subscription.GetID()
	data, err := subscription.Marshal()
	if err != nil {
		return err
	}
	return c.broker.store.Put(fmt.Sprintf("%s/%s", c.subscriptionDir, subscription.ID), string(data))
}

func (c *Cluster) AddORUpdateSubscription(subscription *model.Subscription) {
	global.LOGGER.Debug("add or update subscription")
	cli, exist := c.broker.clients.Load(subscription.ClientID)
	if !exist {
		return
	}

	cliObj, ok := cli.(*client)
	if !ok {
		return
	}

	if cliObj.typ == REMOTE {
		cliObj.processRemoteSubscribe(subscription)
	}
}

func (c *Cluster) StoreRemoveSubscription(topic string, clientID string) error {
	global.LOGGER.Debug("store: remove subscription", zap.String("topic", topic), zap.String("client", clientID))
	return c.broker.store.Del(fmt.Sprintf("%s/%x:%s", c.subscriptionDir, md5.Sum([]byte(topic)), clientID))
}

func (c *Cluster) RemoveSubscription(subscription *model.Subscription) {
	global.LOGGER.Debug("remove subscription")
	cli, exist := c.broker.clients.Load(subscription.ClientID)
	if !exist {
		return
	}

	cliObj, ok := cli.(*client)
	if !ok {
		return
	}

	cliObj.subMapMu.Lock()
	sub, exist := cliObj.subMap[subscription.Topic]
	if exist {
		_ = cliObj.topicsMgr.Unsubscribe([]byte(sub.topic), sub)
		_ = cliObj.session.RemoveTopic(subscription.Topic)
		delete(cliObj.subMap, subscription.Topic)
	}
	cliObj.subMapMu.Unlock()
}

func (c *Cluster) GetSubscription(id string) (*model.Subscription, error) {
	subscription := &model.Subscription{}

	value, err := c.broker.store.Get(fmt.Sprintf("%s/%s", c.subscriptionDir, id))
	if err != nil {
		return nil, err
	}

	err = subscription.Unmarshal(value.([]byte))
	if err != nil {
		return nil, err
	}

	return subscription, nil
}

func (c *Cluster) GetSubscriptions() (map[string]*model.Subscription, error) {
	subscriptions := map[string]*model.Subscription{}
	objs, err := c.broker.store.All(c.subscriptionDir)
	kvs, ok := objs.([]*mvccpb.KeyValue)
	if !ok {
		return nil, err
	}

	for _, kv := range kvs {
		sbscription := &model.Subscription{}
		if ok := sbscription.Unmarshal(kv.Value); ok == nil {
			subscriptions[sbscription.ID] = sbscription
		}
	}

	return subscriptions, nil
}
