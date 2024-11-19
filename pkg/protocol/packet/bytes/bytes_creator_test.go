package bytes

import (
	"testing"

	"github.com/liuscraft/spider-network/pkg/protocol"
)

func TestNewBytesCreator(t *testing.T) {
	tests := []struct {
		name        string
		baseOptions []protocol.Option
	}{
		{
			name:        "无基础选项创建",
			baseOptions: nil,
		},
		{
			name: "带基础选项创建",
			baseOptions: []protocol.Option{
				WithType(protocol.PacketType(3)),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			creator := NewBytesCreator(tt.baseOptions...)

			if creator.PacketType() != protocol.BytesType {
				t.Errorf("PacketType() = %v, 期望 %v", creator.PacketType(), protocol.BytesType)
			}

			// 验证基础选项是否被正确保存
			if len(tt.baseOptions) > 0 {
				packet := creator.NewProtocol()
				if packet.PacketType() == protocol.BytesType {
					t.Error("基础选项应该修改了协议类型，但未生效")
				}
			}
		})
	}
}

func TestBytesCreator_NewProtocol(t *testing.T) {
	tests := []struct {
		name          string
		baseOptions   []protocol.Option
		extraOptions  []protocol.Option
		expectedType  protocol.PacketType
	}{
		{
			name:          "无选项创建",
			baseOptions:   nil,
			extraOptions:  nil,
			expectedType:  protocol.BytesType,
		},
		{
			name: "仅基础选项",
			baseOptions: []protocol.Option{
				WithType(protocol.PacketType(3)),
			},
			extraOptions:  nil,
			expectedType:  protocol.PacketType(3),
		},
		{
			name: "仅额外选项",
			baseOptions: nil,
			extraOptions: []protocol.Option{
				WithType(protocol.PacketType(4)),
			},
			expectedType: protocol.PacketType(4),
		},
		{
			name: "基础选项和额外选项",
			baseOptions: []protocol.Option{
				WithType(protocol.PacketType(3)),
			},
			extraOptions: []protocol.Option{
				WithType(protocol.PacketType(4)),
			},
			expectedType: protocol.PacketType(4), // 额外选项应该覆盖基础选项
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			creator := NewBytesCreator(tt.baseOptions...)
			packet := creator.NewProtocol(tt.extraOptions...)

			if packet == nil {
				t.Error("NewProtocol() 不应该返回 nil")
				return
			}

			if packet.PacketType() != tt.expectedType {
				t.Errorf("协议类型错误, got = %v, want = %v", 
					packet.PacketType(), tt.expectedType)
			}

			// 验证初始状态
			if packet.PacketSize() != 0 {
				t.Errorf("新创建的协议大小应该为 0, got = %v", packet.PacketSize())
			}

			if len(packet.Bytes()) != 0 {
				t.Error("新创建的协议不应该包含数据")
			}
		})
	}
}

func TestBytesCreator_GzipAndUnzip(t *testing.T) {
	creator := NewBytesCreator()
	packet := creator.NewProtocol()

	if creator.Gzip(packet) {
		t.Error("Bytes 协议不应该支持 Gzip")
	}

	if creator.Unzip(packet) {
		t.Error("Bytes 协议不应该支持 Unzip")
	}
} 