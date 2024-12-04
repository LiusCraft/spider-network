package types

import "github.com/liuscraft/spider-network/pkg/protocol"

// Handler 定义基础处理器接口
type Handler interface {
    Start() error
    Stop() error
    Handle(packet protocol.Packet) error
}

// ConnectionHandler 定义连接处理器接口
type ConnectionHandler interface {
    Handler
    GetClients() map[string]*ClientInfo
    GetClient(id string) (*ClientInfo, bool)
    GetClientStatus(id string) (*ClientStatus, bool)
} 