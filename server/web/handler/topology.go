package handler

import (
	"html/template"
	"net/http"

	"github.com/liuscraft/spider-network/server/client_mgr"
)

type TopologyHandler struct {
	clientMgr *client_mgr.ClientManager
	templates *template.Template
}

func NewTopologyHandler(mgr *client_mgr.ClientManager, tmpl *template.Template) *TopologyHandler {
	return &TopologyHandler{
		clientMgr: mgr,
		templates: tmpl,
	}
}

func (h *TopologyHandler) HandleView(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"Title": "网络拓扑",
		"ContentTpl": "content-topology",
	}
	if err := h.templates.ExecuteTemplate(w, "base", data); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
