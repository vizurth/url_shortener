package transport

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/vizurth/url_shortener/pkg/logger"
	"go.uber.org/zap"
)

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := uuid.New().String()
		ctx := logger.WithRequestID(r.Context(), requestID)
		r = r.WithContext(ctx)

		log := logger.From(ctx)
		start := time.Now()

		rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}

		log.Info(ctx, "incoming request",
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
		)

		next.ServeHTTP(rw, r)

		log.Info(ctx, "request completed",
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.Int("status", rw.status),
			zap.Duration("duration", time.Since(start)),
		)
	})
}
