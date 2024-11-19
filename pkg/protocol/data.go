/*
Package protocol 定义了网络通信的核心接口

这个包提供了网络通信的基础抽象层，主要包含：
- ClientHandler：客户端通信处理接口，定义了数据包的处理、发送和接收行为
- ServerHandler：服务器通信处理接口，继承自ClientHandler，额外提供了服务器的启动和停止功能

使用方式：
1. 实现 ClientHandler 接口来处理客户端的网络通信
2. 实现 ServerHandler 接口来处理服务器端的网络通信
3. 配合 Packet 类型进行数据传输

@author: liuscraft
@date: 2024-03-19
*/
package protocol

type ClientHandler interface {
	Handle(packet Packet) error
	Send(packet Packet) error
	Receive(packet Packet) error
}

type ServerHandler interface {
	ClientHandler
	Start() error
	Stop() error
}
