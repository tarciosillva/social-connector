package middleware

import (
	"fmt"
	"net/http"
	"social-connector/internal/infra/logger"
	"strings"
)

func LoggingMiddleware(log *logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, "/swagger/") {
				next.ServeHTTP(w, r)
				return
			}

			wrappedWriter := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			log.Info(fmt.Sprintf("Request: %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr))

			next.ServeHTTP(wrappedWriter, r)
		})
	}
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
	body       []byte
}
