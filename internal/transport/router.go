package transport

import (
	"net/http"

	"github.com/vizurth/url_shortener/pkg/logger"
)

func NewRouter(h *Handler, log *logger.Logger) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /", h.Shorten)
	mux.HandleFunc("GET /{code}", h.Resolve)
	return LoggingMiddleware(log, mux)
}
