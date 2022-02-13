package bridge

import (
	"encoding/json"
	"errors"
	"github.com/sestack/smq/config"
	"github.com/sestack/smq/global"
	"strings"
	"sync"
	"time"

	"github.com/Shopify/sarama"
	"go.uber.org/zap"
)

type kafka struct {
	Mu          sync.Mutex
	deliverMap  map[string]string
	kafakConfig config.KafakConfig
	kafkaClient sarama.AsyncProducer
}

//Init init kafak client
func InitKafka() *kafka {
	global.LOGGER.Info("start connect kafka")
	c := &kafka{
		kafakConfig: global.CONFIG.Bridge.Kafka,
		deliverMap:  map[string]string{},
	}
	c.connect()
	return c
}

//connect
func (k *kafka) connect() {
	conf := sarama.NewConfig()
	conf.Version = sarama.V1_1_1_0
	kafkaClient, err := sarama.NewAsyncProducer(k.kafakConfig.Addr, conf)
	if err != nil {
		global.LOGGER.Error("create kafka async producer failed: ", zap.Error(err))
	}

	go func() {
		for err := range kafkaClient.Errors() {
			global.LOGGER.Error("send msg to kafka failed: ", zap.Error(err))
		}
	}()

	k.kafkaClient = kafkaClient
}

//Publish publish to kafka
func (k *kafka) Publish(e *Elements) error {
	key := e.Topic
	topics := []string{}

	k.Mu.Lock()
	deliverMaps := k.deliverMap
	k.Mu.Unlock()
	for reg, topic := range deliverMaps {
		match := matchTopic(reg, e.Topic)
		if match {
			topics = append(topics, topic)
		}
	}

	return k.publish(topics, key, e)
}

func (k *kafka) publish(topics []string, key string, msg *Elements) error {
	var payload []byte

	payload, ok := msg.Payload.([]byte)
	if !ok {
		data, err := json.Marshal(msg)
		if err != nil {
			return err
		}
		payload = data
	}

	for _, topic := range topics {
		select {
		case k.kafkaClient.Input() <- &sarama.ProducerMessage{
			Topic: topic,
			Key:   sarama.ByteEncoder(key),
			Value: sarama.ByteEncoder(payload),
		}:
			continue
		case <-time.After(5 * time.Second):
			return errors.New("write kafka timeout")
		}

	}

	return nil
}

func match(subTopic []string, topic []string) bool {
	if len(subTopic) == 0 {
		if len(topic) == 0 {
			return true
		}
		return false
	}

	if len(topic) == 0 {
		if subTopic[0] == "#" {
			return true
		}
		return false
	}

	if subTopic[0] == "#" {
		return true
	}

	if (subTopic[0] == "+") || (subTopic[0] == topic[0]) {
		return match(subTopic[1:], topic[1:])
	}
	return false
}

func matchTopic(subTopic string, topic string) bool {
	return match(strings.Split(subTopic, "/"), strings.Split(topic, "/"))
}

func (k *kafka) Load(objs map[string]string) {
	if objs == nil {
		return
	}
	k.Mu.Lock()
	defer k.Mu.Unlock()
	k.deliverMap = objs
}
func (k *kafka) AddOrUpdateObj(key string, value string) {
	k.Mu.Lock()
	defer k.Mu.Unlock()
	k.deliverMap[key] = value
}
func (k *kafka) RemoveObj(key string) {
	k.Mu.Lock()
	defer k.Mu.Unlock()
	if _, ok := k.deliverMap[key]; ok {
		delete(k.deliverMap, key)
	}
}
func (k *kafka) GetObj(key string) (value string, ok bool) {
	k.Mu.Lock()
	defer k.Mu.Unlock()
	value, ok = k.deliverMap[key]
	return
}
