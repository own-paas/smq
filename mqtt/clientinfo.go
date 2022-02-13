package mqtt

import "encoding/json"

type ClientInfo struct {
	HostName     string `json:"host_name"`
	IntranetIP   string `json:"intranet_ip"`
	PublicIP     string `json:"public_ip"`
	Platform     string `json:"platform"`
	Arch         string `json:"arch"`
	OS           string `json:"os"`
	AgentVersion string `json:"agent_version"`
}

func (self *ClientInfo) Marshal() ([]byte, error) {
	return json.Marshal(self)
}

func (self *ClientInfo) Unmarshal(byte []byte) error {
	return json.Unmarshal(byte, self)
}
