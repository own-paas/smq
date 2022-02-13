package acl

import (
	"github.com/sestack/smq/model"
	"sync"
)

const (
	CLIENTID = "clientid"
	USERNAME = "username"
	IP       = "ip"
)

type etcdAcl struct {
	Mu   sync.Mutex
	Acls map[string]*model.Acl
}

func NewEtcdAcl() *etcdAcl {
	return &etcdAcl{
		Acls: map[string]*model.Acl{},
	}
}

func (a *etcdAcl) Load(acls map[string]*model.Acl) {
	if acls == nil {
		return
	}
	a.Mu.Lock()
	defer a.Mu.Unlock()
	a.Acls = acls
}

func (a *etcdAcl) GetAcl(id string) (acl *model.Acl, ok bool) {
	a.Mu.Lock()
	defer a.Mu.Unlock()
	acl, ok = a.Acls[id]
	return
}

func (a *etcdAcl) AddOrUpdateAcl(acl *model.Acl) {
	a.Mu.Lock()
	defer a.Mu.Unlock()
	a.Acls[acl.ID] = acl
}

func (a *etcdAcl) RemoveAcl(id string) {
	a.Mu.Lock()
	defer a.Mu.Unlock()
	if _, ok := a.Acls[id]; ok {
		delete(a.Acls, id)
	}
}

func (a *etcdAcl) CheckAcl(classify, clientid, username, ip, topic string) bool {
	a.Mu.Lock()
	objs := a.Acls
	defer a.Mu.Unlock()
	acls := sortByPriority(classify, objs)
	for _, acl := range acls {
		typ := acl.Type
		switch typ {
		case CLIENTID:
			if match, auth := acl.CheckWithClientID(classify, clientid, topic); match {
				return auth
			}
		case USERNAME:
			if match, auth := acl.CheckWithUsername(classify, username, topic); match {
				return auth
			}
		case IP:
			if match, auth := acl.CheckWithIP(classify, ip, topic); match {
				return auth
			}
		}
	}
	return false
}
