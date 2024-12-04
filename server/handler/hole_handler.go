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
	"time"

	"bufio"

	"github.com/liuscraft/spider-network/pkg/config"
	"github.com/liuscraft/spider-network/pkg/protocol"
	"github.com/liuscraft/spider-network/pkg/protocol/hole"
	"github.com/liuscraft/spider-network/pkg/xlog"
	"github.com/liuscraft/spider-network/server/client_mgr"
	"github.com/liuscraft/spider-network/server/types"
)

type HoleHandler struct {
	config    config.HoleConfig
	listener  net.Listener
	clientMgr *client_mgr.ClientManager
}

func NewHoleHandler(config config.HoleConfig) (h *HoleHandler, err error) {
	listener, err := net.Listen("tcp", config.BindAddr)
	if err != nil {
		return nil, err
	}

	return &HoleHandler{
		config:    config,
		listener:  listener,
		clientMgr: client_mgr.NewClientManager(),
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
		h.clientMgr.HandleDisconnect(conn)
	}()

	reader := bufio.NewReader(conn)
	for {
		// 读取消息
		data, err := reader.ReadBytes('\n')
		if err != nil {
			if err != io.EOF {
				xl.Errorf("read error: %v", err)
			}
			return
		}

		// 解析消息
		var msg hole.Message
		if err := json.Unmarshal(data, &msg); err != nil {
			xl.Errorf("unmarshal error: %v", err)
			continue
		}

		// 处理消息
		switch msg.Type {
		case hole.TypeRegister:
			if err := h.handleRegister(xl, conn, &msg); err != nil {
				xl.Errorf("handle register error: %v", err)
				return
			}
		case hole.TypePunch:
			if err := h.handlePunch(xl, &msg); err != nil {
				xl.Errorf("handle punch error: %v", err)
				continue
			}
		case hole.TypeConnect:
			if err := h.handleConnect(xl, &msg); err != nil {
				xl.Errorf("handle connect error: %v", err)
				continue
			}
		case hole.TypeHeartbeat:
			if err := h.handleHeartbeat(xl, &msg); err != nil {
				xl.Errorf("handle heartbeat error: %v", err)
				continue
			}
		default:
			xl.Warnf("unknown message type: %s", msg.Type)
		}
	}
}

func (h *HoleHandler) handleRegister(xl xlog.Logger, conn net.Conn, msg *hole.Message) error {
	var payload hole.RegisterPayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		return fmt.Errorf("unmarshal register payload error: %v", err)
	}

	// 检查是否是重连
	if oldClient, exists := h.clientMgr.GetClient(payload.ClientID); exists {
		// 如果是重连，关闭旧连接
		if oldClient.Conn != nil {
			oldClient.Conn.Close()
		}
		xl.Infof("Client %s (%s) reconnected from %s", payload.ClientID, payload.Name, conn.RemoteAddr())
	} else {
		xl.Infof("New client %s (%s) registered from %s", payload.ClientID, payload.Name, conn.RemoteAddr())
	}

	// 创建或更新客户端信息
	client := types.NewClientInfo(conn, payload.ClientID, payload.Name)
	h.clientMgr.AddClient(client)

	// 发送注册确认
	response := &hole.Message{
		Type: hole.TypeRegister,
		From: "server",
		To:   payload.ClientID,
	}
	data, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("marshal register response error: %v", err)
	}

	if _, err := conn.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("write register response error: %v", err)
	}

	return nil
}

func (h *HoleHandler) handlePunch(xl xlog.Logger, msg *hole.Message) error {
	var payload hole.PunchPayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		xl.Errorf("parse punch payload error: %v", err)
		return err
	}

	// 查找目标客户端
	target, ok := h.clientMgr.GetClient(msg.To)
	if !ok {
		xl.Warnf("target client not found: %s", msg.To)
		return nil
	}

	// 向目标客户端转发打洞消息
	punchMsg := &hole.Message{
		Type:    hole.TypePunch,
		From:    msg.From,
		To:      msg.To,
		Payload: msg.Payload,
	}

	data, err := json.Marshal(punchMsg)
	if err != nil {
		xl.Errorf("marshal punch message error: %v", err)
		return err
	}

	if _, err := target.Conn.Write(append(data, '\n')); err != nil {
		xl.Errorf("write punch message error: %v", err)
		return err
	}

	xl.Infof("Forwarded punch message from %s to %s", msg.From, msg.To)
	return nil
}

func (h *HoleHandler) handleConnect(xl xlog.Logger, msg *hole.Message) error {
	// 查找目标客户端
	target, ok := h.clientMgr.GetClient(msg.To)
	if !ok {
		xl.Warnf("target client not found: %s", msg.To)
		return nil
	}

	// 转发连接请求
	data, err := json.Marshal(msg)
	if err != nil {
		xl.Errorf("marshal connect message error: %v", err)
		return err
	}

	if _, err := target.Conn.Write(append(data, '\n')); err != nil {
		xl.Errorf("write connect message error: %v", err)
		return err
	}

	xl.Infof("Forwarded connect message from %s to %s", msg.From, msg.To)
	return nil
}

func (h *HoleHandler) handleHeartbeat(xl xlog.Logger, msg *hole.Message) error {
    // 从消息中获取客户端ID
    clientID := msg.From
    if clientID == "" {
        xl.Error("heartbeat message missing client ID")
        return nil
    }

    var heartbeat hole.HeartbeatPayload
    if err := json.Unmarshal(msg.Payload, &heartbeat); err != nil {
        xl.Errorf("unmarshal heartbeat data error: %v", err)
        return nil
    }

    // 验证客户端ID是否匹配
    if clientID != heartbeat.ClientID {
        xl.Errorf("client ID mismatch: message from %s but heartbeat data claims %s", 
            clientID, heartbeat.ClientID)
        return nil
    }

    // 更新客户端状态
    client, ok := h.clientMgr.GetClient(clientID)
    if !ok {
        xl.Warnf("heartbeat from unknown client: %s", clientID)
        return nil
    }

    // 计算延迟
    latency := time.Now().UnixNano() - heartbeat.Timestamp
    latencyMs := latency / int64(time.Millisecond)

    // 验证并过滤 peers 列表
    validPeers := make([]string, 0)
    for _, peerID := range heartbeat.Peers {
        // 确保 peer 存在且不是自己
        if peerID != clientID {
            if _, exists := h.clientMgr.GetClient(peerID); exists {
                validPeers = append(validPeers, peerID)
            }
        }
    }

    // 计算传输速率
    now := time.Now()
    timeSinceLastUpdate := now.Sub(client.Status.LastSeen)
    var bytesRate float64
    var p2pBytesRate float64
    if timeSinceLastUpdate > 0 {
        // 计算总传输速率
        totalBytesDelta := (heartbeat.BytesSent + heartbeat.BytesRecv) - 
            (client.Status.BytesSent + client.Status.BytesRecv)
        bytesRate = float64(totalBytesDelta) / timeSinceLastUpdate.Seconds()

        // 计算点对点传输速率
        p2pBytesDelta := (heartbeat.P2PBytesSent + heartbeat.P2PBytesRecv) - 
            (client.Status.P2PBytesSent + client.Status.P2PBytesRecv)
        p2pBytesRate = float64(p2pBytesDelta) / timeSinceLastUpdate.Seconds()
    }

    // 更新客户端状态
    status := types.ClientStatus{
        Connected:     true,
        LastSeen:     now,
        Peers:        validPeers,
        BytesSent:    heartbeat.BytesSent,
        BytesRecv:    heartbeat.BytesRecv,
        P2PBytesSent: heartbeat.P2PBytesSent,
        P2PBytesRecv: heartbeat.P2PBytesRecv,
        BytesRate:    bytesRate,
        P2PBytesRate: p2pBytesRate,
        Latency:      latencyMs,
        ConnectedAt:  client.Status.ConnectedAt,
        LastError:    client.Status.LastError,
        LastErrorTime: client.Status.LastErrorTime,
        PunchStatus:  client.Status.PunchStatus,
    }
    h.clientMgr.UpdateClientStatus(clientID, status)

    xl.Debugf("Updated client status: id=%s, latency=%dms, server=%d/%d (%.2f B/s), p2p=%d/%d (%.2f B/s), peers=%v",
        clientID, latencyMs, 
        heartbeat.BytesSent, heartbeat.BytesRecv, bytesRate,
        heartbeat.P2PBytesSent, heartbeat.P2PBytesRecv, p2pBytesRate,
        validPeers)

    return nil
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

func (h *HoleHandler) GetClientManager() *client_mgr.ClientManager {
	return h.clientMgr
}
