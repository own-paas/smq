package initialize

import (
	"fmt"
	"github.com/sestack/smq/global"
	"github.com/spf13/viper"
	"github.com/toolkits/file"
)

func InitConfig(config string) (*viper.Viper, error) {
	if config == "" {
		return nil, fmt.Errorf("use -c to specify configuration file")
	}

	if !file.IsExist(config) {
		return nil, fmt.Errorf("config file:%s is not existent.", config)
	}

	v := viper.New()
	v.SetConfigFile(config)
	v.SetDefault("cluster", "emq")
	v.SetDefault("broker", "f7637a07-a081-42d0-bd56-6643b4ee522b")
	v.SetDefault("worker", 1024)
	v.SetDefault("host", "0.0.0.0")
	v.SetDefault("port", "1883")

	v.SetDefault("etcd.endpoints", []string{"127.0.0.1:2379"})

	v.SetDefault("http.host", "0.0.0.0")
	v.SetDefault("http.port", "8080")

	v.SetDefault("bridge.type", "kafka")
	v.SetDefault("bridge.kafka.addr", []string{"127.0.0.1:9092"})

	v.SetDefault("bridge.type", "etcd")
	v.SetDefault("auth.type", "etcd")
	v.SetDefault("notify.type", "http")
	v.SetDefault("notify.server", "http://127.0.0.1/api")

	v.SetDefault("log.level", "info")
	v.SetDefault("log.format", "json")
	v.SetDefault("log.director", "log")
	v.SetDefault("log.show_line", false)
	v.SetDefault("log.log_in_console", false)
	v.SetDefault("log.encode_level", "LowercaseColorLevelEncoder")
	v.SetDefault("log.stacktrace_key", "stacktrace")

	v.SetDefault("jwt.signing-key", "McnJ+4qMWnIvMunAIiafE7TuLmJhOb2C5qANnu7StVM=")
	v.SetDefault("jwt.expires-time", 86400)
	v.SetDefault("jwt.buffer-time", 0)
	v.SetDefault("jwt.issuer", "smq")

	err := v.ReadInConfig()
	if err != nil {
		return nil, fmt.Errorf("fatal error config file: %s \n", err)
	}
	if err := v.Unmarshal(&global.CONFIG); err != nil {
		return nil, err
	}

	return v, nil
}
