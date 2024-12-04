package middleware

import (
	"net/http"
	"time"
	"github.com/liuscraft/spider-network/pkg/xlog"
)

type ResponseWriter struct {
	http.ResponseWriter
	status int
	size   int
}

func (w *ResponseWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *ResponseWriter) Write(b []byte) (int, error) {
	size, err := w.ResponseWriter.Write(b)
	w.size += size
	return size, err
}

func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &ResponseWriter{ResponseWriter: w, status: http.StatusOK}
		
		next.ServeHTTP(rw, r)
		
		xlog.Infof("%s %s %d %d %v", 
			r.Method, 
			r.URL.Path, 
			rw.status, 
			rw.size,
			time.Since(start),
		)
	})
} 