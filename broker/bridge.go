package broker

import (
	"github.com/sestack/smq/bridge"
	"github.com/sestack/smq/global"
	"go.uber.org/zap"
)

func (b *Broker) Bridge(e *bridge.Elements) {
	if b.bridgeMQ != nil {
		err := b.bridgeMQ.Publish(e)
		if err != nil {
			global.LOGGER.Error("send message to mq error.", zap.Error(err))
		}
	}
}
