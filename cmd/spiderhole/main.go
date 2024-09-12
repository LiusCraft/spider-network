package main

import (
	"os"

	"github.com/liuscraft/spider-network/pkg/config"
	"github.com/liuscraft/spider-network/pkg/xlog"
	"github.com/liuscraft/spider-network/server"
)

func main() {
	xl := xlog.NewLogger()
	serverConfig := &config.ServerConfig{}
	err := config.LoadFile(serverConfig, "server.conf")
	if err != nil {
		xl.Fatalf("load config error: %v", err)
	}
	srv, err := server.NewService(serverConfig)
	if err != nil {
		os.Exit(1)
	}
	srv.Run()
}
