package middleware

import (
	"context"
	"net/http"
	"time"

	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	log "github.com/jorge-j1m/hackspark_server/internal/infrastructure/logger"
	"github.com/rs/zerolog"
	"go.jetify.com/typeid/v2"
)

// RequestID is a middleware that injects a request ID into the context of each request
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-Id")
		if requestID == "" {
			requestID = typeid.MustGenerate("rid").String()
		}

		ctx := context.WithValue(r.Context(), log.RequestIDCtxKey, requestID)
		w.Header().Set("X-Request-Id", requestID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Logger is a middleware that logs request details
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ctx := r.Context()

		// Create a custom response writer to capture the status code
		ww := chiMiddleware.NewWrapResponseWriter(w, r.ProtoMajor)

		// Call the next handler
		next.ServeHTTP(ww, r)

		// Calculate request duration
		duration := time.Since(start)

		var l *zerolog.Event
		if ww.Status() >= 500 {
			l = log.Error(ctx)
		} else if ww.Status() >= 400 {
			l = log.Warn(ctx)
		} else {
			l = log.Info(ctx)
		}

		l.Str("remote_addr", r.RemoteAddr).
			Str("method", r.Method).
			Str("url", r.URL.String()).
			Int("status", ww.Status()).
			Int("bytes", ww.BytesWritten()).
			Dur("duration", duration)

		if ww.Status() >= 500 {
			l.Msg("Server error")
		} else if ww.Status() >= 400 {
			l.Msg("Client error")
		} else {
			l.Msg("Request completed")
		}
	})
}

// SecurityHeaders adds security headers to the response
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set security headers
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		w.Header().Set("Content-Security-Policy", "default-src 'self'")
		w.Header().Set("Referrer-Policy", "no-referrer-when-downgrade")
		w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
		w.Header().Set("Pragma", "no-cache")

		next.ServeHTTP(w, r)
	})
}
