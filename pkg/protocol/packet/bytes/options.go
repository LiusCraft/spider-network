package bytes

import (
	"github.com/liuscraft/spider-network/pkg/protocol"
)

func WithType(dataType protocol.PacketType) protocol.Option {
	return func(packet protocol.Packet) {
		packet.(*bytesProtocol).dataType = dataType
	}
}
