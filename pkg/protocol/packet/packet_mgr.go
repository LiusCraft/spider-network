package packet

import (
	"fmt"

	"github.com/liuscraft/spider-network/pkg/protocol"
	"github.com/liuscraft/spider-network/pkg/protocol/packet/bytes"
	"github.com/liuscraft/spider-network/pkg/protocol/packet/json"
)

var (
	protocols = map[protocol.PacketType]protocol.Creator{}
)

func init() {
	_ = RegisterProtocol(bytes.NewBytesCreator())
	_ = RegisterProtocol(json.NewJsonCreator())
}

func RegisterProtocol(packetCreator protocol.Creator) error {
	// register protocol
	_, ok := protocols[packetCreator.PacketType()]
	if ok {
		return fmt.Errorf("register protocol, but already registered")
	}
	creator, ok := packetCreator.(protocol.Creator)
	if !ok {
		return fmt.Errorf("register protocol, but not implement protocol.Creator")
	}
	protocols[packetCreator.PacketType()] = creator
	return nil
}

func CreateProtocol(packetType protocol.PacketType, defaultType ...protocol.PacketType) (protocol.Packet, error) {
	creator, _ := GetCreator(packetType, defaultType...)
	if creator == nil {
		return nil, fmt.Errorf("create protocol failed, not registered")
	}
	return creator.NewProtocol(), nil
}

func GetCreator(packetType protocol.PacketType, defaultType ...protocol.PacketType) (retCreator protocol.Creator, useType protocol.PacketType) {
	creator, ok := protocols[packetType]
	if ok {
		// no need to find defaultTypes
		defaultType = nil
		useType = packetType
	}
	for _, dfType := range defaultType {
		creator, ok = protocols[dfType]
		if ok {
			useType = dfType
			break
		}
	}
	return creator, useType
}

func SupportProtocol(packetType protocol.PacketType) bool {
	_, ok := protocols[packetType]
	return ok
}
