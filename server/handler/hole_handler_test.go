package handler

import (
	"encoding/json"
	"net"
	"testing"
	"time"

	"github.com/liuscraft/spider-network/pkg/config"
	"github.com/liuscraft/spider-network/pkg/protocol/hole"
	"github.com/liuscraft/spider-network/pkg/protocol/packet_io"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// 模拟客户端
type mockClient struct {
	conn     net.Conn
	clientID string
	name     string
	messages chan *hole.HoleMessage
}

func newMockClient(t *testing.T, serverAddr string, clientID, name string) *mockClient {
	conn, err := net.Dial("tcp", serverAddr)
	require.NoError(t, err)
	
	client := &mockClient{
		conn:     conn,
		clientID: clientID,
		name:     name,
		messages: make(chan *hole.HoleMessage, 10),
	}
	
	// 启动消息接收
	go client.receiveMessages(t)
	
	return client
}

func (c *mockClient) receiveMessages(t *testing.T) {
	for {
		packet, err := packet_io.ReceivePacket(c.conn)
		if err != nil {
			close(c.messages)
			return
		}
		
		var msg hole.HoleMessage
		_, err = packet.Read(&msg)
		require.NoError(t, err)
		
		c.messages <- &msg
	}
}

func (c *mockClient) register(t *testing.T) {
	msg := &hole.HoleMessage{
		Type: hole.TypeRegister,
		From: c.clientID,
		Payload: []byte(`{
			"client_id": "` + c.clientID + `",
			"name": "` + c.name + `"
		}`),
	}
	
	packet, err := hole.CreateHolePacket(msg)
	require.NoError(t, err)
	
	_, err = packet_io.WritePacket(c.conn, packet)
	require.NoError(t, err)
}

func (c *mockClient) sendPunchRequest(t *testing.T, targetID string) {
	payload := hole.PunchPayload{
		PublicAddr:  c.conn.LocalAddr().String(),
		PrivateAddr: c.conn.LocalAddr().String(),
	}
	payloadBytes, err := json.Marshal(payload)
	require.NoError(t, err)
	
	msg := &hole.HoleMessage{
		Type:    hole.TypePunch,
		From:    c.clientID,
		To:      targetID,
		Payload: payloadBytes,
	}
	
	packet, err := hole.CreateHolePacket(msg)
	require.NoError(t, err)
	
	_, err = packet_io.WritePacket(c.conn, packet)
	require.NoError(t, err)
}

func (c *mockClient) close() {
	c.conn.Close()
}

func TestHoleHandlerRegistration(t *testing.T) {
	// 创建服务器
	cfg := config.HoleConfig{
		BindAddr: "127.0.0.1:0",
	}
	handler, err := NewHoleHandler(cfg)
	require.NoError(t, err)
	
	go handler.Start()
	defer handler.Stop()
	
	// 创建客户端
	client := newMockClient(t, handler.listener.Addr().String(), "test-1", "Test Client 1")
	defer client.close()
	
	// 注册客户端
	client.register(t)
	
	// 等待注册确认
	select {
	case msg := <-client.messages:
		assert.Equal(t, hole.TypeRegister, msg.Type)
		assert.Equal(t, "server", msg.From)
		assert.Equal(t, "test-1", msg.To)
	case <-time.After(time.Second):
		t.Fatal("Registration confirmation timeout")
	}
}

func TestHolePunching(t *testing.T) {
	// 创建服务器
	cfg := config.HoleConfig{
		BindAddr: "127.0.0.1:0",
	}
	handler, err := NewHoleHandler(cfg)
	require.NoError(t, err)
	
	go handler.Start()
	defer handler.Stop()
	
	// 创建两个客户端
	client1 := newMockClient(t, handler.listener.Addr().String(), "test-1", "Test Client 1")
	defer client1.close()
	
	client2 := newMockClient(t, handler.listener.Addr().String(), "test-2", "Test Client 2")
	defer client2.close()
	
	// 注册客户端
	client1.register(t)
	client2.register(t)
	
	// 等待注册确认
	for i := 0; i < 2; i++ {
		select {
		case <-client1.messages:
		case <-client2.messages:
		case <-time.After(time.Second):
			t.Fatal("Registration confirmation timeout")
		}
	}
	
	// 客户端1发送打洞请求
	client1.sendPunchRequest(t, "test-2")
	
	// 客户端2应该收到打洞准备消息
	select {
	case msg := <-client2.messages:
		assert.Equal(t, hole.TypePunchReady, msg.Type)
		assert.Equal(t, "test-1", msg.From)
		assert.Equal(t, "test-2", msg.To)
		
		var payload hole.PunchPayload
		err := json.Unmarshal(msg.Payload, &payload)
		require.NoError(t, err)
		assert.NotEmpty(t, payload.PublicAddr)
		assert.NotEmpty(t, payload.PrivateAddr)
	case <-time.After(time.Second):
		t.Fatal("Punch ready message timeout")
	}
}
