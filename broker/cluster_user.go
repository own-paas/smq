package broker

import (
	"fmt"
	"github.com/sestack/smq/global"
	"github.com/sestack/smq/model"
	"github.com/sestack/smq/utils"
	"go.etcd.io/etcd/api/v3/mvccpb"
	"go.uber.org/zap"
	"time"
)

func (c *Cluster) LoadUsers() {
	global.LOGGER.Debug("load users")
	users, err := c.GetUsers()
	if err != nil {
		return
	}
	c.broker.auth.Load(users)

	if _, ok := users["admin"]; !ok {
		password, _ := utils.HashSaltPassword("admin")
		c.StoreAddORUpdateUser(&model.User{
			Name:     "admin",
			Role:     1,
			Password: password,
			CreateAt: time.Now().Unix(),
		})
	}
}

func (c *Cluster) GetUsers() (map[string]*model.User, error) {
	users := map[string]*model.User{}
	objs, err := c.broker.store.All(c.userDir)
	kvs, ok := objs.([]*mvccpb.KeyValue)
	if !ok {
		return nil, err
	}

	for _, kv := range kvs {
		user := &model.User{}
		if ok := user.Unmarshal(kv.Value); ok == nil {
			users[user.Name] = user
		}
	}

	return users, nil
}

func (c *Cluster) GetUser(name string) (*model.User, error) {
	user := &model.User{}

	value, err := c.broker.store.Get(fmt.Sprintf("%s/%s", c.userDir, name))
	if err != nil {
		return nil, err
	}

	err = user.Unmarshal(value.([]byte))
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (c *Cluster) StoreAddORUpdateUser(obj *model.User) error {
	global.LOGGER.Debug("store: add or update user")

	data, err := obj.Marshal()
	if err != nil {
		return err
	}

	return c.broker.store.Put(fmt.Sprintf("%s/%s", c.userDir, obj.Name), string(data))
}

func (c *Cluster) AddORUpdateUser(obj *model.User) {
	global.LOGGER.Debug("add or update user")
	if c.broker.auth != nil {
		c.broker.auth.AddOrUpdateUser(obj)
	}
}

func (c *Cluster) StoreRemoveUser(user string) error {
	global.LOGGER.Debug("store: remove user", zap.String("name", user))
	return c.broker.store.Del(fmt.Sprintf("%s/%s", c.userDir, user))
}

func (c *Cluster) RemoveUser(obj *model.User) {
	global.LOGGER.Debug("remove user")
	if c.broker.auth != nil {
		c.broker.auth.RemoveUser(obj)
	}
}
