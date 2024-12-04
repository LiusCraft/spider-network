package types

import "net"

// Spider 定义蜘蛛节点接口
type Spider interface {
	// GetID 获取蜘蛛ID
	GetID() string
	// GetName 获取蜘蛛名称
	GetName() string
	// GetConnection 获取连接
	GetConnection() net.Conn
	// GetStatus 获取状态
	GetStatus() *ClientStatus
	// GetPeers 获取已连接的对等节点
	GetPeers() []string
	// AddPeer 添加对等节点
	AddPeer(peerID string)
	// RemovePeer 移除对等节点
	RemovePeer(peerID string)
	// Close 关闭连接
	Close() error
}
