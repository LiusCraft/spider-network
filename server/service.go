package server

import (
	"fmt"
	"io"
	"net"

	"github.com/liuscraft/spider-network/pkg/config"
	"github.com/liuscraft/spider-network/pkg/protocol"
	"github.com/liuscraft/spider-network/pkg/protocol/packet_io"
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
	xl := xlog.New()
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
	xl := xlog.New()
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
	xl := xlog.WithLogId(xlog.New(), fmt.Sprintf("spider-hole-conn[%s]", conn.RemoteAddr().String()))
	defer conn.Close()
	for {
		receivePacket, err := packet_io.ReceivePacket(conn)
		if err != nil {
			if err == io.EOF {
				xl.Warnf("spider-hole-conn leave connection")
				break
			}
			xl.Errorf("read packet error: %v", err)
			return
		}
		xl.Infof("received packet: %+v", receivePacket)
		if receivePacket.PacketType() == protocol.JsonType {
			_, err2 := packet_io.WritePacket(conn, receivePacket)
			if err2 != nil {
				xl.Errorf("write response packet error: %v", err2)
				return
			}
		}
	}
}
