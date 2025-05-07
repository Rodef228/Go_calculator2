package handler

import (
	"net/http"
	"time"

	"calculator/pkg/logger"

	"github.com/go-chi/chi/v5/middleware"
)

func LoggerMiddleware(log logger.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			defer func() {
				log.Infow("request completed",
					"method", r.Method,
					"path", r.URL.Path,
					"status", ww.Status(),
					"duration", time.Since(start),
					"bytes", ww.BytesWritten(),
				)
			}()

			next.ServeHTTP(ww, r)
		})
	}
}
