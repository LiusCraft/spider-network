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

func NewService(cfg *config.ServerConfig) (srv *Service, err error) {
	xl := xlog.NewLogger()
	xl.Info("spider-hole service starting...")
	xl.Infof("spider-hole service listening on %s", cfg.BindAddr)
	listener, err := net.Listen("tcp", cfg.BindAddr)
	if err != nil {
		xl.Errorf("spider-hole service listen error: %v", err)
	}
	srv = &Service{
		listener: listener,
	}
	return
}

func (s *Service) Close() error {
	return s.listener.Close()
}

func (s *Service) Run() error {
	xl := xlog.NewLogger()
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			xl.Errorf("spider-hole service accept error: %v", err)
			continue
		}
		go s.handleConn(conn)
	}
}

func (s *Service) handleConn(conn net.Conn) {
	defer conn.Close()
	// TODO: handle connection
}
