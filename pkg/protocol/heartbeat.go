package protocol

const (
    MessageType_HEARTBEAT = "heartbeat"
)

// HeartbeatData 心跳数据
type HeartbeatData struct {
    ClientID   string   `json:"client_id"`   // 客户端ID
    BytesSent  int64    `json:"bytes_sent"`  // 已发送字节数
    BytesRecv  int64    `json:"bytes_recv"`  // 已接收字节数
    Peers      []string `json:"peers"`       // 当前连接的节点
    Timestamp  int64    `json:"timestamp"`   // 时间戳
}
