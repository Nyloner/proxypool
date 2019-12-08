package spiders

import (
	"bytes"
	"fmt"
	"sync"

	"github.com/Nyloner/proxypool/common/utils"
	"github.com/Nyloner/proxypool/logs"
	"github.com/PuerkitoBio/goquery"
)

type XiciSpider struct {
	name string
}

func NewXiciSpider() ProxySpider {
	return &XiciSpider{
		name: "西刺代理",
	}
}

func (s *XiciSpider) Name() string {
	return s.name
}

func (s *XiciSpider) Crawl() []string {
	urls := []string{
		"https://www.xicidaili.com/nn/1",
		"https://www.xicidaili.com/nn/2",
	}
	wg := utils.WaitWrapper{}
	sMap := sync.Map{}
	for _, u := range urls {
		url := u
		wg.Wrap(func() {
			res, err := crawlProxies(url)
			if err != nil {
				logs.Error("XiciSpider crawlProxies fail.[url]=%#v [err]=%#v", url, err)
				return
			}
			logs.Info("XiciSpider crawlProxies success.[url]=%#v [res]=%#v", url, res)
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

func crawlProxies(url string) (res []string, err error) {
	resp, err := utils.GET(url)
	if err != nil {
		return nil, err
	}
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(resp.Content()))
	if err != nil {
		logs.Error("crawlProxies parse fail.[url]=%#v [err]=%#v", url, err)
		return nil, err
	}
	doc.Find("table#ip_list>tbody>tr").Each(func(i int, s *goquery.Selection) {
		tds := s.Find("td")
		if tds.Size() < 3 {
			return
		}
		ipNode := tds.Eq(1)
		portNode := tds.Eq(2)
		if ipNode == nil || portNode == nil {
			return
		}
		proxy := fmt.Sprintf("%s:%s", ipNode.Text(), portNode.Text())
		res = append(res, proxy)
	})
	return
}
