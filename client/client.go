package client

import (
	"net"

	"github.com/liuscraft/spider-network/pkg/protocol"
	"github.com/liuscraft/spider-network/pkg/xlog"
)

type Client struct {
}

func NewClient() *Client {
	// tcp connection
	xl := xlog.NewLogger()
	conn, err := net.Dial("tcp", ":19730")
	if err != nil {
		xl.Error("Error connecting to server:", err)
		return nil
	}
	defer conn.Close()

	body := "test"
	protocol, err := protocol.CreateProtocol(protocol.BytesType, len(body))
	if err != nil {
		xl.Errorf("Error creating protocol, error: %v", err)
		return nil
	}
	_, err = protocol.Write([]byte(body))
	if err != nil {
		xl.Errorf("Error writing to server, error: %v", err)
		return nil
	}
	_, err = protocol.Writer(conn)
	if err != nil {
		xl.Errorf("Error writing to server, error: %v", err)
		return nil
	}
	return &Client{}
}
