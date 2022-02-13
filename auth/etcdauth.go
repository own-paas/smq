package auth

import (
	"github.com/sestack/smq/model"
	"github.com/sestack/smq/utils"
	"sync"
)

type etcdAuth struct {
	Mu    sync.Mutex
	Users map[string]*model.User
}

func NewEtcdAuth() *etcdAuth {
	return &etcdAuth{
		Users: map[string]*model.User{},
	}
}

func (a *etcdAuth) Load(users map[string]*model.User) {
	if users == nil {
		return
	}
	a.Mu.Lock()
	defer a.Mu.Unlock()
	a.Users = users
}

func (a *etcdAuth) GetUser(name string) (user *model.User, ok bool) {
	a.Mu.Lock()
	defer a.Mu.Unlock()
	user, ok = a.Users[name]
	return
}

func (a *etcdAuth) AddOrUpdateUser(user *model.User) {
	a.Mu.Lock()
	defer a.Mu.Unlock()
	a.Users[user.Name] = user
}

func (a *etcdAuth) RemoveUser(user *model.User) {
	a.Mu.Lock()
	defer a.Mu.Unlock()
	if _, ok := a.Users[user.Name]; ok {
		delete(a.Users, user.Name)
	}
}

func (a *etcdAuth) CheckConnect(clientID, username, password string) bool {
	user, ok := a.GetUser(username)
	if !ok {
		return false
	}
	if user.ClientID != "" && user.ClientID != clientID {
		return false
	}
	return utils.ComparePassword(user.Password, password)
}

func (a *etcdAuth) CheckACL(action, clientID, username, ip, topic string) bool {
	return true
}
