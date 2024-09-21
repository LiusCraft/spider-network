package client

import (
	"net"

	"github.com/liuscraft/spider-network/pkg/model"
	"github.com/liuscraft/spider-network/pkg/protocol"
	"github.com/liuscraft/spider-network/pkg/protocol/packet"
	"github.com/liuscraft/spider-network/pkg/protocol/packet_io"
	"github.com/liuscraft/spider-network/pkg/xlog"
)

type Client struct {
}

func NewClient() *Client {
	// tcp connection
	xl := xlog.New()
	conn, err := net.Dial("tcp", ":19730")
	if err != nil {
		xl.Error("Error connecting to server:", err)
		return nil
	}
	defer conn.Close()

	body := "test"
	bytesPacket, err := packet.CreateProtocol(protocol.BytesType)
	if err != nil {
		xl.Errorf("Error creating protocol, error: %v", err)
		return nil
	}
	_, err = bytesPacket.Write([]byte(body))
	if err != nil {
		xl.Errorf("Error writing to server, error: %v", err)
		return nil
	}
	packet_io.WritePacket(conn, bytesPacket, true)
	_, err = bytesPacket.Write([]byte("test"))
	if err != nil {
		xl.Errorf("Error writing to server, error: %v", err)
		return nil
	}
	packet_io.WritePacket(conn, bytesPacket)
	jsonProtocol, _ := packet.CreateProtocol(protocol.JsonType)
	jsonProtocol.Write(model.UserModel{ID: 22, Username: "小明"})
	_, err = packet_io.WritePacket(conn, jsonProtocol)
	if err != nil {
		xl.Errorf("Error writing to server, error: %v", err)
		return nil
	}
	receivePacket, err := packet_io.ReceivePacket(conn)
	if err != nil {
		xl.Errorf("Error receiving bytesPacket, error: %v", err)
		return nil
	}
	if receivePacket.PacketType() == protocol.JsonType {
		jsonProtocol.ToPacket(receivePacket.Bytes())
		result := &model.UserModel{}
		jsonProtocol.Read(result)
		xl.Infof("receive bytesPacket: %+v", result)
	}

	return &Client{}
}
