package handler

import (
	"html/template"
	"net/http"

	"github.com/liuscraft/spider-network/pkg/xlog"
	"github.com/liuscraft/spider-network/server/client_mgr"
)

type ClientHandler struct {
	clientMgr *client_mgr.ClientManager
	templates *template.Template
}

func NewClientHandler(mgr *client_mgr.ClientManager, tmpl *template.Template) *ClientHandler {
	return &ClientHandler{
		clientMgr: mgr,
		templates: tmpl,
	}
}

func (h *ClientHandler) HandleList(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"Title":      "客户端管理",
		"Clients":    h.clientMgr.GetClients(),
		"ContentTpl": "content-clients",
	}
	if err := h.templates.ExecuteTemplate(w, "base", data); err != nil {
		xlog.Debug(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func (h *ClientHandler) HandleDetail(w http.ResponseWriter, r *http.Request) {
	clientID := r.URL.Query().Get("id")
	client, ok := h.clientMgr.GetClient(clientID)
	if !ok {
		http.NotFound(w, r)
		return
	}

	data := map[string]interface{}{
		"Title":  "客户端详情",
		"Client": client,
	}
	if err := h.templates.ExecuteTemplate(w, "base", data); err != nil {
		xlog.Debug(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
