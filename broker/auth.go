package broker

import (
	"github.com/sestack/smq/global"
	"go.uber.org/zap"
)

func (b *Broker) CheckConnectAuth(clientID, username, password string) bool {
	ok := b.auth.CheckConnect(clientID, username, password)
	if !ok {
		global.LOGGER.Error("authentication failed", zap.String("user", username), zap.String("clientID", clientID))
	}
	return ok
}
