package engine

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Nyloner/proxypool/common/redis"
	"github.com/Nyloner/proxypool/common/utils"
	"github.com/Nyloner/proxypool/engine/spiders"
	"github.com/Nyloner/proxypool/logs"
	redisv7 "github.com/go-redis/redis/v7"
)

var ProxyCheckAPI = "https://www.nyloner.cn/checkip"
var RedisProxyPoolKey = "proxy_pool"

func Run() {
	var crawlTicker = time.NewTicker(30 * time.Second)
	go func() {
		for t := range crawlTicker.C {
			logs.Info("proxypool run crawler at %#v", t)
			go CrawlProxy()
		}
	}()
}

func CrawlProxy() {
	sMap := sync.Map{}
	wg := utils.WaitWrapper{}
	for _, spider := range spiders.Spiders {
		wg.Wrap(func() {
			res := spider.Crawl()
			logs.Info("spider run %#v success.[res]=%#v", spider.Name(), res)
			if len(res) > 0 {
				sMap.Store(spider.Name(), res)
			}
		})
	}
	wg.Wait()
	var oProxies []string
	sMap.Range(func(key, value interface{}) bool {
		proxies := value.([]string)
		oProxies = append(oProxies, proxies...)
		return true
	})
	proxyCh := make(chan string, len(oProxies))
	for _, p := range oProxies {
		proxyIP := p
		wg.Wrap(func() {
			effective := IsProxyEnable(fmt.Sprintf("http://%s", proxyIP))
			if effective {
				proxyCh <- proxyIP
			}
		})
	}
	wg.Wait()
	close(proxyCh)
	for proxyIP := range proxyCh {
		_, err := redis.ProxyRedisCli.ZAdd(RedisProxyPoolKey, &redisv7.Z{
			Member: proxyIP,
			Score:  float64(time.Now().Unix()),
		}).Result()
		if err != nil {
			logs.Warn("CrawlProxy insert proxy to pool fail.[proxyIP]=%#v [err]=%#v", proxyIP, err)
			continue
		}
		logs.Info("CrawlProxy crawl proxy success.[proxyIP]=%#v", proxyIP)
	}
	logs.Info("Run spiders success.")
}

func IsProxyEnable(proxy string) bool {
	resp, err := utils.GETByProxy(ProxyCheckAPI, proxy)
	if err != nil {
		logs.Warn("VerifyProxy fail.[proxy]=%#v [err]=%#v", proxy, err)
		return false
	}
	var proxyResp struct {
		RemoteIP string `json:"remote_ip"`
	}
	if err := json.Unmarshal(resp.Content(), &proxyResp); err != nil {
		logs.Warn("VerifyProxy parse resp fail.[proxy]=%#v [err]=%#v", proxy, err)
		return false
	}
	if strings.Contains(proxy, proxyResp.RemoteIP) {
		logs.Info("VerifyProxy success.[proxy]=%#v", proxy)
		return true
	}
	return false
}
