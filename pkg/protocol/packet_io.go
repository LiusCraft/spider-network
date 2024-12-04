package protocol

import (
	"fmt"
	"io"
	"sync"
)

var (
	creators = make(map[PacketType]Creator)
	mu       sync.RWMutex
)

// RegisterCreator 注册协议创建器
func RegisterCreator(c Creator) {
	mu.Lock()
	defer mu.Unlock()
	creators[c.PacketType()] = c
}

// GetCreator 获取协议创建器
func GetCreator(typ PacketType) (Creator, bool) {
	mu.RLock()
	defer mu.RUnlock()
	c, ok := creators[typ]
	return c, ok
}

// ReadPacket 从Reader中读取一个完整的数据包
func ReadPacket(r io.Reader) (Packet, error) {
	// 读取包头
	headerBuf := make([]byte, HeaderSize)
	if _, err := io.ReadFull(r, headerBuf); err != nil {
		return nil, err
	}

	header, err := DecodeHeader(headerBuf)
	if err != nil {
		return nil, err
	}

	// 获取协议创建器
	creator, ok := GetCreator(header.Type)
	if !ok {
		return nil, ErrInvalidPacket
	}

	// 读取数据
	dataBuf := make([]byte, header.Length)
	if _, err := io.ReadFull(r, dataBuf); err != nil {
		return nil, err
	}

	// 创建数据包并写入数据
	packet := creator.NewPacket()
	_, err = packet.Write(dataBuf)
	if err != nil {
		return nil, err
	}

	return packet, nil
}

// WritePacket 将数据包写入Writer
func WritePacket(w io.Writer, packet Packet) (n int, err error) {
	data := packet.Bytes()
	return w.Write(data)
}

// ReceivePacket 从Reader中读取一个完整的数据包，并处理错误
func ReceivePacket(r io.Reader) (Packet, error) {
	packet, err := ReadPacket(r)
	if err != nil {
		if err == io.EOF {
			return nil, err
		}
		// 其他错误处理
		return nil, fmt.Errorf("read packet error: %v", err)
	}
	return packet, nil
}

func init() {
	// 注册默认的协议创建器
	RegisterCreator(&BytesCreator{})
	RegisterCreator(&JSONCreator{})
}

// PacketIO 包装了读写操作
type PacketIO struct {
	reader io.Reader
	writer io.Writer
}

// NewPacketIO 创建新的PacketIO
func NewPacketIO(r io.Reader, w io.Writer) *PacketIO {
	return &PacketIO{
		reader: r,
		writer: w,
	}
}

// ReadPacket 读取数据包
func (p *PacketIO) ReadPacket() (Packet, error) {
	return ReceivePacket(p.reader)
}

// WritePacket 写入数据包
func (p *PacketIO) WritePacket(packet Packet) error {
	if err := ValidatePacket(packet); err != nil {
		return err
	}
	_, err := WritePacket(p.writer, packet)
	return err
}
