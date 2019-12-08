package utils

import (
	"bytes"
	"compress/gzip"
	"crypto/tls"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/Nyloner/proxypool/logs"
)

type Response struct {
	content    []byte
	header     http.Header
	statusCode int
}

func (r *Response) Text() string {
	switch r.header.Get("Content-Encoding") {
	case "gzip":
		reader, err := gzip.NewReader(bytes.NewReader(r.content))
		if err != nil {
			logs.Warn("GetText fail.[err]=%#v", err)
			return string(r.content)
		}
		defer reader.Close()
		buf := new(bytes.Buffer)
		buf.ReadFrom(reader)
		return buf.String()
	}
	return string(r.content)
}

func (r *Response) Content() []byte {
	switch r.header.Get("Content-Encoding") {
	case "gzip":
		reader, err := gzip.NewReader(bytes.NewReader(r.content))
		if err != nil {
			logs.Warn("GetContent gzip fail.[err]=%#v", err)
			return r.content
		}
		defer reader.Close()
		buf := new(bytes.Buffer)
		buf.ReadFrom(reader)
		return buf.Bytes()
	}
	return r.content
}

var (
	ErrInvalidResp = errors.New("invalid resp")
)

func NewHttpClient(timeout int, proxyUrl string) (*http.Client, error) {
	transPort := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	if proxyUrl != "" {
		proxy, err := url.Parse(proxyUrl)
		if err != nil {
			logs.Warn("NewHttpClient proxy parse err.[proxyUrl]=%#v [err]=%#v", proxyUrl, err)
			return nil, err
		}
		transPort.Proxy = http.ProxyURL(proxy)
	}
	return &http.Client{
		Timeout:   time.Duration(timeout) * time.Second,
		Transport: transPort,
	}, nil
}

func GET(reqUrl string) (res *Response, err error) {
	client, err := NewHttpClient(5, "")
	if err != nil {
		logs.Error("GetHttpResponse create client fail.[err]=%#v", err)
		return nil, err
	}
	return doGetRequest(client, reqUrl)
}

func GETByProxy(reqUrl string, proxy string) (res *Response, err error) {
	client, err := NewHttpClient(5, proxy)
	if err != nil {
		logs.Error("GetHttpResponseByProxy create client fail.[err]=%#v", err)
		return nil, err
	}
	return doGetRequest(client, reqUrl)
}

func doGetRequest(client *http.Client, reqUrl string) (res *Response, err error) {
	parsedUrl, err := url.Parse(reqUrl)
	if err != nil {
		logs.Error("GetHttpResponse Parse url fail.[reqUrl]=%#v [err]=%#v", reqUrl, err)
		return nil, err
	}
	req, err := http.NewRequest("GET", reqUrl, nil)
	if err != nil {
		logs.Info("GetHttpResponse NewRequest fail.[err]=%#v", err)
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/78.0.3904.87 Safari/537.36")
	req.Header.Set("Accept", " text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3")
	req.Header.Set("Host", parsedUrl.Hostname())
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8,zh-TW;q=0.7,ja;q=0.6,ko;q=0.5")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	resp, err := client.Do(req)
	if err != nil {
		logs.Error("GetHttpResponse get fail.[reqUrl]=%#v [err]=%#v", reqUrl, err.Error())
		return nil, err
	}
	defer resp.Body.Close()
	if resp == nil {
		logs.Error("GetHttpResponse resp is nil.[reqUrl]=%#v [err]=%#v", reqUrl, err)
		return nil, ErrInvalidResp
	}
	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logs.Error("GetHttpResponse read body fail.[reqUrl]=%#v [err]=%#v", reqUrl, err)
		return nil, err
	}
	res = &Response{
		content:    content,
		statusCode: resp.StatusCode,
		header:     resp.Header,
	}
	return
}
