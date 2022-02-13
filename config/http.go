package config

type HttpConfig struct {
	Enable bool   `json:"enable"`
	Host   string `json:"host"`
	Port   string `json:"port"`
}
