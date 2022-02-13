package broker

import (
	"fmt"
	"github.com/sestack/smq/global"
	"go.etcd.io/etcd/api/v3/mvccpb"
	"strings"
)

func (c *Cluster) LoadBridges() {
	global.LOGGER.Debug("load bridges")
	objs, err := c.GetBridges()
	if err != nil {
		return
	}
	c.broker.bridgeMQ.Load(objs)
}

func (c *Cluster) GetBridges() (map[string]string, error) {
	bridges := map[string]string{}
	objs, err := c.broker.store.All(c.bridgeDir)
	kvs, ok := objs.([]*mvccpb.KeyValue)
	if !ok {
		return nil, err
	}

	for _, kv := range kvs {
		kPath := strings.Split(string(kv.Key), "/")
		bridges[kPath[len(kPath)-1]] = string(kv.Value)
	}

	return bridges, nil
}

func (c *Cluster) GetBridge(name string) (string, error) {
	value, err := c.broker.store.Get(fmt.Sprintf("%s/%s", c.bridgeDir, name))
	if err != nil {
		return "", err
	}

	return string(value.([]byte)), nil
}

func (c *Cluster) StoreAddORUpdateBridge(key string, value string) error {
	global.LOGGER.Debug("store: add or update bridge")

	return c.broker.store.Put(fmt.Sprintf("%s/%s", c.bridgeDir, key), value)
}

func (c *Cluster) AddORUpdateBridge(key string, value string) {
	global.LOGGER.Debug("add or update bridge")
	if c.broker.bridgeMQ != nil {
		c.broker.bridgeMQ.AddOrUpdateObj(key, value)
	}
}

func (c *Cluster) StoreRemoveBridge(key string) error {
	global.LOGGER.Debug("store:remove bridge")

	return c.broker.store.Del(fmt.Sprintf("%s/%s", c.bridgeDir, key))
}

func (c *Cluster) RemoveBridge(key string) {
	global.LOGGER.Debug("remove bridge")
	if c.broker.bridgeMQ != nil {
		c.broker.bridgeMQ.RemoveObj(key)
	}
}
