package bytes

import (
	"testing"

	"github.com/liuscraft/spider-network/pkg/errors"
	"github.com/liuscraft/spider-network/pkg/protocol"
)

func TestNewBytesProtocol(t *testing.T) {
	tests := []struct {
		name     string
		dataType protocol.PacketType
	}{
		{
			name:     "默认类型",
			dataType: protocol.BytesType,
		},
		{
			name:     "自定义类型",
			dataType: protocol.PacketType(3),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newBytesProtocol(tt.dataType)

			if p.PacketType() != tt.dataType {
				t.Errorf("PacketType() = %v, want %v", p.PacketType(), tt.dataType)
			}

			if p.PacketSize() != 0 {
				t.Errorf("新创建的协议大小应该为 0, got %v", p.PacketSize())
			}

			if len(p.Bytes()) != 0 {
				t.Error("新创建的协议不应该包含数据")
			}
		})
	}
}

func TestBytesProtocol_WriteAndBytes(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		wantErr bool
		want    []byte
	}{
		{
			name:    "写入字节切片",
			input:   []byte("test data"),
			wantErr: false,
			want:    []byte("test data"),
		},
		{
			name:    "写入非字节切片",
			input:   "test data",
			wantErr: true,
			want:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newBytesProtocol(protocol.BytesType)
			n, err := p.Write(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("Write() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if n != len(tt.want) {
					t.Errorf("Write() returned length = %v, want %v", n, len(tt.want))
				}

				if string(p.Bytes()) != string(tt.want) {
					t.Errorf("Bytes() = %v, want %v", string(p.Bytes()), string(tt.want))
				}

				if p.PacketSize() != len(tt.want) {
					t.Errorf("PacketSize() = %v, want %v", p.PacketSize(), len(tt.want))
				}
			}
		})
	}
}

func TestBytesProtocol_Read(t *testing.T) {
	// 创建一个模拟的目标协议
	mockPacket := newBytesProtocol(protocol.BytesType)

	tests := []struct {
		name    string
		data    []byte
		target  interface{}
		wantErr bool
	}{
		{
			name:    "读取到协议接口",
			data:    []byte("test data"),
			target:  mockPacket,
			wantErr: false,
		},
		{
			name:    "读取到非协议接口",
			data:    []byte("test data"),
			target:  "invalid target",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newBytesProtocol(protocol.BytesType)
			p.Write(tt.data)

			n, err := p.Read(tt.target)

			if (err != nil) != tt.wantErr {
				t.Errorf("Read() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if n != len(tt.data) {
					t.Errorf("Read() returned length = %v, want %v", n, len(tt.data))
				}

				targetPacket, ok := tt.target.(protocol.Packet)
				if !ok {
					t.Error("目标类型断言失败")
					return
				}

				if string(targetPacket.Bytes()) != string(tt.data) {
					t.Errorf("目标数据不匹配, got = %v, want %v", 
						string(targetPacket.Bytes()), string(tt.data))
				}
			} else {
				if err != errors.ErrToPacketTypeNotImplemented {
					t.Errorf("期望错误类型 = %v, got = %v", 
						errors.ErrToPacketTypeNotImplemented, err)
				}
			}
		})
	}
}

func TestBytesProtocol_ToPacket(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{
			name: "空数据",
			data: []byte{},
		},
		{
			name: "普通数据",
			data: []byte("test data"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newBytesProtocol(protocol.BytesType)
			n, err := p.ToPacket(tt.data)

			if err != nil {
				t.Errorf("ToPacket() unexpected error = %v", err)
				return
			}

			if n != len(tt.data) {
				t.Errorf("ToPacket() returned length = %v, want %v", n, len(tt.data))
			}

			if string(p.Bytes()) != string(tt.data) {
				t.Errorf("数据不匹配, got = %v, want %v", 
					string(p.Bytes()), string(tt.data))
			}

			if p.PacketSize() != len(tt.data) {
				t.Errorf("PacketSize() = %v, want %v", p.PacketSize(), len(tt.data))
			}
		})
	}
}

func TestBytesProtocol_Clear(t *testing.T) {
	p := newBytesProtocol(protocol.BytesType)
	
	// 写入一些数据
	testData := []byte("test data")
	p.Write(testData)

	// 清除数据
	p.Clear()

	if p.PacketSize() != 0 {
		t.Errorf("Clear() 后大小应该为 0, got %v", p.PacketSize())
	}

	if len(p.Bytes()) != 0 {
		t.Error("Clear() 后不应该包含数据")
	}
}

func TestBytesProtocol_String(t *testing.T) {
	p := newBytesProtocol(protocol.BytesType)
	testData := []byte("test data")
	p.Write(testData)

	if p.String() != string(testData) {
		t.Errorf("String() = %v, want %v", p.String(), string(testData))
	}
} 