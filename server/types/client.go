package types

import (
    "net"
    "time"
)

// ClientInfo 客户端信息
type ClientInfo struct {
    Conn       net.Conn    `json:"-"`           // 连接对象，不序列化
    ClientID   string      `json:"client_id"`   // 客户端ID
    Name       string      `json:"name"`        // 客户端名称
    PublicAddr string      `json:"public_addr"` // 公网地址
    Status     ClientStatus `json:"status"`      // 客户端状态
}

// ClientStatus 客户端状态
type ClientStatus struct {
    Connected      bool      `json:"connected"`           // 是否在线
    LastSeen      time.Time `json:"last_seen"`          // 最后在线时间
    ConnectedAt   time.Time `json:"connected_at"`       // 首次连接时间
    LastError     string    `json:"last_error"`         // 最后一次错误
    LastErrorTime time.Time `json:"last_error_time"`    // 最后一次错误时间
    PunchStatus   string    `json:"punch_status"`       // 打洞状态
    Peers         []string  `json:"peers"`              // 已连接的节点
    BytesSent     int64     `json:"bytes_sent"`         // 已发送字节数
    BytesRecv     int64     `json:"bytes_recv"`         // 已接收字节数
    P2PBytesSent  int64     `json:"p2p_bytes_sent"`     // 点对点发送字节数
    P2PBytesRecv  int64     `json:"p2p_bytes_recv"`     // 点对点接收字节数
    BytesRate     float64   `json:"bytes_rate"`         // 传输速率（字节/秒）
    P2PBytesRate  float64   `json:"p2p_bytes_rate"`     // 点对点传输速率（字节/秒）
    Latency       int64     `json:"latency"`            // 延迟（毫秒）
}

// NewClientInfo 创建新的客户端信息
func NewClientInfo(conn net.Conn, clientID, name string) *ClientInfo {
    now := time.Now()
    return &ClientInfo{
        Conn:       conn,
        ClientID:   clientID,
        Name:       name,
        PublicAddr: conn.RemoteAddr().String(),
        Status: ClientStatus{
            Connected:      true,
            LastSeen:      now,
            ConnectedAt:   now,
            LastError:     "",
            LastErrorTime: time.Time{},
            PunchStatus:   "",
            Peers:         make([]string, 0),
            BytesSent:     0,
            BytesRecv:     0,
            P2PBytesSent:  0,
            P2PBytesRecv:  0,
            BytesRate:     0,
            P2PBytesRate:  0,
            Latency:       0,
        },
    }
} 