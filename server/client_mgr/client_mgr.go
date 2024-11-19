/*
	@author: liuscraft
	@date: 2024-11-19
	ClientMgr is a manager for clients
	1. add client
	2. remove client
	3. get client
	4. get all clients
	5. close all clients
	6. get client count
*/

package clientmgr

import "github.com/liuscraft/spider-network/server/types"

type ClientMgr struct {
	clients map[string]*types.Client
}

func NewClientMgr() *ClientMgr {
	return &ClientMgr{
		clients: make(map[string]*types.Client),
	}
}

func (cm *ClientMgr) AddClient(client *types.Client) {
	cm.clients[client.GetConn().RemoteAddr().String()] = client
}

func (cm *ClientMgr) RemoveClient(client *types.Client) {
	delete(cm.clients, client.GetConn().RemoteAddr().String())
}

func (cm *ClientMgr) GetClient(addr string) *types.Client {
	return cm.clients[addr]
}

func (cm *ClientMgr) GetAllClients() map[string]*types.Client {
	return cm.clients
}

func (cm *ClientMgr) CloseAllClients() {
	for _, client := range cm.clients {
		client.Close()
	}
}

func (cm *ClientMgr) GetClientCount() int {
	return len(cm.clients)
}
