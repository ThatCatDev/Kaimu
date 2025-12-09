package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"sync"
	"testing"

	"github.com/rs/zerolog"
)

func TestLoggerWithEnvironment(t *testing.T) {
	// Reset global logger for testing
	once.Do(func() {})
	// Reset the once so we can test initialization
	once = sync.Once{}

	// Capture log output
	var buf bytes.Buffer

	// Save original global log and restore after test
	originalGlobalLog := zerolog.GlobalLevel()
	defer func() {
		zerolog.SetGlobalLevel(originalGlobalLog)
	}()

	// Set test writer
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	// Initialize logger with test environment
	Logger(
		WithServerName("test-service"),
		WithVersion("test-version"),
		WithEnvironment("test-environment"),
	)

	// Get logger and write to our test buffer instead
	logger := Get().Output(&buf)

	// Test logging
	logger.Info().Msg("test message")

	// Parse the log output
	var logEntry map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &logEntry)
	if err != nil {
		t.Fatalf("Failed to parse log output: %v", err)
	}

	// Verify all expected fields are present
	expectedFields := map[string]string{
		"service":     "test-service",
		"version":     "test-version",
		"environment": "test-environment",
		"message":     "test message",
	}

	for field, expectedValue := range expectedFields {
		if value, exists := logEntry[field]; !exists {
			t.Errorf("Expected field '%s' not found in log output", field)
		} else if value != expectedValue {
			t.Errorf("Field '%s' = '%v', expected '%s'", field, value, expectedValue)
		}
	}
}

func TestFromCtxWithEnvironment(t *testing.T) {
	// Reset global logger for testing
	once.Do(func() {})
	once = sync.Once{}

	// Capture log output
	var buf bytes.Buffer

	// Initialize logger with test environment
	Logger(
		WithServerName("ctx-test-service"),
		WithVersion("ctx-test-version"),
		WithEnvironment("ctx-test-environment"),
	)

	// Test context logger
	ctx := context.Background()
	logger := FromCtx(ctx).Output(&buf)

	logger.Info().Str("request_id", "test-123").Msg("context log test")

	// Parse the log output
	var logEntry map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &logEntry)
	if err != nil {
		t.Fatalf("Failed to parse log output: %v", err)
	}

	// Verify environment is included in context logger
	if env, exists := logEntry["environment"]; !exists {
		t.Error("Environment field not found in context logger output")
	} else if env != "ctx-test-environment" {
		t.Errorf("Environment = '%v', expected 'ctx-test-environment'", env)
	}

	// Verify additional fields are preserved
	if requestID, exists := logEntry["request_id"]; !exists {
		t.Error("Request ID field not found in context logger output")
	} else if requestID != "test-123" {
		t.Errorf("Request ID = '%v', expected 'test-123'", requestID)
	}
}