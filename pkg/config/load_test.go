package config

import "testing"

func TestLoadFile(t *testing.T) {
	t.Run("Load ServerConfig", func(t *testing.T) {
		srvConf := &ServerConfig{}
		err := LoadFile(srvConf, "testdata/server.conf")
		if err != nil {
			t.Errorf("LoadFile() error = %v", err)
		}
		if srvConf.BindAddr != ":19730" {
			t.Errorf("LoadFile() error = %v", err)
		}
	})
}
