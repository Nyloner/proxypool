package spiders

type ProxySpider interface {
	Name() string
	Crawl() []string
}

var Spiders []ProxySpider

func init() {
	Spiders = []ProxySpider{
		NewXiciSpider(),
	}
}
