package protocol

import (
	"encoding/binary"
	"io"
)

const BytesType PacketType = 0

type bytesProtocol struct {
	packetSize int    // packet real size
	body       []byte // packet data
}

func newBytesProtocol(len int) *bytesProtocol {
	body := make([]byte, 5)
	body[0] = BytesType
	binary.BigEndian.PutUint32(body, uint32(len))
	return &bytesProtocol{
		packetSize: len,
		body:       body,
	}
}

func (b *bytesProtocol) Write(p []byte) (n int, err error) {
	if b.packetSize == 0 {
		return 0, io.EOF
	}
	b.body = append(b.body, p...)
	return b.packetSize, nil
}

func (b *bytesProtocol) Read(p []byte) (n int, err error) {
	p = append(p, b.body...)
	return b.packetSize, nil
}

func (b *bytesProtocol) PacketSize() int {
	return b.packetSize
}

func (b *bytesProtocol) PacketType() PacketType {
	return BytesType
}

func (b *bytesProtocol) Writer(w io.Writer) (n int, err error) {
	return w.Write(b.body)
}

func (b *bytesProtocol) NewProtocol(size int) Protocol {
	return newBytesProtocol(size)
}

func (b *bytesProtocol) String() string {
	return string(b.body)
}
