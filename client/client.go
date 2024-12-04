package client

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/liuscraft/spider-network/pkg/protocol"
	"github.com/liuscraft/spider-network/pkg/protocol/hole"
	"github.com/liuscraft/spider-network/pkg/xlog"
)

type Client struct {
	clientID   string
	name       string
	serverConn net.Conn
	peers      sync.Map
	xl         xlog.Logger
	listener   net.Listener
	// 添加统计信息
	stats struct {
		bytesSent    int64
		bytesRecv    int64
		p2pBytesSent int64
		p2pBytesRecv int64
		startTime    time.Time
		mutex        sync.Mutex
	}
	heartbeatCtx    context.Context
	heartbeatCancel context.CancelFunc
}

func NewClient(clientID, name string) *Client {
	ctx, cancel := context.WithCancel(context.Background())
	c := &Client{
		clientID:        clientID,
		name:            name,
		xl:              xlog.New(),
		heartbeatCtx:    ctx,
		heartbeatCancel: cancel,
	}
	c.stats.startTime = time.Now()
	return c
}

// 添加统计方法
func (c *Client) addBytesSent(n int64) {
	c.stats.mutex.Lock()
	defer c.stats.mutex.Unlock()
	c.stats.bytesSent += n
}

func (c *Client) addBytesRecv(n int64) {
	c.stats.mutex.Lock()
	defer c.stats.mutex.Unlock()
	c.stats.bytesRecv += n
}

func (c *Client) addP2PBytesSent(n int64) {
	c.stats.mutex.Lock()
	defer c.stats.mutex.Unlock()
	c.stats.p2pBytesSent += n
}

func (c *Client) addP2PBytesRecv(n int64) {
	c.stats.mutex.Lock()
	defer c.stats.mutex.Unlock()
	c.stats.p2pBytesRecv += n
}

func (c *Client) getStats() (int64, int64, int64, int64) {
	c.stats.mutex.Lock()
	defer c.stats.mutex.Unlock()
	return c.stats.bytesSent, c.stats.bytesRecv, c.stats.p2pBytesSent, c.stats.p2pBytesRecv
}

// 添加心跳机制
func (c *Client) startHeartbeat() {
	// 取消之前的心跳（如果有）
	if c.heartbeatCancel != nil {
		c.heartbeatCancel()
	}

	// 创建新的心跳上下文
	c.heartbeatCtx, c.heartbeatCancel = context.WithCancel(context.Background())

	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-c.heartbeatCtx.Done():
				return
			case <-ticker.C:
				if c.serverConn == nil {
					continue
				}

				// 收集当前状态
				sent, recv, p2pSent, p2pRecv := c.getStats()
				peers := make([]string, 0)
				c.peers.Range(func(key, value interface{}) bool {
					if peerID, ok := key.(string); ok {
						peers = append(peers, peerID)
					}
					return true
				})

				// 构造心跳消息
				heartbeatData := hole.HeartbeatPayload{
					ClientID:     c.clientID,
					BytesSent:    sent,
					BytesRecv:    recv,
					P2PBytesSent: p2pSent,
					P2PBytesRecv: p2pRecv,
					Peers:        peers,
					Timestamp:    time.Now().UnixNano(),
				}

				// 序列化心跳数据
				payloadBytes, err := json.Marshal(heartbeatData)
				if err != nil {
					c.xl.Errorf("Failed to marshal heartbeat data: %v", err)
					continue
				}

				// 构造消息
				msg := &hole.Message{
					Type:    hole.TypeHeartbeat,
					From:    c.clientID,
					To:      "server",
					Payload: payloadBytes,
				}

				// 发送心跳
				data, err := json.Marshal(msg)
				if err != nil {
					c.xl.Errorf("Failed to marshal message: %v", err)
					continue
				}

				if _, err := c.serverConn.Write(append(data, '\n')); err != nil {
					c.xl.Errorf("Failed to send heartbeat: %v", err)
				} else {
					c.xl.Debugf("Sent heartbeat: server=%d/%d, p2p=%d/%d bytes, peers=%v",
						sent, recv, p2pSent, p2pRecv, peers)
				}
			}
		}
	}()
}

func (c *Client) Connect(serverAddr string) error {
	var backoff = time.Second
	maxBackoff := 30 * time.Second

	// 首先创建监听器
	listener, err := net.Listen("tcp", "127.0.0.1:0") // 使用0让系统分配端口
	if err != nil {
		return fmt.Errorf("failed to create listener: %v", err)
	}
	c.listener = listener
	c.xl.Infof("Listening for peer connections on %s", listener.Addr())

	// 启动监听协程
	go c.acceptPeerConnections()

	for {
		// 尝试连接服务器
		conn, err := net.Dial("tcp", serverAddr)
		if err != nil {
			c.xl.Errorf("Failed to connect to server: %v", err)
			time.Sleep(backoff)

			// 指数退避
			backoff *= 2
			if backoff > maxBackoff {
				backoff = maxBackoff
			}
			continue
		}

		// 重置退避时间
		backoff = time.Second
		c.serverConn = conn
		c.xl.Infof("Connected to server %s", serverAddr)

		// 发送注册消息
		registerPayload := hole.RegisterPayload{
			ClientID:    c.clientID,
			Name:        c.name,
			PublicAddr:  c.listener.Addr().String(), // 添加监听地址
			PrivateAddr: c.listener.Addr().String(), // 添加监听地址
		}

		payloadBytes, err := json.Marshal(registerPayload)
		if err != nil {
			c.xl.Errorf("Failed to marshal register payload: %v", err)
			conn.Close()
			continue
		}

		msg := &hole.Message{
			Type:    hole.TypeRegister,
			From:    c.clientID,
			To:      "server",
			Payload: payloadBytes,
		}

		data, err := json.Marshal(msg)
		if err != nil {
			c.xl.Errorf("Failed to marshal register message: %v", err)
			conn.Close()
			continue
		}

		if _, err := conn.Write(append(data, '\n')); err != nil {
			c.xl.Errorf("Failed to send register message: %v", err)
			conn.Close()
			continue
		}

		// 等待注册响应
		reader := bufio.NewReader(conn)
		response, err := reader.ReadBytes('\n')
		if err != nil {
			c.xl.Errorf("Failed to read register response: %v", err)
			conn.Close()
			continue
		}

		var respMsg hole.Message
		if err := json.Unmarshal(response, &respMsg); err != nil {
			c.xl.Errorf("Failed to unmarshal register response: %v", err)
			conn.Close()
			continue
		}

		if respMsg.Type != hole.TypeRegister {
			c.xl.Errorf("Unexpected response type: %s", respMsg.Type)
			conn.Close()
			continue
		}

		c.xl.Infof("Successfully registered with server")

		// 启动心跳
		c.startHeartbeat()

		// 启动消息处理循环
		go c.handleServerMessages()

		return nil
	}
}

func (c *Client) handleServerMessages() {
	reader := bufio.NewReader(c.serverConn)
	for {
		data, err := reader.ReadBytes('\n')
		if err != nil {
			if err != io.EOF {
				c.xl.Errorf("Failed to read from server: %v", err)
			}
			c.serverConn.Close()
			return
		}

		var msg hole.Message
		if err := json.Unmarshal(data, &msg); err != nil {
			c.xl.Errorf("Failed to unmarshal message: %v", err)
			continue
		}

		// 处理不同类型的消息
		switch msg.Type {
		case hole.TypePunch:
			c.handlePunchMessage(&msg)
		case hole.TypeConnect:
			c.handleConnectMessage(&msg)
		default:
			c.xl.Warnf("Unknown message type: %s", msg.Type)
		}
	}
}

func (c *Client) acceptPeerConnections() {
	for {
		conn, err := c.listener.Accept()
		if err != nil {
			if err != net.ErrClosed {
				c.xl.Errorf("Failed to accept connection: %v", err)
			}
			return
		}

		// 启动一个新的 goroutine 来处理连接
		go func(conn net.Conn) {
			// 读取第一条消息以获取对方的ID
			reader := bufio.NewReader(conn)
			data, err := reader.ReadBytes('\n')
			if err != nil {
				c.xl.Errorf("Failed to read initial message: %v", err)
				conn.Close()
				return
			}

			var msg hole.Message
			if err := json.Unmarshal(data, &msg); err != nil {
				c.xl.Errorf("Failed to unmarshal initial message: %v", err)
				conn.Close()
				return
			}

			peerID := msg.From
			c.xl.Infof("Accepted connection from peer %s", peerID)

			// 保存连接
			c.peers.Store(peerID, conn)

			// 启动消息处理
			c.startPeerMessageHandler(peerID, conn)
		}(conn)
	}
}

func (c *Client) ConnectToPeer(peerID string) error {
	c.xl.Infof("Initiating connection to peer %s...", peerID)

	// 构造打洞消息
	payload := hole.PunchPayload{
		PublicAddr:  c.listener.Addr().String(),
		PrivateAddr: c.listener.Addr().String(),
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal punch payload: %v", err)
	}

	msg := &hole.Message{
		Type:    hole.TypePunch,
		From:    c.clientID,
		To:      peerID,
		Payload: payloadBytes,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal punch message: %v", err)
	}

	if _, err := c.serverConn.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("failed to send punch message: %v", err)
	}

	c.xl.Infof("Punch message sent to %s", peerID)
	return nil
}

func (c *Client) handlePunchMessage(msg *hole.Message) {
	c.xl.Infof("Received punch message from %s", msg.From)

	// 解析打洞消息
	var payload hole.PunchPayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		c.xl.Errorf("Failed to unmarshal punch payload: %v", err)
		return
	}

	// 尝试连接对方
	addrs := []string{payload.PublicAddr}
	if payload.PrivateAddr != "" && payload.PrivateAddr != payload.PublicAddr {
		addrs = append(addrs, payload.PrivateAddr)
	}

	var conn net.Conn
	var err error
	for _, addr := range addrs {
		c.xl.Infof("Trying to connect to %s at %s", msg.From, addr)
		conn, err = net.DialTimeout("tcp", addr, 5*time.Second)
		if err == nil {
			break
		}
		c.xl.Warnf("Failed to connect to %s at %s: %v", msg.From, addr, err)
	}

	if err != nil {
		c.xl.Errorf("Failed to connect to peer %s: %v", msg.From, err)
		return
	}

	// 发送初始消息
	initMsg := &hole.Message{
		Type: hole.TypeMessage,
		From: c.clientID,
		To:   msg.From,
	}
	data, err := json.Marshal(initMsg)
	if err != nil {
		c.xl.Errorf("Failed to marshal initial message: %v", err)
		conn.Close()
		return
	}

	if _, err := conn.Write(append(data, '\n')); err != nil {
		c.xl.Errorf("Failed to send initial message: %v", err)
		conn.Close()
		return
	}

	// 保存连接
	c.peers.Store(msg.From, conn)
	c.xl.Infof("Successfully connected to peer %s", msg.From)

	// 发送连接确认消息给服务器
	connectMsg := &hole.Message{
		Type:    hole.TypeConnect,
		From:    c.clientID,
		To:      msg.From,
		Payload: msg.Payload, // 使用相同的地址信息
	}

	data, err = json.Marshal(connectMsg)
	if err != nil {
		c.xl.Errorf("Failed to marshal connect message: %v", err)
		return
	}

	if _, err := c.serverConn.Write(append(data, '\n')); err != nil {
		c.xl.Errorf("Failed to send connect message: %v", err)
		return
	}

	// 启动消息处理
	go c.startPeerMessageHandler(msg.From, conn)
}

func (c *Client) handleConnectMessage(msg *hole.Message) {
	c.xl.Infof("Received connect message from %s", msg.From)

	// 检查是否已经连接
	if _, exists := c.peers.Load(msg.From); exists {
		c.xl.Infof("Already connected to peer %s", msg.From)
		return
	}

	// 解析地址信息
	var payload hole.PunchPayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		c.xl.Errorf("Failed to unmarshal connect payload: %v", err)
		return
	}

	// 尝试主动连接对方
	addrs := []string{payload.PublicAddr}
	if payload.PrivateAddr != "" && payload.PrivateAddr != payload.PublicAddr {
		addrs = append(addrs, payload.PrivateAddr)
	}

	var conn net.Conn
	var err error
	for _, addr := range addrs {
		c.xl.Infof("Trying to connect to %s at %s", msg.From, addr)
		conn, err = net.DialTimeout("tcp", addr, 5*time.Second)
		if err == nil {
			break
		}
		c.xl.Warnf("Failed to connect to %s at %s: %v", msg.From, addr, err)
	}

	if err != nil {
		c.xl.Errorf("Failed to connect to peer %s: %v", msg.From, err)
		return
	}

	// 发送初始消息
	initMsg := &hole.Message{
		Type: hole.TypeMessage,
		From: c.clientID,
		To:   msg.From,
	}
	data, err := json.Marshal(initMsg)
	if err != nil {
		c.xl.Errorf("Failed to marshal initial message: %v", err)
		conn.Close()
		return
	}

	if _, err := conn.Write(append(data, '\n')); err != nil {
		c.xl.Errorf("Failed to send initial message: %v", err)
		conn.Close()
		return
	}

	// 保存连接并启动消息处理
	c.peers.Store(msg.From, conn)
	c.xl.Infof("Successfully connected to peer %s", msg.From)
	go c.startPeerMessageHandler(msg.From, conn)
}

func (c *Client) startPeerMessageHandler(peerID string, conn net.Conn) {
	defer func() {
		conn.Close()
		c.peers.Delete(peerID)
		c.xl.Infof("Connection with peer %s closed", peerID)
	}()

	reader := bufio.NewReader(conn)
	for {
		// 读取消息
		data, err := reader.ReadBytes('\n')
		if err != nil {
			if err != io.EOF {
				c.xl.Errorf("Failed to read from peer %s: %v", peerID, err)
			}
			return
		}

		// 更新接收字节数
		c.addP2PBytesRecv(int64(len(data)))

		// 解析消息
		var msg hole.Message
		if err := json.Unmarshal(data, &msg); err != nil {
			c.xl.Errorf("Failed to unmarshal message from peer %s: %v", peerID, err)
			continue
		}

		// 验证消息来源
		if msg.From != peerID {
			c.xl.Warnf("Message from %s claims to be from %s", peerID, msg.From)
			continue
		}

		// 处理不同类型的消息
		switch msg.Type {
		case hole.TypeMessage:
			if len(msg.Payload) > 0 {
				// 尝试作为JSON解析，如果失败则作为普通文本处理
				var jsonContent string
				if err := json.Unmarshal(msg.Payload, &jsonContent); err != nil {
					// 作为普通文本处理
					c.xl.Infof("Message from %s: %s", peerID, string(msg.Payload))
				} else {
					// 作为JSON内容处理
					c.xl.Infof("Message from %s: %s", peerID, jsonContent)
				}
			}
		case hole.TypeHeartbeat:
			// 处理心跳消息
			var heartbeat hole.HeartbeatPayload
			if err := json.Unmarshal(msg.Payload, &heartbeat); err != nil {
				c.xl.Errorf("Failed to unmarshal heartbeat from peer %s: %v", peerID, err)
				continue
			}
			c.xl.Debugf("Heartbeat from %s: sent=%d, recv=%d",
				peerID, heartbeat.BytesSent, heartbeat.BytesRecv)
		default:
			c.xl.Warnf("Unknown message type from peer %s: %s", peerID, msg.Type)
		}
	}
}

func (c *Client) SendMessage(peerID string, message string) {
	// 检查是否已连接到peer
	conn, ok := c.peers.Load(peerID)
	if !ok {
		c.xl.Errorf("Not connected to peer %s", peerID)
		return
	}

	// 构造消息
	msg := &hole.Message{
		Type:    hole.TypeMessage,
		From:    c.clientID,
		To:      peerID,
		Payload: []byte(message), // 直接使用文本内容
	}

	// 序列化整个消息
	data, err := json.Marshal(msg)
	if err != nil {
		c.xl.Errorf("Failed to marshal message: %v", err)
		return
	}

	// 添加换行符并发送
	data = append(data, '\n')
	if _, err := conn.(net.Conn).Write(data); err != nil {
		c.xl.Errorf("Failed to send message to peer %s: %v", peerID, err)
		return
	}

	// 更新发送字节数
	c.addP2PBytesSent(int64(len(data)))

	c.xl.Infof("Message sent to %s: %s", peerID, message)
}

func (c *Client) Close() {
	// 停止心跳
	if c.heartbeatCancel != nil {
		c.heartbeatCancel()
	}

	// 关闭连接
	if c.serverConn != nil {
		c.serverConn.Close()
	}

	// 关闭监听器
	if c.listener != nil {
		c.listener.Close()
	}

	// 关闭所有对等连接
	c.peers.Range(func(key, value interface{}) bool {
		if conn, ok := value.(net.Conn); ok {
			conn.Close()
			c.xl.Infof("Connection with peer %s closed", key)
		}
		return true
	})
}

func (c *Client) StartCLI() {
	c.xl.Info("Starting CLI interface. Available commands:")
	c.xl.Info("  list              - List all connected peers")
	c.xl.Info("  connect <peer_id> - Connect to a peer")
	c.xl.Info("  send <peer_id> <message> - Send message to a peer")
	c.xl.Info("  exit              - Exit the program")

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}

		cmd := scanner.Text()
		parts := strings.Fields(cmd)
		if len(parts) == 0 {
			continue
		}

		switch parts[0] {
		case "list":
			c.handleListCommand()
		case "connect":
			if len(parts) != 2 {
				c.xl.Error("Usage: connect <peer_id>")
				continue
			}
			c.handleConnectCommand(parts[1])
		case "send":
			if len(parts) < 3 {
				c.xl.Error("Usage: send <peer_id> <message>")
				continue
			}
			c.handleSendCommand(parts[1], strings.Join(parts[2:], " "))
		case "exit":
			c.xl.Info("Exiting...")
			return
		default:
			c.xl.Errorf("Unknown command: %s", parts[0])
		}
	}
}

func (c *Client) handleListCommand() {
	count := 0
	c.peers.Range(func(key, value interface{}) bool {
		c.xl.Infof("Connected peer: %v", key)
		count++
		return true
	})
	if count == 0 {
		c.xl.Info("No peers connected")
	}
}

func (c *Client) handleConnectCommand(peerID string) {
	if peerID == c.clientID {
		c.xl.Warn("Cannot connect to self")
		return
	}

	// 检查是否已经连接
	if _, exists := c.peers.Load(peerID); exists {
		c.xl.Infof("Already connected to peer %s", peerID)
		return
	}

	// 尝试连接对等节点
	if err := c.ConnectToPeer(peerID); err != nil {
		c.xl.Errorf("Failed to connect to peer %s: %v", peerID, err)
	}
}

func (c *Client) handleSendCommand(peerID, message string) {
	peerConn, ok := c.peers.Load(peerID)
	if !ok {
		c.xl.Errorf("Peer %s not connected", peerID)
		return
	}

	if !strings.HasPrefix(message, "{") {
		message = fmt.Sprintf(`{"message": "%s"}`, message)
	}

	conn := peerConn.(net.Conn)
	msg := &hole.Message{
		Type:    hole.TypeMessage,
		From:    c.clientID,
		To:      peerID,
		Payload: []byte(message),
	}

	data, err := json.Marshal(msg)
	if err != nil {
		c.xl.Errorf("Failed to marshal message: %v", err)
		return
	}

	_, err = conn.Write(append(data, '\n'))
	if err != nil {
		c.xl.Errorf("Failed to send message: %v", err)
		return
	}

	c.addP2PBytesSent(int64(len(data)))

	c.xl.Infof("Message sent to %s: %s", peerID, message)
}

func (c *Client) register() error {
	// 注册客户端信息
	c.xl.Infof("Registering client (ID: %s, Name: %s)...", c.clientID, c.name)
	msg := &hole.Message{
		Type: hole.TypeRegister,
		From: c.clientID,
		Payload: []byte(`{
			"client_id": "` + c.clientID + `",
			"name": "` + c.name + `"
		}`),
	}
	packet, _ := hole.CreateHolePacket(msg)
	if err := protocol.NewPacketIO(nil, c.serverConn).WritePacket(packet); err != nil {
		c.xl.Errorf("Failed to register with server: %v", err)
		return fmt.Errorf("register to server error: %v", err)
	}
	c.xl.Info("Client registered successfully")
	return nil
}
