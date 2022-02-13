package bridge

//Elements kafka publish elements
type Elements struct {
	ClientID  string      `json:"clientid"`
	Username  string      `json:"username"`
	Topic     string      `json:"topic"`
	Payload   interface{} `json:"payload"`
	Timestamp int64       `json:"ts"`
}

const (
	Kafka = "kafka"
)

type BridgeMQ interface {
	Load(objs map[string]string)
	AddOrUpdateObj(key string, value string)
	RemoveObj(key string)
	GetObj(key string) (string, bool)
	Publish(e *Elements) error
}

func NewBridgeMQ(name string) BridgeMQ {
	switch name {
	case Kafka:
		return InitKafka()
	default:
		return InitKafka()
	}
}
