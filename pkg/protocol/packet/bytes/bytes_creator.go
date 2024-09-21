package bytes

import (
	"github.com/liuscraft/spider-network/pkg/protocol"
)

type BytesCreator struct {
	DataType    protocol.PacketType
	baseOptions []protocol.Option
}

func NewBytesCreator(baseOptions ...protocol.Option) BytesCreator {
	return BytesCreator{DataType: protocol.BytesType, baseOptions: baseOptions}
}

func (b BytesCreator) NewProtocol(options ...protocol.Option) protocol.Packet {
	packet := &bytesProtocol{packetSize: 0, dataType: protocol.BytesType, body: make([]byte, 0)}
	for _, opt := range b.baseOptions {
		opt(packet)
	}
	for _, opt := range options {
		opt(packet)
	}
	return packet
}

func (b BytesCreator) Gzip(packet protocol.Packet) bool {
	return false // Bytes protocol doesn't support gzip.'
}

func (b BytesCreator) Unzip(packet protocol.Packet) bool {
	return false // Bytes protocol doesn't support gzip.'
}

func (b BytesCreator) PacketType() protocol.PacketType {
	return b.DataType
}
