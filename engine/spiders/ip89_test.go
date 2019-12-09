package spiders

import "testing"

func TestIP89(t *testing.T) {
	spider := NewIP89Spider()
	res := spider.Crawl()
	t.Logf("res=%#v", res)
}
