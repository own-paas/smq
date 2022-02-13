package broker

import (
	"fmt"
	"github.com/sestack/smq/global"
	"github.com/sestack/smq/model"
	"go.etcd.io/etcd/api/v3/mvccpb"
	"go.uber.org/zap"
)

func (c *Cluster) LoadAcls() {
	global.LOGGER.Debug("load acls")
	acls, err := c.GetAcls()
	if err != nil {
		return
	}
	c.broker.acl.Load(acls)
}

func (c *Cluster) GetAcls() (map[string]*model.Acl, error) {
	acls := map[string]*model.Acl{}
	objs, err := c.broker.store.All(c.aclDir)
	kvs, ok := objs.([]*mvccpb.KeyValue)
	if !ok {
		return nil, err
	}

	for _, kv := range kvs {
		acl := &model.Acl{}
		if ok := acl.Unmarshal(kv.Value); ok == nil {
			acls[acl.ID] = acl
		}
	}

	return acls, nil
}

func (c *Cluster) GetAcl(id string) (*model.Acl, error) {
	acl := &model.Acl{}

	value, err := c.broker.store.Get(fmt.Sprintf("%s/%s", c.aclDir, id))
	if err != nil {
		return nil, err
	}

	err = acl.Unmarshal(value.([]byte))
	if err != nil {
		return nil, err
	}

	return acl, nil
}

func (c *Cluster) StoreAddORUpdateAcl(obj *model.Acl) error {
	global.LOGGER.Debug("store: add or update acl")

	data, err := obj.Marshal()
	if err != nil {
		return err
	}

	return c.broker.store.Put(fmt.Sprintf("%s/%s", c.aclDir, obj.ID), string(data))
}

func (c *Cluster) AddORUpdateAcl(obj *model.Acl) {
	global.LOGGER.Debug("add or update acl")
	if c.broker.acl != nil {
		c.broker.acl.AddOrUpdateAcl(obj)
	}
}

func (c *Cluster) StoreRemoveAcl(id string) error {
	global.LOGGER.Debug("store: remove acl", zap.String("id", id))
	return c.broker.store.Del(fmt.Sprintf("%s/%s", c.aclDir, id))
}

func (c *Cluster) RemoveAcl(id string) {
	global.LOGGER.Debug("remove acl", zap.String("id", id))
	if c.broker.acl != nil {
		c.broker.acl.RemoveAcl(id)
	}
}
