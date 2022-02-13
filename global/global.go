package global

import (
	"github.com/sestack/smq/config"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var (
	VP     *viper.Viper
	LOGGER *zap.Logger
	CONFIG *config.GlobalConfig
)
