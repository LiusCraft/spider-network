package packet

import (
	"testing"

	"github.com/liuscraft/spider-network/pkg/protocol"
	"github.com/liuscraft/spider-network/pkg/protocol/packet/bytes"
)

// 模拟一个新的协议创建器用于测试
type mockCreator struct {
	packetType protocol.PacketType
}

func (m *mockCreator) PacketType() protocol.PacketType {
	return m.packetType
}

func (m *mockCreator) NewProtocol(options ...protocol.Option) protocol.Packet {
	return bytes.NewBytesCreator().NewProtocol()
}

func (m *mockCreator) Gzip(packet protocol.Packet) bool {
	return false
}

func (m *mockCreator) Unzip(packet protocol.Packet) bool {
	return false
}

func TestRegisterProtocol(t *testing.T) {
	// 测试注册已存在的协议
	err := RegisterProtocol(bytes.NewBytesCreator())
	if err == nil {
		t.Error("期望重复注册返回错误，但获得 nil")
	}

	// 测试注册新协议
	mockType := protocol.PacketType(3) // 使用一个未使用的协议类型
	mock := &mockCreator{packetType: mockType}
	err = RegisterProtocol(mock)
	if err != nil {
		t.Errorf("注册新协议失败: %v", err)
	}

	// 验证是否已注册成功
	if !SupportProtocol(mockType) {
		t.Error("协议注册后应该被支持")
	}
}

func TestCreateProtocol(t *testing.T) {
	tests := []struct {
		name        string
		packetType  protocol.PacketType
		defaultType []protocol.PacketType
		wantErr     bool
	}{
		{
			name:       "创建JSON协议",
			packetType: protocol.JsonType,
			wantErr:    false,
		},
		{
			name:       "创建Bytes协议",
			packetType: protocol.BytesType,
			wantErr:    false,
		},
		{
			name:       "创建未注册的协议",
			packetType: protocol.PacketType(5),
			wantErr:    true,
		},
		{
			name:        "使用默认类型创建",
			packetType:  protocol.PacketType(5),
			defaultType: []protocol.PacketType{protocol.JsonType},
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			packet, err := CreateProtocol(tt.packetType, tt.defaultType...)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateProtocol() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && packet == nil {
				t.Error("期望创建成功的协议不应该为 nil")
			}
		})
	}
}

func TestGetCreator(t *testing.T) {
	// 测试获取已注册的创建器
	creator, useType := GetCreator(protocol.JsonType)
	if creator == nil {
		t.Error("应该能获取到 JSON 协议创建器")
	}
	if useType != protocol.JsonType {
		t.Error("返回的协议类型不正确")
	}

	// 测试使用默认类型
	creator, useType = GetCreator(protocol.PacketType(5), protocol.JsonType)
	if creator == nil {
		t.Error("使用默认类型时应该能获取到创建器")
	}
	if useType != protocol.JsonType {
		t.Error("使用默认类型时返回的协议类型不正确")
	}

	// 测试获取未注册的创建器
	creator, _ = GetCreator(protocol.PacketType(5))
	if creator != nil {
		t.Error("获取未注册的协议创建器应该返回 nil")
	}
}

func TestSupportProtocol(t *testing.T) {
	// 测试支持的协议
	if !SupportProtocol(protocol.JsonType) {
		t.Error("JSON 协议应该被支持")
	}
	if !SupportProtocol(protocol.BytesType) {
		t.Error("Bytes 协议应该被支持")
	}

	// 测试不支持的协议
	if SupportProtocol(protocol.PacketType(5)) {
		t.Error("未注册的协议不应该被支持")
	}
}
