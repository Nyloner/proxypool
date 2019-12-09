package engine

import (
	"fmt"
	"io"
	"math/rand"
	"net"
	"time"

	"github.com/Nyloner/proxypool/common/redis"
	"github.com/Nyloner/proxypool/common/utils"
	"github.com/Nyloner/proxypool/logs"
)

const BufSize = 1024
const ProxyRetryMaxTimes = 3
const ProxyLoadSize = 10

func HandleProxyConnect(conn net.Conn) {
	defer conn.Close()
	proxyConn, err := createProxyConn()
	if err != nil {
		logs.Info("HandleProxyConnect loadProxyConn fail.[err]=%#v", err)
		return
	}
	defer proxyConn.Close()
	wg := utils.WaitWrapper{}
	wg.Wrap(func() {
		copySocketData(conn, proxyConn)
	})
	wg.Wrap(func() {
		copySocketData(proxyConn, conn)
	})
	wg.Wait()
	logs.Info("HandleProxyConnect complete.[source]=%#v [proxy]=%#v", conn.RemoteAddr().String(), proxyConn.RemoteAddr().String())
}

func copySocketData(src net.Conn, dst net.Conn) {
	buf := make([]byte, BufSize)
	for {
		err := src.SetReadDeadline(time.Now().Add(time.Second * 10))
		if err != nil {
			logs.Warn("copySocketData SetReadDeadline fail.[err]=%#v", err)
		}
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
			err := dst.SetWriteDeadline(time.Now().Add(time.Second * 10))
			if err != nil {
				logs.Warn("copySocketData SetWriteDeadline fail.[err]=%#v", err)
			}
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
	conn_times := 0
	for {
		if conn_times >= ProxyRetryMaxTimes {
			logs.Warn("createProxyConn fail, over max retry times.")
			break
		}
		conn, err := net.DialTimeout("tcp", proxyIP, time.Second*1)
		if err != nil {
			conn_times += 1
			logs.Warn("createProxyConn connect fail.[conn_times]=%#v [err]=%#v", conn_times, err.Error())
			continue
		}
		logs.Info("loadProxyConn success.[proxyIP]=%#v", proxyIP)
		return conn, nil
	}
	return nil, fmt.Errorf("loadProxyConn create proxy connection fail.")
}

func loadProxyIP() (ip string, err error) {
	ips, err := redis.ProxyRedisCli.ZRevRange(RedisProxyPoolKey, 0, ProxyLoadSize).Result()
	if err != nil {
		return "", err
	}
	if len(ips) == 0 {
		return "", fmt.Errorf("loadProxyIP fail,ip_pool empty")
	}
	logs.Info("loadProxyIP from redis success.[ips]=%#v", ips)
	index := rand.Int() % len(ips)
	return ips[index], nil
}
