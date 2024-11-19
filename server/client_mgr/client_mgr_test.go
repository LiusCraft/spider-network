package clientmgr

import (
	"net"
	"testing"

	"github.com/liuscraft/spider-network/server/types"
)

// 创建模拟的 net.Conn 用于测试
type mockConn struct {
	net.Conn
	addr net.Addr
}

type mockAddr struct {
	network string
	address string
}

func (a *mockAddr) Network() string { return a.network }
func (a *mockAddr) String() string  { return a.address }

func newMockClient(addr string) *types.Client {
	mockAddr := &mockAddr{network: "tcp", address: addr}
	conn := &mockConn{addr: mockAddr}
	return types.NewClient(conn)
}

func TestClientMgr(t *testing.T) {
	cm := NewClientMgr()

	// 测试初始状态
	if cm.GetClientCount() != 0 {
		t.Errorf("期望客户端数量为 0，实际为 %d", cm.GetClientCount())
	}

	// 测试添加客户端
	client1 := newMockClient("127.0.0.1:8001")
	client2 := newMockClient("127.0.0.1:8002")
	
	cm.AddClient(client1)
	if cm.GetClientCount() != 1 {
		t.Errorf("添加一个客户端后，期望客户端数量为 1，实际为 %d", cm.GetClientCount())
	}

	cm.AddClient(client2)
	if cm.GetClientCount() != 2 {
		t.Errorf("添加两个客户端后，期望客户端数量为 2，实际为 %d", cm.GetClientCount())
	}

	// 测试获取客户端
	if cm.GetClient("127.0.0.1:8001") != client1 {
		t.Error("获取客户端失败")
	}

	// 测试获取所有客户端
	allClients := cm.GetAllClients()
	if len(allClients) != 2 {
		t.Errorf("获取所有客户端失败，期望数量为 2，实际为 %d", len(allClients))
	}

	// 测试移除客户端
	cm.RemoveClient(client1)
	if cm.GetClientCount() != 1 {
		t.Errorf("移除一个客户端后，期望客户端数量为 1，实际为 %d", cm.GetClientCount())
	}

	// 测试关闭所有客户端
	cm.CloseAllClients()
	// 注意：实际测试中可能需要添加更多的验证来确保客户端确实被正确关闭
}

func (m *mockConn) RemoteAddr() net.Addr {
	return m.addr
}

func (m *mockConn) Close() error {
	return nil
}

