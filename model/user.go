package model

import (
	"encoding/json"
)

type User struct {
	ClientID string `json:"client_id"`
	Name     string `json:"name"`
	Role     uint   `json:"role"`
	Password string `json:"password"`
	CreateAt int64  `json:"create_at"`
}

func (self *User) Marshal() ([]byte, error) {
	return json.Marshal(self)
}

func (self *User) Unmarshal(byte []byte) error {
	return json.Unmarshal(byte, self)
}
