package engine

import (
	"testing"

	"github.com/Nyloner/proxypool/common/redis"
	"github.com/Nyloner/proxypool/settings"
)

func TestHandleProxyConnect(t *testing.T) {
	settings.InitConfig()
	redis.InitRedis()

	_, err := loadProxyIP()
	t.Logf("err=%#v", err)
}
