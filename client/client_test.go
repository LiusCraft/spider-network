package client

import (
	"encoding/json"
	"net"
	"testing"
	"time"

	"github.com/liuscraft/spider-network/pkg/protocol"
	"github.com/liuscraft/spider-network/pkg/protocol/hole"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// 模拟服务器
type mockServer struct {
	listener net.Listener
	clients  map[string]net.Conn
}

func newMockServer(t *testing.T) *mockServer {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	server := &mockServer{
		listener: listener,
		clients:  make(map[string]net.Conn),
	}

	// 启动服务器
	go server.start(t)
	return server
}

func (s *mockServer) start(t *testing.T) {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			return
		}

		go s.handleClient(t, conn)
	}
}

func (s *mockServer) handleClient(t *testing.T, conn net.Conn) {
	defer conn.Close()

	// 等待注册消息
	packet, err := protocol.NewPacketIO(conn, nil).ReadPacket()
	require.NoError(t, err)

	var msg hole.Message
	_, err = packet.Read(&msg)
	require.NoError(t, err)
	require.Equal(t, hole.TypeRegister, msg.Type)

	var payload struct {
		ClientID string `json:"client_id"`
		Name     string `json:"name"`
	}
	err = json.Unmarshal(msg.Payload, &payload)
	require.NoError(t, err)

	// 保存客户端连接
	s.clients[payload.ClientID] = conn

	// 发送注册确认
	response := &hole.Message{
		Type: hole.TypeRegister,
		From: "server",
		To:   payload.ClientID,
	}
	packet, err = hole.CreateHolePacket(response)
	require.NoError(t, err)

	err = protocol.NewPacketIO(nil, conn).WritePacket(packet)
	require.NoError(t, err)

	// 处理后续消息
	for {
		packet, err := protocol.NewPacketIO(conn, nil).ReadPacket()
		if err != nil {
			return
		}

		_, err = packet.Read(&msg)
		require.NoError(t, err)

		switch msg.Type {
		case hole.TypePunch:
			// 转发打洞消息给目标客户端
			targetConn := s.clients[msg.To]
			if targetConn == nil {
				continue
			}

			readyMsg := &hole.Message{
				Type:    hole.TypePunchReady,
				From:    msg.From,
				To:      msg.To,
				Payload: msg.Payload,
			}
			packet, _ = hole.CreateHolePacket(readyMsg)
			err = protocol.NewPacketIO(nil, targetConn).WritePacket(packet)
			require.NoError(t, err)

		case hole.TypeConnect:
			// 转发连接确认消息
			targetConn := s.clients[msg.To]
			if targetConn == nil {
				continue
			}
			packet, _ = hole.CreateHolePacket(&msg)
			err = protocol.NewPacketIO(nil, targetConn).WritePacket(packet)
			require.NoError(t, err)
		}
	}
}

func (s *mockServer) close() {
	s.listener.Close()
	for _, conn := range s.clients {
		conn.Close()
	}
}

func TestClientRegistration(t *testing.T) {
	server := newMockServer(t)
	defer server.close()

	client := NewClient("test-1", "Test Client 1")
	err := client.Connect(server.listener.Addr().String())
	require.NoError(t, err)
	defer client.Close()

	// 等待注册完成
	time.Sleep(100 * time.Millisecond)

	// 验证客户端已注册
	_, exists := server.clients["test-1"]
	assert.True(t, exists)
}

func TestPeerConnection(t *testing.T) {
	server := newMockServer(t)
	defer server.close()

	// 创建两个客户端
	client1 := NewClient("test-1", "Test Client 1")
	err := client1.Connect(server.listener.Addr().String())
	require.NoError(t, err)
	defer client1.Close()

	client2 := NewClient("test-2", "Test Client 2")
	err = client2.Connect(server.listener.Addr().String())
	require.NoError(t, err)
	defer client2.Close()

	// 等待注册完成
	time.Sleep(100 * time.Millisecond)

	// 客户端1连接客户端2
	err = client1.ConnectToPeer("test-2")
	require.NoError(t, err)

	// 等待连接建立
	time.Sleep(500 * time.Millisecond)

	// 验证连接是否建立
	var conn1, conn2 interface{}
	var ok bool

	conn1, ok = client1.peers.Load("test-2")
	assert.True(t, ok)
	assert.NotNil(t, conn1)

	conn2, ok = client2.peers.Load("test-1")
	assert.True(t, ok)
	assert.NotNil(t, conn2)
}

func TestMessageExchange(t *testing.T) {
	server := newMockServer(t)
	defer server.close()

	// 创建两个客户端
	client1 := NewClient("test-1", "Test Client 1")
	err := client1.Connect(server.listener.Addr().String())
	require.NoError(t, err)
	defer client1.Close()

	client2 := NewClient("test-2", "Test Client 2")
	err = client2.Connect(server.listener.Addr().String())
	require.NoError(t, err)
	defer client2.Close()

	// 等待注册完成
	time.Sleep(100 * time.Millisecond)

	// 客户端1连接客户端2
	err = client1.ConnectToPeer("test-2")
	require.NoError(t, err)

	// 等待连接建立
	time.Sleep(500 * time.Millisecond)

	// 发送消息
	testMessage := "Hello, test message!"
	client1.handleSendCommand("test-2", testMessage)

	// 等待消息传递
	time.Sleep(100 * time.Millisecond)

	// TODO: 添加消息接收验证
	// 由于消息处理是异步的，这里需要一个机制来验证消息是否被正确接收
	// 可以通过添加消息回调或消息队列来实现
}
