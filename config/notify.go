package config

type NotifyConfig struct {
	Enable bool   `json:"enable"`
	Type   string `json:"type"`
	Server string `json:"server"`
}
