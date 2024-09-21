package protocol

type Option func(packet Packet)

type Creator interface {
	NewProtocol(options ...Option) Packet
	PacketType() PacketType
	Gzip(packet Packet) bool
	Unzip(packet Packet) bool
}
