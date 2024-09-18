package protocol

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/liuscraft/spider-network/pkg/xlog"
)

const (
	// packet header tag: "uint8", length: 1 byte
	// packet total size length: 4 bytes
	packetHeadLen = 5
)

type PacketType = byte

var (
	defaultPacketType = BytesType
	protocols         = map[PacketType]Creator{}
)

type Packet interface {
	/* Read write to target packet
	example:
		bytesPacket to stringPacket(target packet)
		src := &bytesPacket{}
	    src.Write([]byte("hello world"))
		targetPacket := &stringPacket{}
		src.Read(targetPacket)
		result := ""
		targetPacket.Read(result)
		fmt.Println(result) => "hello world"
	*/
	Read(p interface{}) (n int, err error)
	/* Write to the packet
	example:
		bytePacket := &bytePacket{}
	    bytePacket.Write([]byte("hello world")) // The internal implementation deals with data of type []byte
		stringPacket := &stringPacket{}
	    stringPacket.Write("hello world") // The internal implementation deals with data of type string
	*/
	Write(p interface{}) (n int, err error)
	// ToPacket Writes bytes to change the current packet content
	ToPacket(p []byte) (n int, err error)
	// Bytes get the packet to bytes
	Bytes() []byte
	// PacketSize get the packet size
	PacketSize() int
	// PacketType default is BytesPacket, also zero
	PacketType() PacketType
	Clear()
}

type Creator interface {
	NewProtocol() Packet
}

func init() {
	_ = RegisterProtocol(newBytesProtocol(BytesType))
}

func RegisterProtocol(protocol Packet) error {
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

func CreateProtocol(packetType PacketType) (Packet, error) {
	creator, ok := protocols[packetType]
	if !ok {
		return nil, fmt.Errorf("create protocol failed, not registered")
	}
	return creator.NewProtocol(), nil
}

func SupportProtocol(packetType PacketType) bool {
	_, ok := protocols[packetType]
	return ok
}

func ReceivePacket(reader io.Reader) (Packet, error) {
	// read packet header
	header := make([]byte, packetHeadLen)
	if _, err := reader.Read(header); err != nil {
		if err == io.EOF {
			return nil, err
		}
		return nil, fmt.Errorf("read packet header failed, err: %v", err)
	}
	tag := header[0]
	// create protocol
	creator, ok := protocols[tag]
	if !ok {
		creator = &bytesProtocol{dataType: tag, body: make([]byte, 0), packetSize: 0}
	}
	// read packet size
	size := binary.BigEndian.Uint32(header[1:])
	protocol := creator.NewProtocol()
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

func WritePacket(writer io.Writer, packet Packet, autoClear ...bool) (n int, err error) {
	xl := xlog.WithLogId(xlog.NewLogger(), "writePacket")
	sendPacket := packet
	support := SupportProtocol(packet.PacketType())
	if !support {
		sendPacket = &bytesProtocol{dataType: packet.PacketType(), body: packet.Bytes(), packetSize: packet.PacketSize()}
		xl.Debugf("useByteProtocol: %v", CombingPacket(sendPacket))
	}
	n, err = writer.Write(CombingPacket(sendPacket))
	if err != nil {
		return 0, err
	}
	if len(autoClear) > 0 && autoClear[0] {
		packet.Clear()
	}
	return
}

func Int32ToBytes(n uint32) []byte {
	buf := make([]byte, 4) // int32 占 4 个字节
	binary.BigEndian.PutUint32(buf, n)
	return buf
}

func CombingPacket(b Packet) []byte {
	out := make([]byte, 0, b.PacketSize()+5)
	out = append(out, b.PacketType())
	out = append(out, Int32ToBytes(uint32(b.PacketSize()))...)
	out = append(out, b.Bytes()...)
	return out
}
