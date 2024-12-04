package client_mgr

import (
	"net"
	"sync"
	"time"

	"github.com/liuscraft/spider-network/pkg/xlog"
	"github.com/liuscraft/spider-network/server/types"
)

// ClientManager 客户端管理器
type ClientManager struct {
	clients sync.Map
	xl      xlog.Logger
}

func NewClientManager() *ClientManager {
	return &ClientManager{
		xl: xlog.New(),
	}
}

// AddClient 添加客户端
func (m *ClientManager) AddClient(client *types.ClientInfo) {
	m.clients.Store(client.ClientID, client)
	m.xl.Infof("Client added: %s (%s)", client.ClientID, client.Name)
}

// RemoveClient 移除客户端
func (m *ClientManager) RemoveClient(clientID string) {
	if client, ok := m.GetClient(clientID); ok {
		client.Conn.Close()
		m.clients.Delete(clientID)
		m.xl.Infof("Client removed: %s", clientID)
	}
}

// GetClient 获取客户端
func (m *ClientManager) GetClient(clientID string) (*types.ClientInfo, bool) {
	if value, ok := m.clients.Load(clientID); ok {
		return value.(*types.ClientInfo), true
	}
	return nil, false
}

// GetClients 获取所有客户端
func (m *ClientManager) GetClients() map[string]*types.ClientInfo {
	clients := make(map[string]*types.ClientInfo)
	m.clients.Range(func(key, value interface{}) bool {
		clients[key.(string)] = value.(*types.ClientInfo)
		return true
	})
	return clients
}

// UpdateClientStatus 更新客户端状态
func (m *ClientManager) UpdateClientStatus(clientID string, status types.ClientStatus) {
	client, ok := m.GetClient(clientID)
	if !ok {
		m.xl.Warnf("Trying to update status for unknown client: %s", clientID)
		return
	}

	// 计算传输速率
	timeSinceLastUpdate := status.LastSeen.Sub(client.Status.LastSeen)
	if timeSinceLastUpdate > 0 {
		bytesDelta := (status.BytesSent + status.BytesRecv) - (client.Status.BytesSent + client.Status.BytesRecv)
		status.BytesRate = float64(bytesDelta) / timeSinceLastUpdate.Seconds()
	}

	// 保持错误信息
	status.LastError = client.Status.LastError
	status.LastErrorTime = client.Status.LastErrorTime

	// 验证并更新 peers 列表
	validPeers := make([]string, 0)
	for _, peerID := range status.Peers {
		// 确保 peer 存在且不是自己
		if peerID != clientID {
			if _, exists := m.GetClient(peerID); exists {
				validPeers = append(validPeers, peerID)
			}
		}
	}
	status.Peers = validPeers

	// 更新状态
	client.Status = status

	m.xl.Debugf("Client status updated: %s, connected=%v, peers=%v",
		clientID, client.Status.Connected, client.Status.Peers)
}

// UpdateClientError 更新客户端错误状态
func (m *ClientManager) UpdateClientError(clientID string, err error) {
	client, ok := m.GetClient(clientID)
	if !ok {
		m.xl.Warnf("Trying to update error for unknown client: %s", clientID)
		return
	}

	// 更新错误信息
	client.Status.LastError = err.Error()
	client.Status.LastErrorTime = time.Now()

	m.xl.Debugf("Updated client error: %s, error=%v", clientID, err)
}

// StartHeartbeat 启动心跳检测
func (m *ClientManager) StartHeartbeat() {
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			now := time.Now()
			timeout := 30 * time.Second

			m.clients.Range(func(key, value interface{}) bool {
				client := value.(*types.ClientInfo)
				if !client.Status.Connected {
					return true
				}

				// 检查最后一次心跳时间
				if now.Sub(client.Status.LastSeen) > timeout {
					// 计算连接持续时间
					duration := now.Sub(client.Status.ConnectedAt)

					m.xl.Warnf("Client %s (%s) timed out after %v",
						client.ClientID,
						client.Name,
						duration.Round(time.Second))

					// 更新客户端状态
					client.Status.Connected = false
					client.Status.LastSeen = now
					client.Status.LastError = "Connection timed out"
					client.Status.LastErrorTime = now
					client.Status.BytesRate = 0
					client.Status.P2PBytesRate = 0
					client.Status.Latency = 0

					// 清空对等节点列表
					client.Status.Peers = make([]string, 0)
				}

				return true
			})
		}
	}()
}

// HandleDisconnect 处理客户端断开连接
func (m *ClientManager) HandleDisconnect(conn net.Conn) {
	var disconnectedClient *types.ClientInfo
	var clientID string

	// 查找断开连接的客户端
	m.clients.Range(func(key, value interface{}) bool {
		client := value.(*types.ClientInfo)
		if client.Conn == conn {
			disconnectedClient = client
			clientID = key.(string)
			return false
		}
		return true
	})

	if disconnectedClient == nil {
		return
	}

	// 更新客户端状态
	now := time.Now()
	disconnectedClient.Status.Connected = false
	disconnectedClient.Status.LastSeen = now

	// 计算连接持续时间
	duration := now.Sub(disconnectedClient.Status.ConnectedAt)

	// 记录断开连接事件
	m.xl.Warnf("Client %s (%s) disconnected after %v",
		disconnectedClient.ClientID,
		disconnectedClient.Name,
		duration.Round(time.Second))

	// 通知相关的对等节点
	for _, peerID := range disconnectedClient.Status.Peers {
		if peer, ok := m.GetClient(peerID); ok {
			// 从对等节点的列表中移除断开连接的客户端
			newPeers := make([]string, 0, len(peer.Status.Peers)-1)
			for _, id := range peer.Status.Peers {
				if id != clientID {
					newPeers = append(newPeers, id)
				}
			}
			peer.Status.Peers = newPeers
			m.xl.Debugf("Removed disconnected client %s from peer %s's peer list",
				clientID, peerID)
		}
	}

	// 清空断开连接客户端的对等节点列表
	disconnectedClient.Status.Peers = make([]string, 0)

	// 添加错误信息
	if disconnectedClient.Status.LastError == "" {
		disconnectedClient.Status.LastError = "Connection closed"
		disconnectedClient.Status.LastErrorTime = now
	}

	// 重置速率相关字段
	disconnectedClient.Status.BytesRate = 0
	disconnectedClient.Status.P2PBytesRate = 0
	disconnectedClient.Status.Latency = 0
}
