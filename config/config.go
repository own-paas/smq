package config

import "github.com/sestack/smq/utils/jwt"

type GlobalConfig struct {
	Cluster string        `json:"cluster"`
	Broker  string        `json:"broker"`
	Name    string        `json:"name"`
	Worker  int           `json:"worker"`
	Host    string        `json:"host"`
	Port    string        `json:"port"`
	JWT     jwt.JWTConfig `json:"jwt"`
	Log     Zap           `json:"log"`
	Etcd    ETCDConfig    `json:"etcd"`
	Http    HttpConfig    `json:"http"`
	Bridge  BridgeConfig  `json:"bridge"`
	Auth    AuthConfig    `json:"auth"`
	Acl     AclConfig     `json:"acl"`
	Notify  NotifyConfig  `json:"notify"`
	Tls     TlsConfig     `json:"tls"`
}
