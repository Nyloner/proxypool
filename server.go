package main

import (
	"fmt"
	"net"

	"github.com/Nyloner/proxypool/common/redis"
	"github.com/Nyloner/proxypool/engine"
	"github.com/Nyloner/proxypool/logs"
	"github.com/Nyloner/proxypool/settings"
)

func Init() {
	err := settings.InitConfig()
	if err != nil {
		panic(err)
	}
	redis.InitRedis()
	engine.Run()
}

func StartServer() {
	server, err := net.Listen("tcp", fmt.Sprintf("%s:%d", settings.AppConfig.ServerIP, settings.AppConfig.ServerPort))
	if err != nil {
		panic(err)
	}
	logs.Info("InitServer success.[server]=%#v", server.Addr().String())
	for {
		conn, err := server.Accept()
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				logs.Info("InitServer accept temp fail.[err]=%#v", ne.Error())
				continue
			}
			logs.Info("InitServer accept err.[err]=%#v", err)
			panic(err)
		}
		go engine.HandleProxyConnect(conn)
	}
}

func main() {
	Init()
	StartServer()
}
