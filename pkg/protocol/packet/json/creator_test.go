package json

import (
	"testing"

	"github.com/liuscraft/spider-network/pkg/protocol"
)

func TestNewJsonCreator(t *testing.T) {
	creator := NewJsonCreator()
	
	if creator == nil {
		t.Error("NewJsonCreator() 不应该返回 nil")
	}
	
	if creator.PacketType() != protocol.JsonType {
		t.Errorf("PacketType() = %v, 期望 %v", creator.PacketType(), protocol.JsonType)
	}
}

func TestJsonCreator_NewProtocol(t *testing.T) {
	creator := NewJsonCreator()
	
	tests := []struct {
		name    string
		options []protocol.Option
	}{
		{
			name:    "无选项创建",
			options: nil,
		},
		{
			name: "带选项创建",
			options: []protocol.Option{
				func(p protocol.Packet) {
					// 空选项,仅测试是否正确传递
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			packet := creator.NewProtocol(tt.options...)
			
			if packet == nil {
				t.Error("NewProtocol() 不应该返回 nil")
				return
			}
			
			if packet.PacketType() != protocol.JsonType {
				t.Errorf("创建的协议类型错误, got = %v, want = %v", 
					packet.PacketType(), protocol.JsonType)
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

func TestJsonCreator_GzipAndUnzip(t *testing.T) {
	creator := NewJsonCreator()
	packet := creator.NewProtocol()

	if creator.Gzip(packet) {
		t.Error("JSON 协议不应该支持 Gzip")
	}

	if creator.Unzip(packet) {
		t.Error("JSON 协议不应该支持 Unzip")
	}
} 