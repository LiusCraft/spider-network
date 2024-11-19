package json

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/liuscraft/spider-network/pkg/protocol"
)

type testStruct struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func TestJsonProtocol_WriteAndRead(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		want    interface{}
		wantErr bool
	}{
		{
			name:    "写入结构体",
			input:   testStruct{Name: "test", Age: 20},
			want:    &testStruct{Name: "test", Age: 20},
			wantErr: false,
		},
		{
			name:    "写入字节切片",
			input:   []byte(`{"name":"test","age":20}`),
			want:    &testStruct{Name: "test", Age: 20},
			wantErr: false,
		},
		{
			name:    "写入无效JSON",
			input:   []byte("invalid json"),
			want:    &testStruct{},
			wantErr: true,
		},
		{
			name:    "写入map",
			input:   map[string]interface{}{"name": "test", "age": 20},
			want:    &testStruct{Name: "test", Age: 20},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &JsonProtocol{}

			// 测试写入
			n, err := p.Write(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Write() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// 验证写入的数据大小
				if n != p.PacketSize() {
					t.Errorf("Write() returned size = %v, actual size = %v", n, p.PacketSize())
				}

				// 测试读取
				result := &testStruct{}
				readN, err := p.Read(result)
				if err != nil {
					t.Errorf("Read() unexpected error = %v", err)
					return
				}

				if readN != p.PacketSize() {
					t.Errorf("Read() returned size = %v, want = %v", readN, p.PacketSize())
				}

				if !reflect.DeepEqual(result, tt.want) {
					t.Errorf("Read() result = %v, want %v", result, tt.want)
				}
			}
		})
	}
}

func TestJsonProtocol_ToPacket(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		wantErr bool
	}{
		{
			name:    "有效JSON数据",
			data:    []byte(`{"name":"test","age":20}`),
			wantErr: false,
		},
		{
			name:    "空数据",
			data:    []byte{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &JsonProtocol{}
			n, err := p.ToPacket(tt.data)

			if (err != nil) != tt.wantErr {
				t.Errorf("ToPacket() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if n != len(tt.data) {
					t.Errorf("ToPacket() returned size = %v, want %v", n, len(tt.data))
				}

				if !reflect.DeepEqual(p.Bytes(), tt.data) {
					t.Errorf("ToPacket() data = %v, want %v", p.Bytes(), tt.data)
				}
			}
		})
	}
}

func TestJsonProtocol_PacketType(t *testing.T) {
	p := &JsonProtocol{}
	if p.PacketType() != protocol.JsonType {
		t.Errorf("PacketType() = %v, want %v", p.PacketType(), protocol.JsonType)
	}
}

func TestJsonProtocol_Clear(t *testing.T) {
	p := &JsonProtocol{}

	// 写入一些数据
	testData := testStruct{Name: "test", Age: 20}
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

func TestJsonProtocol_Bytes(t *testing.T) {
	p := &JsonProtocol{}
	testData := testStruct{Name: "test", Age: 20}

	p.Write(testData)

	expectedJSON, _ := json.Marshal(testData)
	if !reflect.DeepEqual(p.Bytes(), expectedJSON) {
		t.Errorf("Bytes() = %v, want %v", p.Bytes(), expectedJSON)
	}
}

func TestJsonProtocol_PacketSize(t *testing.T) {
	p := &JsonProtocol{}
	testData := testStruct{Name: "test", Age: 20}

	p.Write(testData)

	expectedJSON, _ := json.Marshal(testData)
	if p.PacketSize() != len(expectedJSON) {
		t.Errorf("PacketSize() = %v, want %v", p.PacketSize(), len(expectedJSON))
	}
}
