package notify

const (
	HTTP = "http"
	GRPC = "grpc"
)

type Notification interface {
	OnClientConnect(data interface{}) bool
	OnClientDisconnected(data interface{}) bool
	OnSubscribe(data interface{}) bool
	OnUnSubscribe(data interface{}) bool
	OnAgentConnect(data interface{}) bool
	OnAgentDisconnected(data interface{}) bool
}

func NewNotify(name string) Notification {
	switch name {
	case HTTP:
		return InitHttp()
	case GRPC:
		return InitGrpc()
	default:
		return InitHttp()
	}
}
