package server

import (
	"github.com/liuscraft/spider-network/pkg/config"
	"github.com/liuscraft/spider-network/server/handler"
)

/*
spider-hole service:
1. spider discovery
4. spider connection management
5. spider configuration management
7. spider security management
*/
type Service struct {
	holeHandler *handler.HoleHandler
}

func NewService(cfg *config.ServerConfig) (srv *Service, err error) {
	holeHandler, err := handler.NewHoleHandler(cfg.HoleConfig)
	if err != nil {
		return nil, err
	}
	srv = &Service{
		holeHandler: holeHandler,
	}
	return
}

func (s *Service) Start() error {
	return s.holeHandler.Start()
}

func (s *Service) Close() error {
	return s.holeHandler.Stop()
}
