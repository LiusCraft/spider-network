package protocol

import "encoding/json"

// JSONPacket JSON协议实现
type JSONPacket struct {
	data []byte
}

func (p *JSONPacket) Read(v interface{}) (n int, err error) {
	// 如果目标是 []byte，直接复制数据
	if b, ok := v.(*[]byte); ok {
		*b = append(*b, p.data...)
		return len(p.data), nil
	}
	// 否则尝试 JSON 解析
	err = json.Unmarshal(p.data, v)
	if err != nil {
		return 0, err
	}
	return len(p.data), nil
}

func (p *JSONPacket) Write(v interface{}) (n int, err error) {
	switch data := v.(type) {
	case []byte:
		// 如果是 []byte，直接存储
		p.data = append(p.data[:0], data...)
		return len(data), nil
	case string:
		// 如果是字符串，转换为 []byte
		p.data = []byte(data)
		return len(p.data), nil
	}
	data, err := json.Marshal(v)
	if err != nil {
		return 0, err
	}
	p.data = data
	return len(data), nil
}

func (p *JSONPacket) Bytes() []byte {
	header := EncodeHeader(JsonType, uint32(len(p.data)))
	return append(header, p.data...)
}

func (p *JSONPacket) PacketSize() int {
	return HeaderSize + len(p.data)
}

func (p *JSONPacket) PacketType() PacketType {
	return JsonType
}

func (p *JSONPacket) Clear() {
	p.data = p.data[:0]
}

// JSONCreator JSON协议创建器
type JSONCreator struct{}

func (c *JSONCreator) NewPacket() Packet {
	return &JSONPacket{
		data: make([]byte, 0),
	}
}

func (c *JSONCreator) PacketType() PacketType {
	return JsonType
}

// NewJSONPacket 创建一个新的JSON数据包
func NewJSONPacket() *JSONPacket {
	return &JSONPacket{
		data: make([]byte, 0),
	}
}
