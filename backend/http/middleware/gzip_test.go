package middleware

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGzipMiddleware(t *testing.T) {
	// Create a test handler that returns a simple response
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, World! This is a test response that should be compressed."))
	})

	// Wrap with gzip middleware
	handler := GzipMiddleware()(testHandler)

	t.Run("Response compression with Accept-Encoding: gzip", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Accept-Encoding", "gzip")

		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, req)

		// Check that response is compressed
		if recorder.Header().Get("Content-Encoding") != "gzip" {
			t.Errorf("Expected Content-Encoding: gzip, got: %s", recorder.Header().Get("Content-Encoding"))
		}

		// Decompress and verify content
		gzipReader, err := gzip.NewReader(recorder.Body)
		if err != nil {
			t.Fatalf("Failed to create gzip reader: %v", err)
		}
		defer gzipReader.Close()

		decompressed, err := io.ReadAll(gzipReader)
		if err != nil {
			t.Fatalf("Failed to decompress response: %v", err)
		}

		expected := "Hello, World! This is a test response that should be compressed."
		if string(decompressed) != expected {
			t.Errorf("Expected: %s, got: %s", expected, string(decompressed))
		}
	})

	t.Run("No compression without Accept-Encoding: gzip", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)

		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, req)

		// Check that response is not compressed
		if recorder.Header().Get("Content-Encoding") == "gzip" {
			t.Error("Response should not be compressed without Accept-Encoding: gzip")
		}

		expected := "Hello, World! This is a test response that should be compressed."
		if recorder.Body.String() != expected {
			t.Errorf("Expected: %s, got: %s", expected, recorder.Body.String())
		}
	})

	t.Run("Request decompression", func(t *testing.T) {
		// Create compressed request body
		var buf bytes.Buffer
		gzipWriter := gzip.NewWriter(&buf)
		gzipWriter.Write([]byte("compressed request body"))
		gzipWriter.Close()

		// Create handler that reads request body
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Write(body)
		})

		handler := GzipMiddleware()(testHandler)

		req := httptest.NewRequest("POST", "/test", &buf)
		req.Header.Set("Content-Encoding", "gzip")
		req.Header.Set("Accept-Encoding", "gzip")

		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, req)

		// The response should be the decompressed request body, then compressed again
		if recorder.Header().Get("Content-Encoding") != "gzip" {
			t.Errorf("Expected Content-Encoding: gzip, got: %s", recorder.Header().Get("Content-Encoding"))
		}

		// Decompress response to verify it contains the original request body
		gzipReader, err := gzip.NewReader(recorder.Body)
		if err != nil {
			t.Fatalf("Failed to create gzip reader: %v", err)
		}
		defer gzipReader.Close()

		decompressed, err := io.ReadAll(gzipReader)
		if err != nil {
			t.Fatalf("Failed to decompress response: %v", err)
		}

		expected := "compressed request body"
		if string(decompressed) != expected {
			t.Errorf("Expected: %s, got: %s", expected, string(decompressed))
		}
	})

	t.Run("Invalid gzip request", func(t *testing.T) {
		// Create invalid gzip data
		invalidGzipData := strings.NewReader("this is not valid gzip data")

		req := httptest.NewRequest("POST", "/test", invalidGzipData)
		req.Header.Set("Content-Encoding", "gzip")

		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, req)

		// Should return 400 Bad Request
		if recorder.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got: %d", recorder.Code)
		}
	})

	t.Run("Metrics endpoint should not be compressed", func(t *testing.T) {
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("# HELP some_metric A test metric\n# TYPE some_metric counter\nsome_metric 42\n"))
		})

		handler := GzipMiddleware()(testHandler)

		req := httptest.NewRequest("GET", "/metrics", nil)
		req.Header.Set("Accept-Encoding", "gzip")

		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, req)

		// Metrics endpoint should NOT be compressed even with Accept-Encoding: gzip
		if recorder.Header().Get("Content-Encoding") == "gzip" {
			t.Error("/metrics endpoint should not be compressed")
		}

		expected := "# HELP some_metric A test metric\n# TYPE some_metric counter\nsome_metric 42\n"
		if recorder.Body.String() != expected {
			t.Errorf("Expected: %s, got: %s", expected, recorder.Body.String())
		}
	})

	t.Run("Healthcheck endpoint should not be compressed", func(t *testing.T) {
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`{"status":"healthy"}`))
		})

		handler := GzipMiddleware()(testHandler)

		req := httptest.NewRequest("GET", "/healthcheck", nil)
		req.Header.Set("Accept-Encoding", "gzip")

		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, req)

		// Healthcheck endpoint should NOT be compressed
		if recorder.Header().Get("Content-Encoding") == "gzip" {
			t.Error("/healthcheck endpoint should not be compressed")
		}

		expected := `{"status":"healthy"}`
		if recorder.Body.String() != expected {
			t.Errorf("Expected: %s, got: %s", expected, recorder.Body.String())
		}
	})

	t.Run("GraphQL endpoint should still be compressed", func(t *testing.T) {
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`{"data":{"anime":[{"title":"Test Anime","episodes":12}]}}`))
		})

		handler := GzipMiddleware()(testHandler)

		req := httptest.NewRequest("POST", "/graphql", nil)
		req.Header.Set("Accept-Encoding", "gzip")

		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, req)

		// GraphQL endpoint should still be compressed
		if recorder.Header().Get("Content-Encoding") != "gzip" {
			t.Error("/graphql endpoint should be compressed")
		}
	})
}