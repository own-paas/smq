package broker

import (
	"fmt"
	"github.com/sestack/smq/global"
	"github.com/sestack/smq/model"
	"go.etcd.io/etcd/api/v3/mvccpb"
)

func (c *Cluster) AddORUpdateNode(node *model.Node) error {
	global.LOGGER.Debug("add or update node")

	data, err := node.Marshal()
	if err != nil {
		return err
	}

	return c.broker.store.Put(fmt.Sprintf("%s/%s", c.nodeDir, node.ID), string(data))
}

func (c *Cluster) RemoveNode(id string) error {
	global.LOGGER.Debug("remove node")
	return c.broker.store.Del(fmt.Sprintf("%s/%s", c.nodeDir, id))
}

func (c *Cluster) GetNode(id string) (*model.Node, error) {
	node := &model.Node{}

	value, err := c.broker.store.Get(fmt.Sprintf("%s/%s", c.nodeDir, id))
	if err != nil {
		return nil, err
	}

	err = node.Unmarshal(value.([]byte))
	if err != nil {
		return nil, err
	}

	return node, nil
}

func (c *Cluster) GetNodes() (map[string]*model.Node, error) {
	nodes := map[string]*model.Node{}
	objs, err := c.broker.store.All(c.nodeDir)
	kvs, ok := objs.([]*mvccpb.KeyValue)
	if !ok {
		return nil, err
	}

	for _, kv := range kvs {
		node := &model.Node{}
		if ok := node.Unmarshal(kv.Value); ok == nil {
			nodes[node.ID] = node
		}
	}

	return nodes, nil
}
