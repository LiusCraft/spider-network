package server

import (
	"testing"

	"github.com/liuscraft/spider-network/pkg/config"
)

func TestService(t *testing.T) {
	srv := NewService(&config.ServerConfig{BindAddr: ":8080"})
	srv.listener.Close() // TODO: test
}
