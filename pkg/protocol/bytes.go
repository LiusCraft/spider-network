package protocol

// BytesPacket 字节协议实现
type BytesPacket struct {
    data []byte
}

func (p *BytesPacket) Read(v interface{}) (n int, err error) {
    if b, ok := v.(*[]byte); ok {
        *b = append(*b, p.data...)
        return len(p.data), nil
    }
    return 0, ErrInvalidPacket
}

func (p *BytesPacket) Write(v interface{}) (n int, err error) {
    if b, ok := v.([]byte); ok {
        p.data = append(p.data, b...)
        return len(b), nil
    }
    return 0, ErrInvalidPacket
}

func (p *BytesPacket) Bytes() []byte {
    header := EncodeHeader(BytesType, uint32(len(p.data)))
    return append(header, p.data...)
}

func (p *BytesPacket) PacketSize() int {
    return HeaderSize + len(p.data)
}

func (p *BytesPacket) PacketType() PacketType {
    return BytesType
}

func (p *BytesPacket) Clear() {
    p.data = p.data[:0]
}

// BytesCreator 字节协议创建器
type BytesCreator struct{}

func (c *BytesCreator) NewPacket() Packet {
    return &BytesPacket{
        data: make([]byte, 0),
    }
}

func (c *BytesCreator) PacketType() PacketType {
    return BytesType
}

// NewBytesPacket 创建一个新的字节数据包
func NewBytesPacket() *BytesPacket {
    return &BytesPacket{
        data: make([]byte, 0),
    }
} 