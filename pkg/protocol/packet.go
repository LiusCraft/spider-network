package protocol

import (
	"encoding/binary"
)

// PacketHeader 数据包头部
type PacketHeader struct {
	Type   PacketType // 1字节 协议类型
	Length uint32     // 4字节 数据长度
}

const (
	HeaderSize    = 5                // 1 + 4 字节
	MaxPacketSize = 1024 * 1024 * 10 // 10MB
)

// EncodeHeader 编码包头
func EncodeHeader(typ PacketType, length uint32) []byte {
	buf := make([]byte, HeaderSize)
	buf[0] = byte(typ)
	binary.BigEndian.PutUint32(buf[1:], length)
	return buf
}

// DecodeHeader 解码包头
func DecodeHeader(data []byte) (PacketHeader, error) {
	if len(data) < HeaderSize {
		return PacketHeader{}, ErrInvalidPacket
	}

	length := binary.BigEndian.Uint32(data[1:HeaderSize])
	if length > MaxPacketSize {
		return PacketHeader{}, ErrPacketTooLarge
	}

	return PacketHeader{
		Type:   PacketType(data[0]),
		Length: length,
	}, nil
}

// NewPacket 创建指定类型的数据包
func NewPacket(typ PacketType) (Packet, error) {
	creator, ok := GetCreator(typ)
	if !ok {
		return nil, ErrUnsupportedType
	}
	return creator.NewPacket(), nil
}

// ValidatePacket 验证数据包的有效性
func ValidatePacket(packet Packet) error {
	if packet == nil {
		return ErrEmptyPacket
	}
	if packet.PacketSize() > MaxPacketSize {
		return ErrPacketTooLarge
	}
	return nil
}
