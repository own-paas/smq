package config

type AuthConfig struct {
	Enable bool   `json:"enable"`
	Type   string `json:"type"`
}
