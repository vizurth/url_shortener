package transport

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"

	"github.com/vizurth/url_shortener/internal/service"
	"github.com/vizurth/url_shortener/internal/storage"
	"github.com/vizurth/url_shortener/pkg/logger"
	"go.uber.org/zap"
)

type Handler struct {
	svc          service.URLService
	shortURLBase string
}

func NewHandler(svc service.URLService, shortURLBase string) *Handler {
	return &Handler{svc: svc, shortURLBase: shortURLBase}
}

func (h *Handler) Shorten(w http.ResponseWriter, r *http.Request) {
	log := logger.From(r.Context())

	var req shortenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.URL == "" {
		http.Error(w, "url is required", http.StatusBadRequest)
		return
	}

	shortCode, isNew, err := h.svc.ShortenURL(r.Context(), req.URL)
	if err != nil {
		if errors.Is(err, service.ErrInvalidURL) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		log.Error(r.Context(), "shorten failed", zap.Error(err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	base, err := url.Parse(h.shortURLBase)
	if err != nil {
		log.Error(r.Context(), "invalid short_url_base", zap.Error(err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	base.Path = "/" + shortCode

	resp := shortenResponse{
		ShortURL:  base.String(),
		ShortCode: shortCode,
	}

	status := http.StatusCreated
	if !isNew {
		status = http.StatusOK
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Error(r.Context(), "encode response failed", zap.Error(err))
	}
}

func (h *Handler) Resolve(w http.ResponseWriter, r *http.Request) {
	log := logger.From(r.Context())

	code := r.PathValue("code")
	if code == "" {
		http.Error(w, "missing code", http.StatusBadRequest)
		return
	}

	originalURL, err := h.svc.Resolve(r.Context(), code)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		log.Error(r.Context(), "resolve failed", zap.Error(err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, originalURL, http.StatusFound)
}
