package model

import (
	"encoding/json"
)

type Client struct {
	ID     string `json:"id"`
	NodeID string `json:"node_id"`

	HostName     string `json:"host_name"`
	ConnectIP    string `json:"connect_ip"`
	IntranetIP   string `json:"intranet_ip"`
	PublicIP     string `json:"public_ip"`
	Platform     string `json:"platform"`
	Arch         string `json:"arch"`
	OS           string `json:"os"`
	AgentVersion string `json:"agent_version"`

	UserName   string `json:"user_name"`
	ProtoName  string `json:"proto_name"`
	ProtoVer   uint   `json:"proto_ver"`
	KeepAlive  uint16 `json:"keep_alive"`
	CleanStart bool   `json:"clean_start"`
	Status     int    `json:"status"`
	Timestamp  int64  `json:"timestamp"`
}

func (self *Client) Marshal() ([]byte, error) {
	return json.Marshal(self)
}

func (self *Client) Unmarshal(byte []byte) error {
	return json.Unmarshal(byte, self)
}
