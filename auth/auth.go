package auth

import "github.com/sestack/smq/model"

const (
	ETCD = "etcd"
)

type Auth interface {
	Load(users map[string]*model.User)
	AddOrUpdateUser(user *model.User)
	RemoveUser(user *model.User)
	GetUser(name string) (*model.User, bool)
	CheckACL(action, clientID, username, ip, topic string) bool
	CheckConnect(clientID, username, password string) bool
}

func NewAuth(name string) Auth {
	switch name {
	case ETCD:
		return NewEtcdAuth()
	default:
		return NewEtcdAuth()
	}
}
