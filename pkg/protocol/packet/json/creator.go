package json

import (
	"github.com/liuscraft/spider-network/pkg/protocol"
)

type JsonCreator struct {
}

func NewJsonCreator() protocol.Creator {
	return &JsonCreator{}
}

func (j *JsonCreator) NewProtocol(options ...protocol.Option) protocol.Packet {
	packet := &JsonProtocol{packetSize: 0, body: make([]byte, 0)}
	for _, opt := range options {
		opt(packet)
	}
	return packet
}

func (j *JsonCreator) PacketType() protocol.PacketType {
	return protocol.JsonType
}

func (j *JsonCreator) Gzip(packet protocol.Packet) bool {
	return false // JSON protocol doesn't support gzip.'
}

func (j *JsonCreator) Unzip(packet protocol.Packet) bool {
	return false // JSON protocol doesn't support gzip.'
}
