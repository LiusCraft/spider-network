package json

import (
	"encoding/json"

	"github.com/liuscraft/spider-network/pkg/protocol"
)

type JsonProtocol struct {
	packetSize int    // packet real size
	body       []byte // packet data
}

func (j *JsonProtocol) Read(p interface{}) (n int, err error) {
	return j.PacketSize(), json.Unmarshal(j.body, p)
}

func (j *JsonProtocol) Write(p interface{}) (n int, err error) {
	// check, if p is []byte, then just copy it
	bytes, ok := p.([]byte)
	if !ok {
		// if p is not []byte, then marshal it
		bytes, err = json.Marshal(p)
		if err != nil {
			return 0, err
		}
	} else {
		// 验证是否为有效的JSON格式
		var js interface{}
		if err := json.Unmarshal(bytes, &js); err != nil {
			return 0, err
		}
	}
	j.packetSize = len(bytes)
	j.body = bytes
	return j.packetSize, nil
}

func (j *JsonProtocol) ToPacket(p []byte) (n int, err error) {
	j.packetSize = len(p)
	j.body = p
	return j.PacketSize(), nil
}

func (j *JsonProtocol) Bytes() []byte {
	return j.body
}

func (j *JsonProtocol) PacketSize() int {
	return j.packetSize
}

func (j *JsonProtocol) PacketType() protocol.PacketType {
	return protocol.JsonType
}

func (j *JsonProtocol) Clear() {
	j.packetSize = 0
	j.body = make([]byte, 0)
}
