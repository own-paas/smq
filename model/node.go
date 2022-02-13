package model

import (
	"encoding/json"
)

type Node struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Version   string `json:"version"`
	Timestamp int64  `json:"timestamp"`
}

func (self *Node) Marshal() ([]byte, error) {
	return json.Marshal(self)
}

func (self *Node) Unmarshal(byte []byte) error {
	return json.Unmarshal(byte, self)
}
