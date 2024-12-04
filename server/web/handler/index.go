package handler

import (
	"html/template"
	"net/http"
)

type IndexHandler struct {
	templates *template.Template
}

func NewIndexHandler(tmpl *template.Template) *IndexHandler {
	return &IndexHandler{
		templates: tmpl,
	}
}

func (h *IndexHandler) HandleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	data := map[string]interface{}{
		"Title": "首页",
		"ContentTpl": "content-index",
	}
	if err := h.templates.ExecuteTemplate(w, "base", data); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
