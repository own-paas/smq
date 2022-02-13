package model

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
)

type Subscription struct {
	ID        string `json:"id"`
	ClientID  string `json:"client_id"`
	NodeID    string `json:"node_id"`
	Topic     string `json:"topic"`
	Qos       uint   `json:"qos"`
	Timestamp int64  `json:"timestamp"`
}

func (self *Subscription) GetID() string {
	return fmt.Sprintf("%x:%s", md5.Sum([]byte(self.Topic)), self.ClientID)
}

func (self *Subscription) Marshal() ([]byte, error) {
	return json.Marshal(self)
}

func (self *Subscription) Unmarshal(byte []byte) error {
	return json.Unmarshal(byte, self)
}
