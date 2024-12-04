package config

type IOBufferConfig struct {
	ReadBufferSize  int `json:"readBufferSize,omitempty"`
	WriteBufferSize int `json:"writeBufferSize,omitempty"`
}

type IOTimeoutConfig struct {
	ReadTimeout  int `json:"readTimeout,omitempty"`
	WriteTimeout int `json:"writeTimeout,omitempty"`
}
