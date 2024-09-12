package config

import (
	"encoding/json"
	"os"
)

func LoadFile(config interface{}, configFileName string) (err error) {
	file, err := os.ReadFile(configFileName)
	if err != nil {
		return
	}
	return json.Unmarshal(file, config)
}
