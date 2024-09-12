package server

import (
	"testing"

	"github.com/liuscraft/spider-network/pkg/config"
)

func TestService(t *testing.T) {
	t.Run("Create Service", func(t *testing.T) {
		srv, err := NewService(&config.ServerConfig{BindAddr: ":8080"})
		if err != nil {
			t.Error(err)
		}
		srv.listener.Close() // TODO: test
	})
}
