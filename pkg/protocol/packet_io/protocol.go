package packet_io

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/liuscraft/spider-network/pkg/protocol"
	"github.com/liuscraft/spider-network/pkg/protocol/packet"
	"github.com/liuscraft/spider-network/pkg/protocol/packet/bytes"
	"github.com/liuscraft/spider-network/pkg/xlog"
)

const (
	// packet header tag: "uint8", length: 1 byte
	// packet total size length: 4 bytes
	packetHeadLen = 5
)

func ReceivePacket(reader io.Reader) (protocol.Packet, error) {
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
	creator, useType := packet.GetCreator(tag, protocol.DefaultPacketType)
	if creator == nil {
		return nil, fmt.Errorf("create protocol failed, not registered")
	}
	var creatorWithOptions []protocol.Option
	if useType == protocol.DefaultPacketType && tag != protocol.DefaultPacketType {
		creatorWithOptions = append(creatorWithOptions, bytes.WithType(tag))
	}
	protocolPacket := creator.NewProtocol(creatorWithOptions...)

	// read packet size
	size := binary.BigEndian.Uint32(header[1:])
	// read packet body
	body := make([]byte, size)
	if _, err := reader.Read(body); err != nil {
		return nil, fmt.Errorf("read packet body failed, err: %v", err)
	}
	// write packet body to protocol
	if _, err := protocolPacket.Write(body); err != nil {
		return nil, fmt.Errorf("write protocol packet failed, err: %v", err)
	}
	return protocolPacket, nil
}

func WritePacket(writer io.Writer, wPacket protocol.Packet, autoClear ...bool) (n int, err error) {
	sendPacket := wPacket
	support := packet.SupportProtocol(wPacket.PacketType())
	if !support {
		xlog.Warn("no support for packet type, use the bytes default protocol")
		creator, _ := packet.GetCreator(protocol.DefaultPacketType)
		sendPacket = creator.NewProtocol(bytes.WithType(wPacket.PacketType()))
		_, err = sendPacket.ToPacket(wPacket.Bytes())
		if err != nil {
			return 0, err
		}
	}
	n, err = writer.Write(CombingPacket(sendPacket))
	if err != nil {
		return 0, err
	}
	if len(autoClear) > 0 && autoClear[0] {
		wPacket.Clear()
	}
	return
}

func Int32ToBytes(n uint32) []byte {
	buf := make([]byte, 4) // int32 占 4 个字节
	binary.BigEndian.PutUint32(buf, n)
	return buf
}

func CombingPacket(b protocol.Packet) []byte {
	out := make([]byte, 0, b.PacketSize()+5)
	out = append(out, b.PacketType())
	out = append(out, Int32ToBytes(uint32(b.PacketSize()))...)
	out = append(out, b.Bytes()...)
	return out
}
