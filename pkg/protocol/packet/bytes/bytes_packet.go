package bytes

import (
	"github.com/liuscraft/spider-network/pkg/errors"
	"github.com/liuscraft/spider-network/pkg/protocol"
)

type bytesProtocol struct {
	packetSize int    // packet real size
	body       []byte // packet data
	dataType   protocol.PacketType
}

func newBytesProtocol(dataType protocol.PacketType) *bytesProtocol {
	return &bytesProtocol{
		packetSize: 0,
		dataType:   dataType,
		body:       make([]byte, 0),
	}
}

func (b *bytesProtocol) Write(p interface{}) (n int, err error) {
	bytes, ok := p.([]byte)
	if !ok {
		return 0, err
	}
	b.packetSize += len(bytes)
	b.body = append(b.body, bytes...)
	return len(bytes), nil
}

func (b *bytesProtocol) Read(p interface{}) (n int, err error) {
	o, ok := p.(protocol.Packet)
	if !ok {
		return 0, errors.ErrToPacketTypeNotImplemented
	}
	return o.ToPacket(b.body)
}

func (b *bytesProtocol) ToPacket(p []byte) (n int, err error) {
	b.body = p
	b.packetSize = len(b.body)
	return b.PacketSize(), nil
}

// protocol to []byte
func (b *bytesProtocol) Bytes() []byte {
	return b.body
}

func (b *bytesProtocol) PacketSize() int {
	return b.packetSize
}

func (b *bytesProtocol) PacketType() protocol.PacketType {
	return b.dataType
}

func (b *bytesProtocol) String() string {
	return string(b.body)
}

func (b *bytesProtocol) Clear() {
	b.packetSize = 0
	b.body = make([]byte, 0)
}
