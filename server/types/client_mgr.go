package types

import "net"

type Client struct {
	conn net.Conn
}

func NewClient(conn net.Conn) *Client {
	return &Client{
		conn: conn,
	}
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) GetConn() net.Conn {
	return c.conn
}
