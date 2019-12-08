package redis

import (
	"github.com/Nyloner/proxypool/settings"
	"github.com/go-redis/redis/v7"
)

var (
	ProxyRedisCli *redis.Client
)

func InitRedis() {
	option := redis.Options{
		Addr: settings.AppConfig.ProxyRedis.Server,
	}
	ProxyRedisCli = redis.NewClient(&option)
}
