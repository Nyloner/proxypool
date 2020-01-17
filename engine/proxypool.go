package engine

import (
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

// var ProxyCheckAPI = "http://pv.sohu.com/cityjson"
// var ProxyCheckAPI = "http://dev.kdlapi.com/testproxy"
var ProxyCheckAPI = "https://123.206.23.237/checkip"
var RedisProxyPoolKey = "proxy_pool"
var CrawlTimeInterval = 120
var VerifyTimeInterval = 30

func Run() {
	var crawlTicker = time.NewTicker(time.Duration(CrawlTimeInterval) * time.Second)
	go CrawlProxy()
	go func() {
		for t := range crawlTicker.C {
			logs.Info("proxypool run crawler at %#v", t.String())
			go CrawlProxy()
		}
	}()
	var verifyTicker = time.NewTicker(time.Duration(VerifyTimeInterval) * time.Second)
	go VerifyProxy()
	go func() {
		for t := range verifyTicker.C {
			logs.Info("proxypool run verify at %#v", t.String())
			go VerifyProxy()
		}
	}()
}

func CrawlProxy() {
	sMap := sync.Map{}
	wg := utils.WaitWrapper{}
	for _, s := range spiders.Spiders {
		spider := s
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
	limitC := make(chan int, 100)
	for _, p := range oProxies {
		proxyIP := p
		wg.Wrap(func() {
			limitC <- 1
			defer func() {
				<-limitC
			}()
			effective := IsProxyAvailable(proxyIP)
			if effective {
				_, err := redis.ProxyRedisCli.ZAdd(RedisProxyPoolKey, &redisv7.Z{
					Member: proxyIP,
					Score:  float64(time.Now().Unix()),
				}).Result()
				if err != nil {
					logs.Warn("CrawlProxy insert proxy to pool fail.[proxyIP]=%#v [err]=%#v", proxyIP, err)
					return
				}
				logs.Info("CrawlProxy add proxy success.[proxyIP]=%#v", proxyIP)
			}
		})
	}
	wg.Wait()
	close(limitC)
	logs.Info("Run spiders success.")
}

func VerifyProxy() {
	ips, err := redis.ProxyRedisCli.ZRange(RedisProxyPoolKey, 0, -1).Result()
	if err != nil {
		logs.Warn("ProxyVerify load proxypool fail.[err]=%#v", err)
		return
	}
	wg := utils.WaitWrapper{}
	for _, ip := range ips {
		proxyIP := ip
		wg.Wrap(func() {
			effective := IsProxyAvailable(proxyIP)
			if !effective {
				_, err := redis.ProxyRedisCli.ZRem(RedisProxyPoolKey, proxyIP).Result()
				if err != nil {
					logs.Warn("ProxyVerify remove proxy fail.[proxyIP]=%#v [err]=%#v", proxyIP, err)
				}
				logs.Info("ProxyVerify remove disabled proxy success.[proxyIP]=%#v", proxyIP)
				return
			}
			_, err := redis.ProxyRedisCli.ZAdd(RedisProxyPoolKey, &redisv7.Z{
				Member: proxyIP,
				Score:  float64(time.Now().Unix()),
			}).Result()
			if err != nil {
				logs.Warn("ProxyVerify add proxy fail.[proxyIP]=%#v [err]=%#v", proxyIP, err)
				return
			}
			logs.Info("ProxyVerify verify proxy success.[proxyIP]=%#v", proxyIP)
		})
	}
	wg.Wait()
}

func IsProxyAvailable(proxyAddress string) bool {
	httpProxy := fmt.Sprintf("http://%s", proxyAddress)
	resp, err := utils.GETByProxy(ProxyCheckAPI, httpProxy)
	if err != nil {
		logs.Warn("IsProxyAvailable run fail.[proxy]=%#v [err]=%#v", httpProxy, err.Error())
		return false
	}
	content := resp.Text()
	IP := strings.Split(proxyAddress, ":")[0]
	if strings.Contains(content, IP) {
		logs.Info("IsProxyAvailable success.[proxy]=%#v [content]=%s", httpProxy, content)
		return true
	}
	logs.Info("IsProxyAvailable verify fail.[proxy]=%#v [content]=%s", httpProxy, content)
	return false
}
