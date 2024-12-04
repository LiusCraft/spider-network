package config

type WebConfig struct {
    Port     int    `yaml:"port"`
    BasePath string `yaml:"base_path"`
} 