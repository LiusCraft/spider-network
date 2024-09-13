package protocol

import (
	"encoding/binary"
	"fmt"
	"io"
)

const (
	// packet header tag: "uint8", length: 1 byte
	// packet total size length: 4 bytes
	packetHeadLen = 5
)

type PacketType = byte

var (
	protocols = map[PacketType]Creator{}
)

type Protocol interface {
	Read(p []byte) (n int, err error)
	Write(p []byte) (n int, err error)
	PacketSize() int
	// PacketType default is BytesPacket, also zero
	PacketType() PacketType
	Writer(writer io.Writer) (n int, err error)
}

type Creator interface {
	NewProtocol(packetSize int) Protocol
}

func init() {
	_ = RegisterProtocol(&bytesProtocol{})
}

func RegisterProtocol(protocol Protocol) error {
	// register protocol
	_, ok := protocols[protocol.PacketType()]
	if ok {
		return fmt.Errorf("register protocol, but already registered")
	}
	creator, ok := protocol.(Creator)
	if !ok {
		return fmt.Errorf("register protocol, but not implement protocol.Creator")
	}
	protocols[protocol.PacketType()] = creator
	return nil
}

func CreateProtocol(packetType PacketType, size int) (Protocol, error) {
	creator, ok := protocols[packetType]
	if !ok {
		return nil, fmt.Errorf("create protocol failed, not registered")
	}
	return creator.NewProtocol(size), nil
}

func ReceivePacket(reader io.Reader) (Protocol, error) {
	// read packet header
	header := make([]byte, packetHeadLen)
	if _, err := reader.Read(header); err != nil {
		if err == io.EOF {
			return nil, err
		}
		return nil, fmt.Errorf("read packet header failed, err: %v", err)
	}
	tag := header[1]
	// create protocol
	creator, ok := protocols[tag]
	if !ok {
		return nil, fmt.Errorf("create protocol failed")
	}
	// read packet size
	size := binary.BigEndian.Uint32(header[1:])
	protocol := creator.NewProtocol(int(size))
	// read packet body
	body := make([]byte, size)
	if _, err := reader.Read(body); err != nil {
		return nil, fmt.Errorf("read packet body failed, err: %v", err)
	}
	// write packet body to protocol
	if _, err := protocol.Write(body); err != nil {
		return nil, fmt.Errorf("write protocol packet failed, err: %v", err)
	}
	return protocol, nil
}
