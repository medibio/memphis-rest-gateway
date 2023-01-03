package main

import (
	"fmt"
	"http-proxy/conf"
	"http-proxy/logger"
	"http-proxy/router"
	"time"

	"github.com/memphisdev/memphis.go"
)

var configuration = conf.GetConfig()

func main() {
	configuration := conf.GetConfig()
	var conn *memphis.Conn
	ticker := time.NewTicker(1 * time.Second)
	for {
		select {
		case <-ticker.C:
			var err error
			conn, err = memphis.Connect(configuration.MEMPHIS_HOST, configuration.ROOT_USER, configuration.CONNECTION_TOKEN)
			if err == nil {
				ticker.Stop()
				goto serverInit
			} else {
				fmt.Printf("Awaiting to establish connection with Memphis - %v\n", err.Error())
			}
		}
	}

serverInit:
	l, err := logger.CreateLogger(configuration.MEMPHIS_HOST, configuration.ROOT_USER, configuration.CONNECTION_TOKEN)
	if err != nil {
		panic("Logger creation failed - " + err.Error())
	}

	app := router.SetupRoutes(conn, l)
	l.Noticef("Memphis Http Proxy is up and running")
	app.Listen(":" + configuration.HTTP_PORT)
}
