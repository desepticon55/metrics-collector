package middleware

import (
	"compress/gzip"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
	"io"
	"net"
	"net/http"
	"strings"
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
					zap.String("request-content-encoding", r.Header.Get("Content-Encoding")),
					zap.String("request-accept-encoding", r.Header.Get("Accept-Encoding")),
					zap.String("request-ip-address", r.Header.Get("X-Real-IP")),
					zap.Int("status", responseWriter.Status()),
					zap.Int("size", responseWriter.BytesWritten()),
					zap.String("response-content-type", responseWriter.Header().Get("Content-Type")),
					zap.String("response-content-encoding", responseWriter.Header().Get("Content-Encoding")),
					zap.Duration("duration", duration),
					zap.String("remote_addr", r.RemoteAddr),
					zap.String("request_id", middleware.GetReqID(r.Context())),
				)
			}()

			next.ServeHTTP(responseWriter, r)
		})
	}
}

func DecompressingMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			if strings.Contains(request.Header.Get("Content-Encoding"), "gzip") {
				reader, err := gzip.NewReader(request.Body)
				if err != nil {
					http.Error(writer, "Error during create gzip reader", http.StatusInternalServerError)
					return
				}
				defer reader.Close()
				request.Body = reader
			}
			next.ServeHTTP(writer, request)
		})
	}
}

func CompressingMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			acceptEncoding := request.Header.Get("Accept-Encoding")
			if strings.Contains(acceptEncoding, "gzip") {
				writer.Header().Set("Content-Encoding", "gzip")
				gzipWriter := gzip.NewWriter(writer)
				defer gzipWriter.Close()
				gzipResponseWriter := &gzipResponseWriter{gzipWriter, writer}
				next.ServeHTTP(gzipResponseWriter, request)
			} else {
				next.ServeHTTP(writer, request)
			}
		})
	}
}

type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

func (w *gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func TrustedSubnetMiddleware(trustedSubnet string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if len(trustedSubnet) != 0 {
				agentIP := r.Header.Get("X-Real-IP")
				if agentIP == "" {
					http.Error(w, "X-Real-IP header missing", http.StatusForbidden)
					return
				}

				if !isIPInTrustedSubnet(agentIP, trustedSubnet) {
					http.Error(w, "Forbidden: IP not in trusted subnet", http.StatusForbidden)
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

func isIPInTrustedSubnet(ipStr, subnetStr string) bool {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}

	_, trustedNet, err := net.ParseCIDR(subnetStr)
	if err != nil {
		return false
	}

	return trustedNet.Contains(ip)
}
