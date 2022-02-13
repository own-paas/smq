package broker

import (
	"fmt"
	"github.com/sestack/smq/global"
	"github.com/sestack/smq/model"
	"go.etcd.io/etcd/api/v3/mvccpb"
	"go.uber.org/zap"
)

func (c *Cluster) LoadClients() {
	global.LOGGER.Debug("load clients")

	clients, err := c.GetClients()
	if err != nil {
		global.LOGGER.Error("load nodes failed, errors:", zap.Error(err))
		return
	}

	for id, obj := range clients {
		cli := &client{
			typ:    REMOTE,
			broker: c.broker,
			info:   obj,
		}

		cli.init()
		cli.broker.clients.Store(id, c)
	}
}

func (c *Cluster) StoreAddORUpdateClient(obj *model.Client) error {
	global.LOGGER.Debug("store: add or update client")

	data, err := obj.Marshal()
	if err != nil {
		return err
	}

	return c.broker.store.Put(fmt.Sprintf("%s/%s", c.clientDir, obj.ID), string(data))
}

func (c *Cluster) AddORUpdateClient(obj *model.Client) {
	global.LOGGER.Debug("add or update client")

	client := &client{
		typ:    REMOTE,
		broker: c.broker,
		info:   obj,
	}

	client.init()
	c.broker.clients.Store(obj.ID, client)
}

func (c *Cluster) StoreRemoveClient(id string) error {
	global.LOGGER.Debug("store remove client")
	return c.broker.store.Del(fmt.Sprintf("%s/%s", c.clientDir, id))
}

func (c *Cluster) RemoveClient(id string) {
	global.LOGGER.Debug("remove client")
	_, exist := c.broker.clients.Load(id)
	if exist {
		c.broker.clients.Delete(id)
	}
}

func (c *Cluster) GetClient(id string) (*model.Client, error) {
	client := &model.Client{}

	value, err := c.broker.store.Get(fmt.Sprintf("%s/%s", c.clientDir, id))
	if err != nil {
		return nil, err
	}

	err = client.Unmarshal(value.([]byte))
	if err != nil {
		return nil, err
	}

	return client, nil
}

func (c *Cluster) GetClients() (map[string]*model.Client, error) {
	clients := map[string]*model.Client{}
	objs, err := c.broker.store.All(c.clientDir)
	kvs, ok := objs.([]*mvccpb.KeyValue)
	if !ok {
		return nil, err
	}

	for _, kv := range kvs {
		client := &model.Client{}
		if ok := client.Unmarshal(kv.Value); ok == nil {
			clients[client.ID] = client
		}
	}

	return clients, nil
}
