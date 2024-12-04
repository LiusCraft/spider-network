package hole

import (
	"encoding/json"

	"github.com/liuscraft/spider-network/pkg/protocol"
)

// MessageType 消息类型
type MessageType string

const (
	TypeRegister   MessageType = "register"    // 注册
	TypePunchReady MessageType = "punch_ready" // 打洞准备
	TypePunch      MessageType = "punch"       // 打洞请求
	TypeConnect    MessageType = "connect"     // 连接请求
	TypeHeartbeat  MessageType = "heartbeat"   // 心跳消息
	TypeMessage    MessageType = "message"     // 消息
)

// Message 打洞消息
type Message struct {
	Type    MessageType     `json:"type"`
	From    string          `json:"from"`
	To      string          `json:"to"`
	Payload json.RawMessage `json:"payload"`
}

// RegisterPayload 注册消息负载
type RegisterPayload struct {
	ClientID    string `json:"client_id"`
	Name        string `json:"name"`
	PublicAddr  string `json:"public_addr"`
	PrivateAddr string `json:"private_addr"`
}

// PunchPayload 打洞消息负载
type PunchPayload struct {
	PublicAddr  string `json:"public_addr"`
	PrivateAddr string `json:"private_addr"`
}

// HeartbeatPayload 心跳消息负载
type HeartbeatPayload struct {
	ClientID     string   `json:"client_id"`  // 客户端ID
	BytesSent    int64    `json:"bytes_sent"` // 已发送字节数
	BytesRecv    int64    `json:"bytes_recv"` // 已接收字节数
	P2PBytesSent int64    `json:"p2p_bytes_sent"`
	P2PBytesRecv int64    `json:"p2p_bytes_recv"`
	Peers        []string `json:"peers"`     // 当前连接的节点
	Timestamp    int64    `json:"timestamp"` // 时间戳
}

// HolePacket 打洞协议包
type HolePacket struct {
	data []byte
}

func (p *HolePacket) Read(v interface{}) (n int, err error) {
	err = json.Unmarshal(p.data, v)
	if err != nil {
		return 0, err
	}
	return len(p.data), nil
}

func (p *HolePacket) Write(v interface{}) (n int, err error) {
	data, err := json.Marshal(v)
	if err != nil {
		return 0, err
	}
	p.data = data
	return len(data), nil
}

func (p *HolePacket) Bytes() []byte {
	header := protocol.EncodeHeader(protocol.JsonType, uint32(len(p.data)))
	return append(header, p.data...)
}

func (p *HolePacket) PacketSize() int {
	return protocol.HeaderSize + len(p.data)
}

func (p *HolePacket) PacketType() protocol.PacketType {
	return protocol.JsonType
}

func (p *HolePacket) Clear() {
	p.data = p.data[:0]
}

// NewHolePacket 创建新的打洞数据包
func NewHolePacket() *HolePacket {
	return &HolePacket{
		data: make([]byte, 0),
	}
}

// CreateHolePacket 创建包含消息的打洞数据包
func CreateHolePacket(msg *Message) (protocol.Packet, error) {
	packet := NewHolePacket()
	_, err := packet.Write(msg)
	return packet, err
}
