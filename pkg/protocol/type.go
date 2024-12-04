package protocol

import (
	"errors"
)

// PacketType 表示协议类型
type PacketType uint8

const (
	BytesType PacketType = iota
	JsonType
)

// Packet 定义数据包接口
type Packet interface {
	// Read 读取数据到目标对象
	Read(v interface{}) (n int, err error)
	// Write 写入数据
	Write(v interface{}) (n int, err error)
	// Bytes 获取原始字节
	Bytes() []byte
	// PacketSize 返回数据包大小
	PacketSize() int
	// PacketType 返回协议类型
	PacketType() PacketType
	// Clear 清理数据
	Clear()
}

// Creator 定义协议创建器接口
type Creator interface {
	// NewPacket 创建新的数据包
	NewPacket() Packet
	// PacketType 返回此创建器支持的协议类型
	PacketType() PacketType
}

// PacketType 的 String 方法
func (t PacketType) String() string {
	switch t {
	case BytesType:
		return "bytes"
	case JsonType:
		return "json"
	default:
		return "unknown"
	}
}

// 添加一些常用错误
var (
	ErrInvalidPacket   = errors.New("invalid packet")
	ErrPacketTooLarge  = errors.New("packet too large")
	ErrUnsupportedType = errors.New("unsupported packet type")
	ErrEmptyPacket     = errors.New("empty packet")
)
