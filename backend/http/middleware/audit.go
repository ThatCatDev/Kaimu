package middleware

import (
	"net/http"
	"strings"

	"github.com/thatcatdev/kaimu/backend/internal/services/audit"
	"go.opentelemetry.io/otel/trace"
)

// AuditContextMiddleware extracts request context for audit logging
func AuditContextMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			// Extract IP address
			ipAddress := extractIPAddress(r)

			// Extract user agent
			userAgent := r.UserAgent()

			// Extract trace ID from span context if available
			var traceID string
			if spanCtx := trace.SpanContextFromContext(ctx); spanCtx.HasTraceID() {
				traceID = spanCtx.TraceID().String()
			}

			// Add audit request context
			reqCtx := &audit.RequestContext{
				IPAddress: ipAddress,
				UserAgent: userAgent,
				TraceID:   traceID,
			}
			ctx = audit.WithRequestContext(ctx, reqCtx)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// extractIPAddress extracts the client IP address from the request
func extractIPAddress(r *http.Request) string {
	// Check X-Forwarded-For header (for proxies/load balancers)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one (client IP)
		if idx := strings.Index(xff, ","); idx != -1 {
			return strings.TrimSpace(xff[:idx])
		}
		return strings.TrimSpace(xff)
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}

	// Fall back to RemoteAddr
	// RemoteAddr is in the form "IP:port", extract just the IP
	addr := r.RemoteAddr
	if idx := strings.LastIndex(addr, ":"); idx != -1 {
		return addr[:idx]
	}
	return addr
}
