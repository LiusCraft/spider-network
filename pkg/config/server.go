package config

// ServerConfig for spider-hole
type ServerConfig struct {
	BindAddr   string     `json:"bindAddr,omitempty"`
	HoleConfig HoleConfig `json:"holeConfig,omitempty"`
}

type HoleConfig struct {
	BindAddr        string          `json:"bindAddr,omitempty"`
	MaxConn         int             `json:"maxConn,omitempty"`
	Timeout         int             `json:"timeout,omitempty"`
	AcceptTimeout   int             `json:"acceptTimeout,omitempty"`
	IOTimeoutConfig IOTimeoutConfig `json:"ioTimeoutConfig,omitempty"`
	IOBufferConfig  IOBufferConfig  `json:"ioBufferConfig,omitempty"`
}
