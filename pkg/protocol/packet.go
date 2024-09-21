package protocol

type Packet interface {
	/* Read write to target packet
	example:
		bytesPacket to stringPacket(target packet)
		src := &bytesPacket{}
	    src.Write([]byte("hello world"))
		targetPacket := &stringPacket{}
		src.Read(targetPacket)
		result := ""
		targetPacket.Read(result)
		fmt.Println(result) => "hello world"
	*/
	Read(p interface{}) (n int, err error)
	/* Write to the packet
	example:
		bytePacket := &bytePacket{}
	    bytePacket.Write([]byte("hello world")) // The internal implementation deals with data of type []byte
		stringPacket := &stringPacket{}
	    stringPacket.Write("hello world") // The internal implementation deals with data of type string
	*/
	Write(p interface{}) (n int, err error)
	// ToPacket Writes bytes to change the current packet content
	ToPacket(p []byte) (n int, err error)
	// Bytes get the packet to bytes
	Bytes() []byte
	// PacketSize get the packet size
	PacketSize() int
	// PacketType default is BytesPacket, also zero
	PacketType() PacketType
	Clear()
}
