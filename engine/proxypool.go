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
	for _, p := range oProxies {
		proxyIP := p
		wg.Wrap(func() {
			effective := IsProxyEnable(fmt.Sprintf("http://%s", proxyIP))
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
			effective := IsProxyEnable(fmt.Sprintf("http://%s", proxyIP))
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

func IsProxyEnable(proxy string) bool {
	resp, err := utils.GETByProxy(ProxyCheckAPI, proxy)
	if err != nil {
		logs.Warn("IsProxyEnable fail.[proxy]=%#v [err]=%#v", proxy, err.Error())
		return false
	}
	var proxyResp struct {
		RemoteIP string `json:"remote_ip"`
	}
	if err := json.Unmarshal(resp.Content(), &proxyResp); err != nil {
		logs.Warn("IsProxyEnable parse resp fail.[proxy]=%#v [err]=%#v", proxy, err.Error())
		return false
	}
	if strings.Contains(proxy, proxyResp.RemoteIP) {
		logs.Info("IsProxyEnable success.[proxy]=%#v", proxy)
		return true
	}
	return false
}
