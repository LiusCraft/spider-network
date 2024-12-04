package client

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/liuscraft/spider-network/pkg/protocol/hole"
	"github.com/liuscraft/spider-network/pkg/protocol/packet_io"
	"github.com/liuscraft/spider-network/pkg/xlog"
)

type Client struct {
	clientID   string
	name       string
	serverConn net.Conn
	peers      sync.Map
	xl         xlog.Logger
	listener   net.Listener
}

func NewClient(clientID, name string) *Client {
	return &Client{
		clientID: clientID,
		name:     name,
		xl:       xlog.New(),
	}
}

func (c *Client) Connect(serverAddr string) error {
	c.xl.Infof("Connecting to server at %s...", serverAddr)

	// 首先创建一个监听器
	listener, err := net.Listen("tcp", "127.0.0.1:0") // 使用0让系统分配端口
	if err != nil {
		return fmt.Errorf("failed to create listener: %v", err)
	}
	c.listener = listener
	c.xl.Infof("Listening for peer connections on %s", listener.Addr())

	// 启动监听协程
	go c.acceptPeerConnections()

	// 连接到服务器
	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		c.listener.Close()
		return fmt.Errorf("connect to server error: %v", err)
	}
	c.serverConn = conn
	c.xl.Info("Connected to server successfully")

	// 注册客户端
	c.xl.Infof("Registering client (ID: %s, Name: %s)...", c.clientID, c.name)
	if err := c.register(); err != nil {
		c.serverConn.Close()
		c.listener.Close()
		return fmt.Errorf("register error: %v", err)
	}
	c.xl.Info("Client registered successfully")

	// 启动消息处理
	c.xl.Info("Starting message handler...")
	go c.handleServerMessages()

	return nil
}

func (c *Client) acceptPeerConnections() {
	for {
		conn, err := c.listener.Accept()
		if err != nil {
			if !strings.Contains(err.Error(), "use of closed network connection") {
				c.xl.Errorf("Failed to accept peer connection: %v", err)
			}
			return
		}
		c.xl.Infof("Accepted peer connection from %s", conn.RemoteAddr())
		
		// 等待对方发送身份信息
		go func(conn net.Conn) {
			scanner := bufio.NewScanner(conn)
			if !scanner.Scan() {
				conn.Close()
				return
			}
			peerID := scanner.Text()
			c.xl.Infof("Peer identified as: %s", peerID)
			
			// 发送我们的身份
			fmt.Fprintf(conn, "%s\n", c.clientID)
			
			// 存储连接
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
		PublicAddr:  c.listener.Addr().String(),  // 使用我们的监听地址
		PrivateAddr: c.listener.Addr().String(),  // 同上
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal punch payload failed: %v", err)
	}

	msg := &hole.HoleMessage{
		Type:    hole.TypePunch,
		From:    c.clientID,
		To:      peerID,
		Payload: payloadBytes,
	}

	packet, err := hole.CreateHolePacket(msg)
	if err != nil {
		return fmt.Errorf("create hole packet failed: %v", err)
	}

	if _, err := packet_io.WritePacket(c.serverConn, packet); err != nil {
		return fmt.Errorf("send punch message failed: %v", err)
	}

	c.xl.Infof("Punch message sent to %s", peerID)
	return nil
}

func (c *Client) handleServerMessages() {
	c.xl.Info("Message handler started")
	for {
		packet, err := packet_io.ReceivePacket(c.serverConn)
		if err != nil {
			c.xl.Errorf("Failed to receive server message: %v", err)
			return
		}

		var msg hole.HoleMessage
		if _, err := packet.Read(&msg); err != nil {
			c.xl.Errorf("Failed to parse hole message: %v", err)
			continue
		}

		c.xl.Infof("Received message: Type=%s, From=%s, To=%s", msg.Type, msg.From, msg.To)

		switch msg.Type {
		case hole.TypeRegister:
			c.xl.Info("Registration confirmed by server")
		case hole.TypePunchReady:
			c.xl.Info("Received punch ready message")
			go c.handlePunchReady(&msg)
		case hole.TypeConnect:
			c.xl.Info("Received connect message")
			go c.handleConnect(&msg)
		default:
			c.xl.Warnf("Unknown message type: %s", msg.Type)
		}
	}
}

func (c *Client) handlePunchReady(msg *hole.HoleMessage) {
	c.xl.Infof("Processing punch ready message from %s", msg.From)
	
	var payload hole.PunchPayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		c.xl.Errorf("Failed to parse punch ready payload: %v", err)
		return
	}

	c.xl.Infof("Attempting to connect to addresses: Public=%s, Private=%s", payload.PublicAddr, payload.PrivateAddr)
	
	// 尝试连接对方的公网和内网地址
	addresses := []string{payload.PublicAddr, payload.PrivateAddr}
	var conn net.Conn
	var err error

	for _, addr := range addresses {
		c.xl.Infof("Trying to connect to %s...", addr)
		conn, err = net.DialTimeout("tcp", addr, 5*time.Second)
		if err == nil {
			c.xl.Infof("Successfully connected to %s", addr)
			break
		}
		c.xl.Warnf("Failed to connect to %s: %v", addr, err)
	}

	if err != nil {
		c.xl.Errorf("Failed to connect to any peer address: %v", err)
		return
	}

	// 发送我们的身份
	fmt.Fprintf(conn, "%s\n", c.clientID)

	// 等待对方的身份确认
	scanner := bufio.NewScanner(conn)
	if !scanner.Scan() {
		c.xl.Error("Failed to receive peer identity")
		conn.Close()
		return
	}
	peerID := scanner.Text()
	if peerID != msg.From {
		c.xl.Errorf("Received unexpected peer ID: %s (expected %s)", peerID, msg.From)
		conn.Close()
		return
	}

	// 存储对等连接
	c.peers.Store(msg.From, conn)
	c.xl.Infof("Peer connection established with %s", msg.From)

	// 启动消息处理
	c.startPeerMessageHandler(msg.From, conn)

	// 发送连接确认消息
	confirmMsg := &hole.HoleMessage{
		Type: hole.TypeConnect,
		From: c.clientID,
		To:   msg.From,
	}
	packet, _ := hole.CreateHolePacket(confirmMsg)
	if _, err := packet_io.WritePacket(c.serverConn, packet); err != nil {
		c.xl.Errorf("Failed to send connect confirmation: %v", err)
		return
	}
	c.xl.Infof("Connection confirmation sent to %s", msg.From)
}

func (c *Client) handleConnect(msg *hole.HoleMessage) {
	c.xl.Infof("Received connect message from %s", msg.From)
	
	// 如果已经建立了连接，不需要再处理
	if _, exists := c.peers.Load(msg.From); exists {
		c.xl.Infof("Connection with %s already exists", msg.From)
		return
	}

	// 等待一段时间，让对方有机会建立连接
	time.Sleep(time.Second)

	// 如果还没有连接，说明可能需要我们主动连接
	if _, exists := c.peers.Load(msg.From); !exists {
		c.xl.Infof("No connection established with %s, trying to connect...", msg.From)
		if err := c.ConnectToPeer(msg.From); err != nil {
			c.xl.Errorf("Failed to connect to peer %s: %v", msg.From, err)
		}
	}
}

func (c *Client) Close() {
	if c.serverConn != nil {
		c.serverConn.Close()
	}

	// 关闭所有对等连接
	c.peers.Range(func(key, value interface{}) bool {
		if conn, ok := value.(net.Conn); ok {
			conn.Close()
		}
		return true
	})
	c.listener.Close()
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
	if err := c.ConnectToPeer(peerID); err != nil {
		c.xl.Errorf("Failed to connect to peer: %v", err)
		return
	}
	c.xl.Infof("Connection request sent to %s", peerID)
}

func (c *Client) handleSendCommand(peerID, message string) {
	peerConn, ok := c.peers.Load(peerID)
	if !ok {
		c.xl.Errorf("Peer %s not connected", peerID)
		return
	}

	conn := peerConn.(net.Conn)
	_, err := conn.Write([]byte(message + "\n"))
	if err != nil {
		c.xl.Errorf("Failed to send message: %v", err)
		return
	}
	c.xl.Infof("Message sent to %s", peerID)
}

func (c *Client) startPeerMessageHandler(peerID string, conn net.Conn) {
	go func() {
		scanner := bufio.NewScanner(conn)
		for scanner.Scan() {
			message := scanner.Text()
			c.xl.Infof("Message from %s: %s", peerID, message)
		}
		if err := scanner.Err(); err != nil {
			c.xl.Errorf("Error reading from peer %s: %v", peerID, err)
		}
		conn.Close()
		c.peers.Delete(peerID)
		c.xl.Infof("Connection with peer %s closed", peerID)
	}()
}

func (c *Client) register() error {
	// 注册客户端信息
	c.xl.Infof("Registering client (ID: %s, Name: %s)...", c.clientID, c.name)
	msg := &hole.HoleMessage{
		Type: hole.TypeRegister,
		From: c.clientID,
		Payload: []byte(`{
			"client_id": "` + c.clientID + `",
			"name": "` + c.name + `"
		}`),
	}
	packet, _ := hole.CreateHolePacket(msg)
	if _, err := packet_io.WritePacket(c.serverConn, packet); err != nil {
		c.xl.Errorf("Failed to register with server: %v", err)
		return fmt.Errorf("register to server error: %v", err)
	}
	c.xl.Info("Client registered successfully")
	return nil
}
