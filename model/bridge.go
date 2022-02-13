package model

import (
	"encoding/json"
)

type Bridge struct {
	Topic string `json:"topic"`
	Queue string `json:"queue"`
}

func (self *Bridge) Marshal() ([]byte, error) {
	return json.Marshal(self)
}

func (self *Bridge) Unmarshal(byte []byte) error {
	return json.Unmarshal(byte, self)
}
