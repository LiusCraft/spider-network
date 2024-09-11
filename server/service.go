package server

import (
	"net"

	"github.com/liuscraft/spider-network/pkg/config"
	"github.com/liuscraft/spider-network/pkg/xlog"
)

/*
spider-hole service:
1. spider discovery
4. spider connection management
5. spider configuration management
7. spider security management
*/
type Service struct {
	listener net.Listener
}

func NewService(cfg *config.ServerConfig) *Service {
	xl := xlog.NewLogger()
	xl.Info("spider-hole service starting...")
	listener, err := net.Listen("tcp", cfg.BindAddr)
	if err != nil {
		xl.Fatalf("spider-hole service listen error: %v", err)
	}
	return &Service{
		listener: listener,
	}
}
