package handler

import "net"

type SpiderHandler struct {
	conn net.Conn
}

func NewSpiderHandler(conn net.Conn) *SpiderHandler {
	return &SpiderHandler{
		conn: conn,
	}
}

func (h *SpiderHandler) Start() error {

	return nil
}
