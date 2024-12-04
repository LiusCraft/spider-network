package api

import (
    "html/template"
    "net/http"
    "github.com/liuscraft/spider-network/server/client_mgr"
)

type ClientAPI struct {
    clientMgr *client_mgr.ClientManager
    templates *template.Template
}

func NewClientAPI(mgr *client_mgr.ClientManager, tmpl *template.Template) *ClientAPI {
    return &ClientAPI{
        clientMgr: mgr,
        templates: tmpl,
    }
}

// GetClients 获取所有客户端列表
func (api *ClientAPI) GetClients(w http.ResponseWriter, r *http.Request) {
    data := map[string]interface{}{
        "Clients": api.clientMgr.GetClients(),
    }
    if err := api.templates.ExecuteTemplate(w, "client_list", data); err != nil {
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }
}

// GetClientDetail 获取客户端详情
func (api *ClientAPI) GetClientDetail(w http.ResponseWriter, r *http.Request) {
    clientID := r.URL.Query().Get("id")
    client, ok := api.clientMgr.GetClient(clientID)
    if !ok {
        http.Error(w, "Client not found", http.StatusNotFound)
        return
    }
    data := map[string]interface{}{
        "Client": client,
    }
    if err := api.templates.ExecuteTemplate(w, "client_detail", data); err != nil {
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }
}