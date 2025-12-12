package audit

import "context"

type contextKey string

const requestContextKey contextKey = "auditRequestContext"

// RequestContext holds HTTP request information for audit logging
type RequestContext struct {
	IPAddress string
	UserAgent string
	TraceID   string
}

// WithRequestContext adds request context to the context for audit logging
func WithRequestContext(ctx context.Context, reqCtx *RequestContext) context.Context {
	return context.WithValue(ctx, requestContextKey, reqCtx)
}

// GetRequestContext extracts request context from the context
func GetRequestContext(ctx context.Context) *RequestContext {
	if reqCtx, ok := ctx.Value(requestContextKey).(*RequestContext); ok {
		return reqCtx
	}
	return nil
}
