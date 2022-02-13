package config

type BridgeConfig struct {
	Enable bool        `json:"enable"`
	Type   string      `json:"type"`
	Kafka  KafakConfig `json:"kafka"`
}
