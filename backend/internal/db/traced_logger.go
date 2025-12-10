package db

import (
	"context"
	"errors"
	"time"

	"github.com/thatcatdev/kaimu/backend/internal/logger"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// TracedLogger implements GORM's logger interface with trace correlation
type TracedLogger struct {
	SlowThreshold              time.Duration
	SourceField                string
	SkipErrRecordNotFoundError bool
}

// NewTracedLogger creates a new traced logger for GORM
func NewTracedLogger() *TracedLogger {
	return &TracedLogger{
		SlowThreshold:              200 * time.Millisecond,
		SourceField:                "",
		SkipErrRecordNotFoundError: true,
	}
}

// LogMode implements gorm logger interface
func (l *TracedLogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	return l
}

// Info logs info messages with trace context
func (l *TracedLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	log := logger.FromCtx(ctx)
	log.Info().Msgf(msg, data...)
}

// Warn logs warning messages with trace context
func (l *TracedLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	log := logger.FromCtx(ctx)
	log.Warn().Msgf(msg, data...)
}

// Error logs error messages with trace context
func (l *TracedLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	log := logger.FromCtx(ctx)
	log.Error().Msgf(msg, data...)
}

// Trace logs SQL queries with trace context and performance metrics
func (l *TracedLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	elapsed := time.Since(begin)
	sql, rows := fc()

	log := logger.FromCtx(ctx)

	// Extract trace information if available
	span := trace.SpanFromContext(ctx)
	var traceID, spanID string
	if span.SpanContext().IsValid() {
		traceID = span.SpanContext().TraceID().String()
		spanID = span.SpanContext().SpanID().String()
	}

	// Create base log event
	logEvent := log.Info().
		Str("sql", sql).
		Int64("rows", rows).
		Dur("elapsed", elapsed)

	// Add trace context if available
	if traceID != "" {
		logEvent = logEvent.
			Str("trace_id", traceID).
			Str("span_id", spanID)
	}

	switch {
	case err != nil && (!errors.Is(err, gorm.ErrRecordNotFound) || !l.SkipErrRecordNotFoundError):
		log.Error().
			Err(err).
			Str("sql", sql).
			Int64("rows", rows).
			Dur("elapsed", elapsed).
			Str("trace_id", traceID).
			Str("span_id", spanID).
			Msg("SQL error")
	case elapsed > l.SlowThreshold && l.SlowThreshold != 0:
		log.Warn().
			Str("sql", sql).
			Int64("rows", rows).
			Dur("elapsed", elapsed).
			Str("trace_id", traceID).
			Str("span_id", spanID).
			Msgf("SLOW SQL >= %v", l.SlowThreshold)
	default:
		logEvent.Msg("SQL query")
	}
}
