package transport

import "net/http"

func NewRouter(h *Handler) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /", h.Shorten)
	mux.HandleFunc("GET /{code}", h.Resolve)
	return LoggingMiddleware(mux)
}
