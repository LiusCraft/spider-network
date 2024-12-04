package web

import (
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"

	"github.com/liuscraft/spider-network/server/client_mgr"
	"github.com/liuscraft/spider-network/server/web/api"
	"github.com/liuscraft/spider-network/server/web/handler"
)

type Server struct {
	clientMgr *client_mgr.ClientManager
	templates *template.Template
	baseDir   string

	// API handlers
	clientAPI *api.ClientAPI
	topoAPI   *api.TopologyAPI

	// Page handlers
	indexHandler    *handler.IndexHandler
	clientHandler   *handler.ClientHandler
	topologyHandler *handler.TopologyHandler
}

func NewServer(mgr *client_mgr.ClientManager, baseDir string) (*Server, error) {
	// 创建基础模板
	baseTemplate := template.New("base").Funcs(template.FuncMap{
		"div": func(a, b int64) float64 {
			return float64(a) / float64(b)
		},
	}).Funcs(api.GetTemplateFuncs())
	
	// 加载所有模板文件
	layoutFiles, err := filepath.Glob(baseDir + "/web/templates/layout/*.html")
	if err != nil {
		return nil, fmt.Errorf("find layout templates error: %v", err)
	}
	pageFiles, err := filepath.Glob(baseDir + "/web/templates/pages/*.html")
	if err != nil {
		return nil, fmt.Errorf("find page templates error: %v", err)
	}
	componentFiles, err := filepath.Glob(baseDir + "/web/templates/components/*.html")
	if err != nil {
		return nil, fmt.Errorf("find component templates error: %v", err)
	}

	// 合并所有模板文件路径
	templateFiles := append(layoutFiles, append(pageFiles, componentFiles...)...)

	// 解析所有模板
	tmpl, err := baseTemplate.ParseFiles(templateFiles...)
	if err != nil {
		return nil, fmt.Errorf("parse templates error: %v", err)
	}

	s := &Server{
		clientMgr: mgr,
		templates: tmpl,
		baseDir:   baseDir,

		// Initialize API handlers
		clientAPI: api.NewClientAPI(mgr, tmpl),
		topoAPI:   api.NewTopologyAPI(mgr),

		// Initialize page handlers
		indexHandler:    handler.NewIndexHandler(tmpl),
		clientHandler:   handler.NewClientHandler(mgr, tmpl),
		topologyHandler: handler.NewTopologyHandler(mgr, tmpl),
	}

	return s, nil
}

func (s *Server) Start(addr string) error {
	// 启动心跳检测
	s.clientMgr.StartHeartbeat()

	// Page routes
	http.HandleFunc("/", s.indexHandler.HandleIndex)
	http.HandleFunc("/clients", s.clientHandler.HandleList)
	http.HandleFunc("/clients/detail", s.clientHandler.HandleDetail)
	http.HandleFunc("/topology", s.topologyHandler.HandleView)

	// API routes
	http.HandleFunc("/api/clients", s.clientAPI.GetClients)
	http.HandleFunc("/api/clients/detail", s.clientAPI.GetClientDetail)
	http.HandleFunc("/api/topology", s.topoAPI.GetTopology)

	// Static files
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(s.baseDir+"/web/static"))))

	return http.ListenAndServe(addr, nil)
}
