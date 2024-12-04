package api

import (
    "encoding/json"
    "net/http"
    "github.com/liuscraft/spider-network/server/client_mgr"
)

type TopologyAPI struct {
    clientMgr *client_mgr.ClientManager
}

func NewTopologyAPI(mgr *client_mgr.ClientManager) *TopologyAPI {
    return &TopologyAPI{
        clientMgr: mgr,
    }
}

// GetTopology 获取网络拓扑
func (api *TopologyAPI) GetTopology(w http.ResponseWriter, r *http.Request) {
    clients := api.clientMgr.GetClients()
    topology := make(map[string][]string)
    
    for id, client := range clients {
        topology[id] = client.Status.Peers
    }
    
    // 格式化为可视化数据
    data := api.formatTopologyData(topology)
    json.NewEncoder(w).Encode(data)
}

// 添加数据格式化方法
type TopologyData struct {
    Nodes []Node `json:"nodes"`
    Edges []Edge `json:"edges"`
}

type Node struct {
    ID    string `json:"id"`
    Label string `json:"label"`
}

type Edge struct {
    From string `json:"from"`
    To   string `json:"to"`
}

func (api *TopologyAPI) formatTopologyData(topology map[string][]string) TopologyData {
    data := TopologyData{
        Nodes: make([]Node, 0),
        Edges: make([]Edge, 0),
    }

    // 添加节点
    for id, client := range api.clientMgr.GetClients() {
        data.Nodes = append(data.Nodes, Node{
            ID:    id,
            Label: client.Name,
        })
    }

    // 添加边
    for from, peers := range topology {
        for _, to := range peers {
            data.Edges = append(data.Edges, Edge{
                From: from,
                To:   to,
            })
        }
    }

    return data
} 