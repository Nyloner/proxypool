package engine

import (
	"testing"

	"github.com/Nyloner/proxypool/common/redis"
	"github.com/Nyloner/proxypool/settings"
)

func TestSpider(t *testing.T) {
	settings.InitConfig()
	redis.InitRedis()
	CrawlProxy()
}
