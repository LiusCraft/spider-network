package handler

import (
	"github.com/liuscraft/spider-network/pkg/protocol"
	"github.com/liuscraft/spider-network/server/types"
)

type SpiderHandler struct {
	client *types.ClientInfo
}

func NewSpiderHandler(client *types.ClientInfo) *SpiderHandler {
	return &SpiderHandler{
		client: client,
	}
}

func (h *SpiderHandler) Start() error {
	// 实现启动逻辑
	return nil
}

func (h *SpiderHandler) Stop() error {
	// 实现停止逻辑
	return nil
}

func (h *SpiderHandler) Handle(packet protocol.Packet) error {
	// 实现消息处理逻辑
	return nil
}
