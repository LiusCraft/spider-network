package handler

import (
	"html/template"
	"net/http"

	"github.com/liuscraft/spider-network/pkg/xlog"
)

type BaseHandler struct {
	xl        xlog.Logger
	templates *template.Template
}

func NewBaseHandler(tmpl *template.Template) BaseHandler {
	return BaseHandler{
		xl:        xlog.New(),
		templates: tmpl,
	}
}

func (h *BaseHandler) handleError(w http.ResponseWriter, err error, status int) {
	h.xl.Errorf("Error: %v", err)
	http.Error(w, err.Error(), status)
}

func (h *BaseHandler) render(w http.ResponseWriter, tmpl string, data interface{}) {
	if err := h.templates.ExecuteTemplate(w, tmpl, data); err != nil {
		h.handleError(w, err, http.StatusInternalServerError)
	}
}
