package settings

import (
	"fmt"
	"os"

	"github.com/Nyloner/proxypool/logs"
	"github.com/spf13/viper"
)

const (
	EnvDev  = "dev"
	EnvProd = "prod"
)

var AppConfig = ServerConfig{}

type ServerConfig struct {
	ServerIP   string
	ServerPort int
	ProxyRedis struct {
		Server string
	}
}

func InitConfig() error {
	env := EnvDev
	envConf := os.Getenv("TCP_PROXY_ENV")
	if envConf != "" {
		env = EnvProd
	}
	v := viper.New()
	v.SetConfigType("json")
	v.SetConfigName(env)
	v.AddConfigPath(".")
	v.AddConfigPath("./conf")
	v.AddConfigPath("./../conf")
	v.AddConfigPath("./../../conf")
	err := v.ReadInConfig()
	if err != nil {
		return fmt.Errorf("app_config file: %s \n", err)
	}
	logs.Info("AllSettings=%#v", v.AllSettings())
	err = v.Unmarshal(&AppConfig)
	if err != nil {
		logs.Info("InitConfig err.[err]=%#v", err)
		return err
	}
	logs.Info("InitConfig success.[AppConfig]=%#v", AppConfig)
	return nil
}
