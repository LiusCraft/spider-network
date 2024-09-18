package client

import (
	"net"

	"github.com/liuscraft/spider-network/pkg/model"
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
	packet, err := protocol.CreateProtocol(protocol.BytesType)
	if err != nil {
		xl.Errorf("Error creating protocol, error: %v", err)
		return nil
	}
	_, err = packet.Write([]byte(body))
	if err != nil {
		xl.Errorf("Error writing to server, error: %v", err)
		return nil
	}
	protocol.WritePacket(conn, packet, true)
	_, err = packet.Write([]byte("test"))
	if err != nil {
		xl.Errorf("Error writing to server, error: %v", err)
		return nil
	}
	protocol.WritePacket(conn, packet)
	jsonProtocol := protocol.NewJsonProtocol()
	jsonProtocol.Write(model.UserModel{ID: 22, Username: "小明"})
	_, err = protocol.WritePacket(conn, jsonProtocol)
	if err != nil {
		xl.Errorf("Error writing to server, error: %v", err)
		return nil
	}
	receivePacket, err := protocol.ReceivePacket(conn)
	if err != nil {
		xl.Errorf("Error receiving packet, error: %v", err)
		return nil
	}
	if receivePacket.PacketType() == protocol.JsonType {
		receiveJson := protocol.NewJsonProtocol()
		receiveJson.ToPacket(receivePacket.Bytes())
		result := &model.UserModel{}
		receiveJson.Read(result)
		xl.Infof("receive packet: %+v", result)
	}

	return &Client{}
}
