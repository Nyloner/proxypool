package spiders

import (
	"github.com/Nyloner/proxypool/common/utils"
	"github.com/Nyloner/proxypool/logs"
	"regexp"
)

type IP66Spider struct {
	name string
}

func NewIP66Spider() ProxySpider {
	return &IP66Spider{
		name: "66IP代理",
	}
}

func (s *IP66Spider) Name() string {
	return s.name
}

func (s *IP66Spider) Crawl() (proxyIPs []string) {
	url := "http://www.66ip.cn/nmtq.php?getnum=600&isp=0&anonymoustype=0&start=&ports=&export=&ipaddress=&area=0&proxytype=2&api=66ip"
	resp, err := utils.GET(url)
	if err != nil {
		logs.Error("IP66Spider Get html fail.[err]=%#v", err.Error())
		return nil
	}
	re, _ := regexp.Compile(`\d+\.\d+\.\d+\.\d+:\d+`)
	ips := re.FindAll(resp.Content(), -1)
	for _, ip := range ips {
		proxyIPs = append(proxyIPs, string(ip))
	}
	return proxyIPs
}
