package engine

import (
	"fmt"
	"io"
	"math/rand"
	"net"
	"time"

	"github.com/Nyloner/proxypool/common/redis"
	"github.com/Nyloner/proxypool/logs"
)

const BufSize = 1024

func HandleProxyConnect(conn net.Conn) {
	defer conn.Close()
	proxyConn, err := createProxyConn()
	if err != nil {
		logs.Info("HandleProxyConnect loadProxyConn fail.[err]=%#v", err)
		return
	}
	defer proxyConn.Close()
	go copySocketData(conn, proxyConn)
	copySocketData(proxyConn, conn)
	logs.Info("HandleProxyConnect complete.[source]=%#v [proxy]=%#v", conn.RemoteAddr().String(), proxyConn.RemoteAddr().String())
}

func copySocketData(src net.Conn, dst net.Conn) {
	buf := make([]byte, BufSize)
	for {
		readSize, err := src.Read(buf)
		if err != nil {
			if err == io.EOF {
				return
			} else {
				logs.Info("copySocketData read fail.[src]=%#v [dst]=%#v [err]=%#v", src.RemoteAddr().String(), dst.RemoteAddr().String(), err.Error())
				return
			}
		}
		if readSize > 0 {
			writeSize, err := dst.Write(buf[0:readSize])
			if err != nil {
				logs.Info("copySocketData write fail.[src]=%#v [dst]=%#v [err]=%#v", src.RemoteAddr().String(), dst.RemoteAddr().String(), err.Error())
				return
			}
			if writeSize != readSize {
				logs.Info("copySocketData write size not equal.[src]=%#v [dst]=%#v [err]=%#v", src.RemoteAddr().String(), dst.RemoteAddr().String(), err.Error())
				return
			}
		}
	}
}

func createProxyConn() (pConn net.Conn, err error) {
	proxyIP, err := loadProxyIP()
	if err != nil {
		return nil, err
	}
	conn, err := net.DialTimeout("tcp", proxyIP, time.Second*1)
	if err != nil {
		return nil, fmt.Errorf("loadProxyConn create proxy connection fail.[err]=%#v", err.Error())
	}
	logs.Info("loadProxyConn success.[proxyIP]=%#v", proxyIP)
	return conn, nil
}

func loadProxyIP() (ip string, err error) {
	ips, err := redis.ProxyRedisCli.ZRevRange(RedisProxyPoolKey, 0, 10).Result()
	if err != nil {
		return "", err
	}
	if len(ips) == 0 {
		return "", fmt.Errorf("loadProxyIP fail,ip_pool empty")
	}
	index := rand.Int() % len(ips)
	return ips[index], nil
}
