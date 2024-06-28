package middleware

import (
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
	"net/http"
	"time"
)

func LoggingMiddleware(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			responseWriter := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			defer func() {
				duration := time.Since(start)
				logger.Info("HTTP request",
					zap.String("method", r.Method),
					zap.String("url", r.URL.String()),
					zap.String("request-content-type", r.Header.Get("Content-Type")),
					zap.Int("status", responseWriter.Status()),
					zap.Int("size", responseWriter.BytesWritten()),
					zap.String("response-content-type", responseWriter.Header().Get("Content-Type")),
					zap.Duration("duration", duration),
					zap.String("remote_addr", r.RemoteAddr),
					zap.String("request_id", middleware.GetReqID(r.Context())),
				)
			}()

			next.ServeHTTP(responseWriter, r)
		})
	}
}
