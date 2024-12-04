/*
	@author: liuscraft
	@date: 2024-11-19
	HoleHandler 是 SpiderHole(Service) 的 Handler
	1. 处理客户端连接
	2. 处理客户端发送的数据
	3. 处理客户端接收的数据
*/

package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"sync"

	"github.com/liuscraft/spider-network/pkg/config"
	"github.com/liuscraft/spider-network/pkg/protocol"
	"github.com/liuscraft/spider-network/pkg/protocol/hole"
	"github.com/liuscraft/spider-network/pkg/protocol/packet_io"
	"github.com/liuscraft/spider-network/pkg/xlog"
)

type HoleHandler struct {
	config   config.HoleConfig
	listener net.Listener
	clients  sync.Map // 存储已连接的客户端信息
}

type clientInfo struct {
	conn       net.Conn
	clientID   string
	name       string
	publicAddr string
}

func NewHoleHandler(config config.HoleConfig) (h *HoleHandler, err error) {
	xl := xlog.New()
	xl.Info("spider-hole service starting...")
	xl.Infof("spider-hole service listening on %s", config.BindAddr)
	listener, err := net.Listen("tcp", config.BindAddr)
	if err != nil {
		xl.Errorf("spider-hole service listen error: %v", err)
		return nil, err
	}
	return &HoleHandler{
		listener: listener,
		config:   config,
	}, nil
}

func (h *HoleHandler) Start() error {
	xl := xlog.New()

	for {
		conn, err := h.listener.Accept()
		if err != nil {
			xl.Errorf("spider-hole service accept error: %v", err)
			continue
		}
		go h.acceptSpiderConn(xl, conn)
	}
}

func (h *HoleHandler) acceptSpiderConn(xl xlog.Logger, conn net.Conn) {
	xl.Infof("spider-hole service accept connection from %s", conn.RemoteAddr())
	xl = xlog.WithLogId(xl, fmt.Sprintf("spider-hole-conn[%s]", conn.RemoteAddr().String()))
	defer func() {
		conn.Close()
		// 清理客户端信息
		h.clients.Range(func(key, value interface{}) bool {
			if info, ok := value.(*clientInfo); ok && info.conn == conn {
				h.clients.Delete(key)
				return false
			}
			return true
		})
	}()

	for {
		receivePacket, err := packet_io.ReceivePacket(conn)
		if err != nil {
			if err == io.EOF {
				xl.Warnf("spider-hole-conn leave connection")
				break
			}
			xl.Errorf("read packet error: %v", err)
			return
		}

		if receivePacket.PacketType() != protocol.JsonType {
			xl.Warnf("invalid packet type: %v", receivePacket.PacketType())
			continue
		}

		var msg hole.HoleMessage
		if _, err := receivePacket.Read(&msg); err != nil {
			xl.Errorf("parse hole message error: %v", err)
			continue
		}

		switch msg.Type {
		case hole.TypeRegister:
			h.handleRegister(xl, conn, &msg)
		case hole.TypePunch:
			h.handlePunch(xl, conn, &msg)
		case hole.TypeConnect:
			h.handleConnect(xl, conn, &msg)
		default:
			xl.Warnf("unknown message type: %s", msg.Type)
		}
	}
}

func (h *HoleHandler) handleRegister(xl xlog.Logger, conn net.Conn, msg *hole.HoleMessage) {
	var payload hole.RegisterPayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		xl.Errorf("parse register payload error: %v", err)
		return
	}

	info := &clientInfo{
		conn:       conn,
		clientID:   payload.ClientID,
		name:       payload.Name,
		publicAddr: conn.RemoteAddr().String(),
	}
	h.clients.Store(payload.ClientID, info)
	xl.Infof("client registered: %+v", info)

	// 发送注册成功响应
	response := &hole.HoleMessage{
		Type: hole.TypeRegister,
		From: "server",
		To:   payload.ClientID,
	}
	packet, _ := hole.CreateHolePacket(response)
	packet_io.WritePacket(conn, packet)
}

func (h *HoleHandler) handlePunch(xl xlog.Logger, conn net.Conn, msg *hole.HoleMessage) {
	var payload hole.PunchPayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		xl.Errorf("parse punch payload error: %v", err)
		return
	}

	// 查找目标客户端
	targetValue, ok := h.clients.Load(msg.To)
	if !ok {
		xl.Warnf("target client not found: %s", msg.To)
		return
	}
	target := targetValue.(*clientInfo)

	// 向目标客户端发送打洞消息
	punchMsg := &hole.HoleMessage{
		Type: hole.TypePunchReady,
		From: msg.From,
		To:   msg.To,
		Payload: json.RawMessage(`{
			"public_addr": "` + payload.PublicAddr + `",
			"private_addr": "` + payload.PrivateAddr + `"
		}`),
	}
	packet, _ := hole.CreateHolePacket(punchMsg)
	packet_io.WritePacket(target.conn, packet)
}

func (h *HoleHandler) handleConnect(xl xlog.Logger, conn net.Conn, msg *hole.HoleMessage) {
	// 查找目标客户端
	targetValue, ok := h.clients.Load(msg.To)
	if !ok {
		xl.Warnf("target client not found: %s", msg.To)
		return
	}
	target := targetValue.(*clientInfo)

	// 转发连接请求
	packet, _ := hole.CreateHolePacket(msg)
	packet_io.WritePacket(target.conn, packet)
}

func (h *HoleHandler) Stop() error {
	if h.listener == nil {
		return nil
	}
	xlog.Info("spider-hole service stopping...")
	return h.listener.Close()
}

func (h *HoleHandler) Handle(packet protocol.Packet) error {
	panic("TODO: Implement")
}

func (h *HoleHandler) Send(packet protocol.Packet) error {
	panic("TODO: Implement")
}

func (h *HoleHandler) Receive(packet protocol.Packet) error {
	panic("TODO: Implement")
}
