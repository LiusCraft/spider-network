package hole

import (
	"encoding/json"

	"github.com/liuscraft/spider-network/pkg/protocol"
	"github.com/liuscraft/spider-network/pkg/protocol/packet"
)

const (
	// 打洞消息类型
	TypeRegister   = "register"    // 客户端注册
	TypePunch      = "punch"       // 打洞请求
	TypePunchReady = "punch_ready" // 打洞就绪
	TypeConnect    = "connect"     // 连接请求
)

// HoleMessage 打洞消息
type HoleMessage struct {
	Type    string          `json:"type"`              // 消息类型
	From    string          `json:"from"`              // 发送方地址
	To      string          `json:"to,omitempty"`      // 接收方地址
	Payload json.RawMessage `json:"payload,omitempty"` // 消息负载
}

// RegisterPayload 注册消息负载
type RegisterPayload struct {
	ClientID string `json:"client_id"` // 客户端ID
	Name     string `json:"name"`      // 客户端名称
}

// PunchPayload 打洞消息负载
type PunchPayload struct {
	PublicAddr  string `json:"public_addr"`  // 公网地址
	PrivateAddr string `json:"private_addr"` // 内网地址
}

// CreateHolePacket 创建打洞消息包
func CreateHolePacket(msg *HoleMessage) (protocol.Packet, error) {
	packet, err := packet.CreateProtocol(protocol.JsonType)
	if err != nil {
		return nil, err
	}

	if _, err := packet.Write(msg); err != nil {
		return nil, err
	}

	return packet, nil
}
