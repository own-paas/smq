package broker

import "github.com/sestack/smq/model"

type Router interface {
	Forward(nodeID string, data *model.Forward) error
	Receive()
}
