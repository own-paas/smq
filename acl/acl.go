package acl

import "github.com/sestack/smq/model"

const (
	ETCD = "etcd"
)

type Acl interface {
	Load(users map[string]*model.Acl)
	AddOrUpdateAcl(user *model.Acl)
	RemoveAcl(id string)
	GetAcl(id string) (*model.Acl, bool)
	CheckAcl(classify, clientID, username, ip, topic string) bool
}

func NewAcl(name string) Acl {
	switch name {
	case ETCD:
		return NewEtcdAcl()
	default:
		return NewEtcdAcl()
	}
}
