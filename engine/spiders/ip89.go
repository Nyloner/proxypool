package spiders

import (
	"bytes"
	"fmt"
	"strings"
	"sync"

	"github.com/Nyloner/proxypool/common/utils"
	"github.com/Nyloner/proxypool/logs"
	"github.com/PuerkitoBio/goquery"
)

type IP89Spider struct {
	name string
}

func NewIP89Spider() ProxySpider {
	return &IP89Spider{
		name: "89IP代理",
	}
}

func (s *IP89Spider) Name() string {
	return s.name
}

func (s *IP89Spider) Crawl() []string {
	urls := []string{
		"http://www.89ip.cn/index_1.html",
		"http://www.89ip.cn/index_2.html",
	}
	wg := utils.WaitWrapper{}
	sMap := sync.Map{}
	for _, u := range urls {
		url := u
		wg.Wrap(func() {
			res, err := crawlIP89Proxies(url)
			if err != nil {
				logs.Error("IP89Spider crawlProxies fail.[url]=%#v [err]=%#v", url, err)
				return
			}
			logs.Info("IP89Spider crawlProxies success.[url]=%#v [res]=%#v", url, res)
			sMap.Store(url, res)
		})
	}
	wg.Wait()
	var proxyList []string
	sMap.Range(func(key, value interface{}) bool {
		values := value.([]string)
		proxyList = append(proxyList, values...)
		return true
	})
	return proxyList
}

func crawlIP89Proxies(url string) (res []string, err error) {
	resp, err := utils.GET(url)
	if err != nil {
		return nil, err
	}
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(resp.Content()))
	if err != nil {
		logs.Error("crawlIP89Proxies parse fail.[url]=%#v [err]=%#v", url, err)
		return nil, err
	}
	doc.Find("table[class=layui-table]>tbody>tr").Each(func(i int, s *goquery.Selection) {
		tds := s.Find("td")
		if tds.Size() < 3 {
			return
		}
		ipNode := tds.Eq(0)
		portNode := tds.Eq(1)
		if ipNode == nil || portNode == nil {
			return
		}
		proxy := fmt.Sprintf("%s:%s", ipNode.Text(), portNode.Text())
		replacer := strings.NewReplacer("\n", "", "\t", "", "\r", "")
		proxy = replacer.Replace(proxy)
		res = append(res, proxy)
	})
	return
}
