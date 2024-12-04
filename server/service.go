package server

import (
	"github.com/liuscraft/spider-network/pkg/config"
	"github.com/liuscraft/spider-network/server/handler"
	"github.com/liuscraft/spider-network/server/web"
	"fmt"
	"os"
)

/*
spider-hole service:
1. spider discovery
4. spider connection management
5. spider configuration management
7. spider security management
*/
type Service struct {
	config      *config.ServerConfig
	holeHandler *handler.HoleHandler
	webServer   *web.Server
}

func NewService(cfg *config.ServerConfig) (srv *Service, err error) {
	// 创建 hole handler
	holeHandler, err := handler.NewHoleHandler(cfg.HoleConfig)
	if err != nil {
		return nil, err
	}

	// 获取当前工作目录作为基础目录
	baseDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("get working directory error: %v", err)
	}

	// 创建 web 服务器
	webServer, err := web.NewServer(holeHandler.GetClientManager(), baseDir)
	if err != nil {
		return nil, err
	}

	srv = &Service{
		config:      cfg,
		holeHandler: holeHandler,
		webServer:   webServer,
	}
	return
}

func (s *Service) Start() error {
	// 启动打洞服务
	go s.holeHandler.Start()

	// 启动心跳检测
	s.holeHandler.GetClientManager().StartHeartbeat()

	// 启动 Web 服务
	return s.webServer.Start(":8080")
}

func (s *Service) Close() error {
	return s.holeHandler.Stop()
}
