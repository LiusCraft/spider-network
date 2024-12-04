package main

import (
	"github.com/liuscraft/spider-network/pkg/config"
	"github.com/liuscraft/spider-network/pkg/xlog"
	"github.com/liuscraft/spider-network/server"
)

func main() {
	xl := xlog.New()
	xl.Info("Starting spider hole server...")

	// 加载配置
	cfg := &config.ServerConfig{
		HoleConfig: config.HoleConfig{
			BindAddr: ":19730",
		},
	}

	// 创建服务
	srv, err := server.NewService(cfg)
	if err != nil {
		xl.Fatalf("Failed to create server: %v", err)
	}

	// 启动服务
	xl.Info("Starting server on :19730")
	if err := srv.Start(); err != nil {
		xl.Fatalf("Failed to start server: %v", err)
	}

	select {}
}
