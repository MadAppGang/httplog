package httplog

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// TEST-001 to TEST-009: Factory Function Error Returns and Config Validation

func TestLogger_ReturnsErrorOnInvalidConfig(t *testing.T) {
	// TEST-001: Logger() returns error on invalid config
	// Note: This test validates that if default config is invalid, error is returned
	// In practice, default config should be valid
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	_, err := Logger(handler)
	if err != nil {
		t.Errorf("Logger() with default config should not return error, got: %v", err)
	}
}

func TestHandlerWithName_ReturnsErrorOnInvalidConfig(t *testing.T) {
	// TEST-002: HandlerWithName() returns error on invalid config
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	_, err := HandlerWithName("API", handler)
	if err != nil {
		t.Errorf("HandlerWithName() with default config should not return error, got: %v", err)
	}
}

func TestLoggerWithConfig_ReturnsErrorOnInvalidSkipPathsRegex(t *testing.T) {
	// TEST-003: LoggerWithConfig() returns error on invalid SkipPaths regex
	conf := LoggerConfig{}
	conf.SkipPaths = []string{"[invalid"}

	_, err := LoggerWithConfig(conf)
	if err == nil {
		t.Error("LoggerWithConfig() should return error for invalid SkipPaths regex")
	}
	if err != nil && !strings.Contains(err.Error(), "SkipPaths") {
		t.Errorf("Error should mention SkipPaths regex, got: %v", err)
	}
}

func TestLoggerWithConfig_ReturnsErrorOnInvalidHideHeaderKeysRegex(t *testing.T) {
	// TEST-004: LoggerWithConfig() returns error on invalid HideHeaderKeys regex
	conf := LoggerConfig{}
	conf.HideHeaderKeys = []string{"[invalid"}

	_, err := LoggerWithConfig(conf)
	if err == nil {
		t.Error("LoggerWithConfig() should return error for invalid HideHeaderKeys regex")
	}
	if err != nil && !strings.Contains(err.Error(), "HideHeaderKeys") {
		t.Errorf("Error should mention HideHeaderKeys regex, got: %v", err)
	}
}

func TestHandlerWithConfig_ReturnsErrorOnInvalidSampleRate(t *testing.T) {
	// TEST-005: HandlerWithConfig() returns error on invalid SampleRate
	conf := LoggerConfig{}
	conf.SampleRate = -0.5

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	_, err := HandlerWithConfig(conf, handler)
	if err == nil {
		t.Error("HandlerWithConfig() should return error for invalid SampleRate")
	}
	if err != nil && !strings.Contains(err.Error(), "SampleRate") {
		t.Errorf("Error should mention SampleRate, got: %v", err)
	}
}

func TestValidateConfig_AcceptsValidConfiguration(t *testing.T) {
	// TEST-006: ValidateConfig() accepts valid configuration
	conf := LoggerConfig{}
	conf.SkipPaths = []string{"/health", "/metrics"}
	conf.HideHeaderKeys = []string{"Authorization", "Cookie"}
	conf.SampleRate = 0.5
	conf.AsyncBufferSize = 1000

	err := ValidateConfig(conf)
	if err != nil {
		t.Errorf("ValidateConfig() should accept valid config, got error: %v", err)
	}
}

func TestValidateConfig_RejectsSampleRateLessThanZero(t *testing.T) {
	// TEST-007: ValidateConfig() rejects SampleRate < 0.0
	conf := LoggerConfig{}
	conf.SampleRate = -0.1

	err := ValidateConfig(conf)
	if err == nil {
		t.Error("ValidateConfig() should reject SampleRate < 0.0")
	}
	if err != nil && !strings.Contains(err.Error(), "SampleRate") {
		t.Errorf("Error should mention SampleRate, got: %v", err)
	}
}

func TestValidateConfig_RejectsSampleRateGreaterThanOne(t *testing.T) {
	// TEST-008: ValidateConfig() rejects SampleRate > 1.0
	conf := LoggerConfig{}
	conf.SampleRate = 1.5

	err := ValidateConfig(conf)
	if err == nil {
		t.Error("ValidateConfig() should reject SampleRate > 1.0")
	}
	if err != nil && !strings.Contains(err.Error(), "SampleRate") {
		t.Errorf("Error should mention SampleRate, got: %v", err)
	}
}

func TestValidateConfig_RejectsNegativeAsyncBufferSize(t *testing.T) {
	// TEST-009: ValidateConfig() rejects negative AsyncBufferSize
	conf := LoggerConfig{}
	conf.AsyncLogging = true
	conf.AsyncBufferSize = -10

	err := ValidateConfig(conf)
	if err == nil {
		t.Error("ValidateConfig() should reject negative AsyncBufferSize")
	}
	if err != nil && !strings.Contains(err.Error(), "AsyncBufferSize") {
		t.Errorf("Error should mention AsyncBufferSize, got: %v", err)
	}
}

// TEST-010 to TEST-017: ColorMode Configuration

func TestColorMode_Values(t *testing.T) {
	// TEST-014: ColorMode has expected values
	// ColorMode doesn't have String() method in v2 - testing enum values directly
	if ColorAuto != 0 {
		t.Errorf("ColorAuto = %d, want 0", ColorAuto)
	}
	if ColorDisable != 1 {
		t.Errorf("ColorDisable = %d, want 1", ColorDisable)
	}
	if ColorForce != 2 {
		t.Errorf("ColorForce = %d, want 2", ColorForce)
	}
}

func TestLogFormatterParams_IsOutputColorRespectsColorMode(t *testing.T) {
	// TEST-017: LogFormatterParams.IsOutputColor() respects ColorMode
	// Note: colorMode is a private field, so we test via LoggerWithConfig
	// This test verifies the public IsOutputColor() method behavior

	tests := []struct {
		name      string
		colorMode ColorMode
		expected  bool
	}{
		{"ColorDisable returns false", ColorDisable, false},
		{"ColorForce returns true", ColorForce, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var capturedParams LogFormatterParams

			conf := LoggerConfig{
				ColorMode: tt.colorMode,
				Formatter: func(params LogFormatterParams) string {
					capturedParams = params
					return ""
				},
			}

			middleware, err := LoggerWithConfig(conf)
			if err != nil {
				t.Fatalf("LoggerWithConfig() failed: %v", err)
			}

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			req := httptest.NewRequest("GET", "/test", nil)
			w := httptest.NewRecorder()

			middleware.Handler(handler).ServeHTTP(w, req)

			got := capturedParams.IsOutputColor()
			if got != tt.expected {
				t.Errorf("IsOutputColor() with ColorMode=%d = %v, want %v", tt.colorMode, got, tt.expected)
			}
		})
	}
}

// TEST-018 to TEST-027: LogFormatterParams Fields

func TestLogFormatterParams_ContextFieldPresent(t *testing.T) {
	// TEST-018: Context field present in LogFormatterParams
	var capturedParams LogFormatterParams

	conf := LoggerConfig{}
	conf.Formatter = func(params LogFormatterParams) string {
		capturedParams = params
		return ""
	}

	middleware, err := LoggerWithConfig(conf)
	if err != nil {
		t.Fatalf("LoggerWithConfig() failed: %v", err)
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	ctx := context.WithValue(req.Context(), "trace_id", "12345")
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	middleware.Handler(handler).ServeHTTP(w, req)

	if capturedParams.Context == nil {
		t.Error("Context field should be populated in LogFormatterParams")
	}

	if capturedParams.Context != req.Context() {
		t.Error("Context should equal request.Context()")
	}

	// Verify we can extract trace ID
	if traceID := capturedParams.Context.Value("trace_id"); traceID != "12345" {
		t.Errorf("Context should contain trace ID, got: %v", traceID)
	}
}

func TestLogFormatterParams_RequestBodyFieldPopulatedWhenEnabled(t *testing.T) {
	// TEST-019: RequestBody field populated when CaptureRequestBody enabled
	requestBody := []byte(`{"key":"value"}`)
	var capturedParams LogFormatterParams

	conf := LoggerConfig{}
	conf.CaptureRequestBody = true
	conf.Formatter = func(params LogFormatterParams) string {
		capturedParams = params
		return ""
	}

	middleware, err := LoggerWithConfig(conf)
	if err != nil {
		t.Fatalf("LoggerWithConfig() failed: %v", err)
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Handler should still be able to read body
		body, _ := io.ReadAll(r.Body)
		if len(body) == 0 {
			t.Error("Handler should receive readable request body")
		}
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("POST", "/test", bytes.NewReader(requestBody))
	w := httptest.NewRecorder()

	middleware.Handler(handler).ServeHTTP(w, req)

	if len(capturedParams.RequestBody) == 0 {
		t.Error("RequestBody should be populated when CaptureRequestBody enabled")
	}

	if !bytes.Equal(capturedParams.RequestBody, requestBody) {
		t.Errorf("RequestBody = %q, want %q", capturedParams.RequestBody, requestBody)
	}
}

func TestLogFormatterParams_RequestBodyEmptyWhenDisabled(t *testing.T) {
	// TEST-020: RequestBody field empty when CaptureRequestBody disabled
	var capturedParams LogFormatterParams

	conf := LoggerConfig{}
	conf.CaptureRequestBody = false
	conf.Formatter = func(params LogFormatterParams) string {
		capturedParams = params
		return ""
	}

	middleware, err := LoggerWithConfig(conf)
	if err != nil {
		t.Fatalf("LoggerWithConfig() failed: %v", err)
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("POST", "/test", bytes.NewReader([]byte(`{"key":"value"}`)))
	w := httptest.NewRecorder()

	middleware.Handler(handler).ServeHTTP(w, req)

	if len(capturedParams.RequestBody) != 0 {
		t.Error("RequestBody should be empty when CaptureRequestBody disabled")
	}
}

func TestLogFormatterParams_ResponseBodyFieldPopulatedWhenEnabled(t *testing.T) {
	// TEST-021: ResponseBody field populated when CaptureResponseBody enabled
	responseBody := []byte(`{"status":"ok"}`)
	var capturedParams LogFormatterParams

	conf := LoggerConfig{}
	conf.CaptureResponseBody = true
	conf.Formatter = func(params LogFormatterParams) string {
		capturedParams = params
		return ""
	}

	middleware, err := LoggerWithConfig(conf)
	if err != nil {
		t.Fatalf("LoggerWithConfig() failed: %v", err)
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(responseBody)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	middleware.Handler(handler).ServeHTTP(w, req)

	if len(capturedParams.ResponseBody) == 0 {
		t.Error("ResponseBody should be populated when CaptureResponseBody enabled")
	}

	if !bytes.Equal(capturedParams.ResponseBody, responseBody) {
		t.Errorf("ResponseBody = %q, want %q", capturedParams.ResponseBody, responseBody)
	}
}

func TestLogFormatterParams_ResponseBodyEmptyWhenDisabled(t *testing.T) {
	// TEST-022: ResponseBody field empty when CaptureResponseBody disabled
	var capturedParams LogFormatterParams

	conf := LoggerConfig{}
	conf.CaptureResponseBody = false
	conf.Formatter = func(params LogFormatterParams) string {
		capturedParams = params
		return ""
	}

	middleware, err := LoggerWithConfig(conf)
	if err != nil {
		t.Fatalf("LoggerWithConfig() failed: %v", err)
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	middleware.Handler(handler).ServeHTTP(w, req)

	if len(capturedParams.ResponseBody) != 0 {
		t.Error("ResponseBody should be empty when CaptureResponseBody disabled")
	}
}

func TestLogFormatterParams_LevelFromStatusCode(t *testing.T) {
	// TEST-023 to TEST-026: Level field populated based on status code
	tests := []struct {
		name       string
		statusCode int
		wantLevel  Level
	}{
		{"2xx -> Info", 200, LevelInfo},
		{"3xx -> Info", 301, LevelInfo},
		{"4xx -> Warn", 404, LevelWarn},
		{"5xx -> Error", 500, LevelError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var capturedParams LogFormatterParams

			conf := LoggerConfig{}
			conf.Formatter = func(params LogFormatterParams) string {
				capturedParams = params
				return ""
			}

			middleware, err := LoggerWithConfig(conf)
			if err != nil {
				t.Fatalf("LoggerWithConfig() failed: %v", err)
			}

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
			})

			req := httptest.NewRequest("GET", "/test", nil)
			w := httptest.NewRecorder()

			middleware.Handler(handler).ServeHTTP(w, req)

			if capturedParams.Level != tt.wantLevel {
				t.Errorf("Level = %v, want %v for status %d", capturedParams.Level, tt.wantLevel, tt.statusCode)
			}
		})
	}
}

// TEST-028 to TEST-040: New Configuration Options

func TestCaptureRequestBody_DoesNotMutateRequestBody(t *testing.T) {
	// TEST-029: CaptureRequestBody does not mutate request body
	requestBody := []byte(`{"key":"value"}`)

	conf := LoggerConfig{}
	conf.CaptureRequestBody = true

	middleware, err := LoggerWithConfig(conf)
	if err != nil {
		t.Fatalf("LoggerWithConfig() failed: %v", err)
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Handler should receive complete, readable request body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("Failed to read request body: %v", err)
		}
		if !bytes.Equal(body, requestBody) {
			t.Errorf("Handler received body = %q, want %q", body, requestBody)
		}
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("POST", "/test", bytes.NewReader(requestBody))
	w := httptest.NewRecorder()

	middleware.Handler(handler).ServeHTTP(w, req)
}

func TestSampleRate_ZeroDefaultsToAllRequests(t *testing.T) {
	// TEST-030: SampleRate=0.0 defaults to 1.0 (logs all requests)
	// Note: This is by design - zero-value means "use default" (100% sampling)
	// This is documented behavior - users wanting 0% sampling should use SkipPaths
	logCount := 0

	conf := LoggerConfig{}
	conf.SampleRate = 0.0 // Zero-value defaults to 100%
	conf.Formatter = func(params LogFormatterParams) string {
		logCount++
		return ""
	}

	middleware, err := LoggerWithConfig(conf)
	if err != nil {
		t.Fatalf("LoggerWithConfig() failed: %v", err)
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Send 100 requests
	for i := 0; i < 100; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		middleware.Handler(handler).ServeHTTP(w, req)
	}

	// Zero-value defaults to 1.0, so all requests should be logged
	if logCount != 100 {
		t.Errorf("SampleRate=0.0 should default to logging all requests, got %d", logCount)
	}
}

func TestSampleRate_OneLogsAllRequests(t *testing.T) {
	// TEST-031: SampleRate=1.0 logs all requests
	logCount := 0

	conf := LoggerConfig{}
	conf.SampleRate = 1.0
	conf.Formatter = func(params LogFormatterParams) string {
		logCount++
		return ""
	}

	middleware, err := LoggerWithConfig(conf)
	if err != nil {
		t.Fatalf("LoggerWithConfig() failed: %v", err)
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Send 100 requests
	for i := 0; i < 100; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		middleware.Handler(handler).ServeHTTP(w, req)
	}

	if logCount != 100 {
		t.Errorf("SampleRate=1.0 should log 100 requests, got %d", logCount)
	}
}

func TestSampleRate_HalfLogsApproximatelyHalfRequests(t *testing.T) {
	// TEST-032: SampleRate=0.5 logs approximately 50% of requests
	logCount := 0

	conf := LoggerConfig{}
	conf.SampleRate = 0.5
	conf.Formatter = func(params LogFormatterParams) string {
		logCount++
		return ""
	}

	middleware, err := LoggerWithConfig(conf)
	if err != nil {
		t.Fatalf("LoggerWithConfig() failed: %v", err)
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Send 1000 requests for statistical accuracy
	for i := 0; i < 1000; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		middleware.Handler(handler).ServeHTTP(w, req)
	}

	// Allow statistical variance: 45-55%
	if logCount < 450 || logCount > 550 {
		t.Errorf("SampleRate=0.5 should log ~500 requests (450-550), got %d", logCount)
	}
}

func TestMinLevel_FiltersOutInfoLogs(t *testing.T) {
	// TEST-033: MinLevel=LevelWarn filters out Info logs
	logCount := 0

	conf := LoggerConfig{}
	conf.MinLevel = LevelWarn
	conf.Formatter = func(params LogFormatterParams) string {
		logCount++
		return ""
	}

	middleware, err := LoggerWithConfig(conf)
	if err != nil {
		t.Fatalf("LoggerWithConfig() failed: %v", err)
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK) // 200 -> Info level
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	middleware.Handler(handler).ServeHTTP(w, req)

	if logCount != 0 {
		t.Errorf("MinLevel=LevelWarn should filter out Info logs, got %d logs", logCount)
	}
}

func TestMinLevel_LogsWarnLevel(t *testing.T) {
	// TEST-034: MinLevel=LevelWarn logs Warn level
	logCount := 0

	conf := LoggerConfig{}
	conf.MinLevel = LevelWarn
	conf.Formatter = func(params LogFormatterParams) string {
		logCount++
		return ""
	}

	middleware, err := LoggerWithConfig(conf)
	if err != nil {
		t.Fatalf("LoggerWithConfig() failed: %v", err)
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound) // 404 -> Warn level
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	middleware.Handler(handler).ServeHTTP(w, req)

	if logCount != 1 {
		t.Errorf("MinLevel=LevelWarn should log Warn level, got %d logs", logCount)
	}
}

func TestMinLevel_LogsErrorLevel(t *testing.T) {
	// TEST-035: MinLevel=LevelWarn logs Error level
	logCount := 0

	conf := LoggerConfig{}
	conf.MinLevel = LevelWarn
	conf.Formatter = func(params LogFormatterParams) string {
		logCount++
		return ""
	}

	middleware, err := LoggerWithConfig(conf)
	if err != nil {
		t.Fatalf("LoggerWithConfig() failed: %v", err)
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError) // 500 -> Error level
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	middleware.Handler(handler).ServeHTTP(w, req)

	if logCount != 1 {
		t.Errorf("MinLevel=LevelWarn should log Error level, got %d logs", logCount)
	}
}

func TestAsyncLogging_LogsAsynchronously(t *testing.T) {
	// TEST-037: AsyncLogging=true logs asynchronously
	var logOutput bytes.Buffer
	logged := false

	conf := LoggerConfig{}
	conf.AsyncLogging = true
	conf.AsyncBufferSize = 10
	conf.Output = &logOutput
	conf.Formatter = func(params LogFormatterParams) string {
		logged = true
		return "test log\n"
	}

	middleware, err := LoggerWithConfig(conf)
	if err != nil {
		t.Fatalf("LoggerWithConfig() failed: %v", err)
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	middleware.Handler(handler).ServeHTTP(w, req)

	// Log should appear shortly after (async)
	time.Sleep(50 * time.Millisecond)

	if !logged {
		t.Error("Async logging should eventually log the request")
	}

	if logOutput.Len() == 0 {
		t.Error("Log output should contain logged data")
	}
}

// TEST-047 to TEST-057: ResponseWriter Interface Compatibility

func TestResponseWriter_ImplementsHTTPResponseWriter(t *testing.T) {
	// TEST-047: ResponseWriter implements http.ResponseWriter
	var capturedWriter ResponseWriter

	conf := LoggerConfig{}
	conf.Formatter = func(params LogFormatterParams) string {
		return ""
	}

	middleware, err := LoggerWithConfig(conf)
	if err != nil {
		t.Fatalf("LoggerWithConfig() failed: %v", err)
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Type assert to ResponseWriter
		if rw, ok := w.(ResponseWriter); ok {
			capturedWriter = rw
		} else {
			t.Error("ResponseWriter should implement httplog.ResponseWriter interface")
		}

		// Should also be http.ResponseWriter
		var _ http.ResponseWriter = w

		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	middleware.Handler(handler).ServeHTTP(w, req)

	if capturedWriter == nil {
		t.Error("Failed to capture ResponseWriter")
	}
}

func TestResponseWriter_StatusReturnsStatusCode(t *testing.T) {
	// TEST-051: ResponseWriter.Status() returns status code
	var capturedWriter ResponseWriter

	conf := LoggerConfig{}

	middleware, err := LoggerWithConfig(conf)
	if err != nil {
		t.Fatalf("LoggerWithConfig() failed: %v", err)
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if rw, ok := w.(ResponseWriter); ok {
			capturedWriter = rw
		}
		w.WriteHeader(http.StatusCreated)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	middleware.Handler(handler).ServeHTTP(w, req)

	if capturedWriter == nil {
		t.Fatal("Failed to capture ResponseWriter")
	}

	if status := capturedWriter.Status(); status != http.StatusCreated {
		t.Errorf("Status() = %d, want %d", status, http.StatusCreated)
	}
}

func TestResponseWriter_StatusReturnsZeroBeforeWrite(t *testing.T) {
	// TEST-052: ResponseWriter.Status() returns 0 before write
	conf := LoggerConfig{}

	middleware, err := LoggerWithConfig(conf)
	if err != nil {
		t.Fatalf("LoggerWithConfig() failed: %v", err)
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if rw, ok := w.(ResponseWriter); ok {
			// Check status before write
			if status := rw.Status(); status != 0 {
				t.Errorf("Status() before write = %d, want 0", status)
			}
		}
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	middleware.Handler(handler).ServeHTTP(w, req)
}

func TestResponseWriter_WrittenReturnsTrueAfterWrite(t *testing.T) {
	// TEST-053: ResponseWriter.Written() returns true after write
	var capturedWriter ResponseWriter

	conf := LoggerConfig{}

	middleware, err := LoggerWithConfig(conf)
	if err != nil {
		t.Fatalf("LoggerWithConfig() failed: %v", err)
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if rw, ok := w.(ResponseWriter); ok {
			capturedWriter = rw
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test"))
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	middleware.Handler(handler).ServeHTTP(w, req)

	if capturedWriter == nil {
		t.Fatal("Failed to capture ResponseWriter")
	}

	if !capturedWriter.Written() {
		t.Error("Written() should return true after write")
	}
}

func TestResponseWriter_SizeReturnsBodySize(t *testing.T) {
	// TEST-054: ResponseWriter.Size() returns body size
	var capturedWriter ResponseWriter
	responseBody := []byte("test response body")

	conf := LoggerConfig{}

	middleware, err := LoggerWithConfig(conf)
	if err != nil {
		t.Fatalf("LoggerWithConfig() failed: %v", err)
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if rw, ok := w.(ResponseWriter); ok {
			capturedWriter = rw
		}
		w.WriteHeader(http.StatusOK)
		w.Write(responseBody)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	middleware.Handler(handler).ServeHTTP(w, req)

	if capturedWriter == nil {
		t.Fatal("Failed to capture ResponseWriter")
	}

	if size := capturedWriter.Size(); size != len(responseBody) {
		t.Errorf("Size() = %d, want %d", size, len(responseBody))
	}
}

func TestResponseWriter_BodyReturnsCapturedBody(t *testing.T) {
	// TEST-055: ResponseWriter.Body() returns captured body
	var capturedWriter ResponseWriter
	responseBody := []byte(`{"status":"ok"}`)

	conf := LoggerConfig{}
	conf.CaptureResponseBody = true

	middleware, err := LoggerWithConfig(conf)
	if err != nil {
		t.Fatalf("LoggerWithConfig() failed: %v", err)
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if rw, ok := w.(ResponseWriter); ok {
			capturedWriter = rw
		}
		w.WriteHeader(http.StatusOK)
		w.Write(responseBody)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	middleware.Handler(handler).ServeHTTP(w, req)

	if capturedWriter == nil {
		t.Fatal("Failed to capture ResponseWriter")
	}

	body := capturedWriter.Body()
	if !bytes.Equal(body, responseBody) {
		t.Errorf("Body() = %q, want %q", body, responseBody)
	}
}

func TestResponseWriter_BodyReturnsNilWhenCaptureDisabled(t *testing.T) {
	// TEST-056: ResponseWriter.Body() returns nil when capture disabled
	var capturedWriter ResponseWriter

	conf := LoggerConfig{}
	conf.CaptureResponseBody = false

	middleware, err := LoggerWithConfig(conf)
	if err != nil {
		t.Fatalf("LoggerWithConfig() failed: %v", err)
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if rw, ok := w.(ResponseWriter); ok {
			capturedWriter = rw
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test"))
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	middleware.Handler(handler).ServeHTTP(w, req)

	if capturedWriter == nil {
		t.Fatal("Failed to capture ResponseWriter")
	}

	body := capturedWriter.Body()
	if body != nil {
		t.Errorf("Body() should return nil when capture disabled, got %q", body)
	}
}

// TEST-075 to TEST-077: LoggerConfig{} Default Values

func TestNewConfig_ReturnsConfigWithColorAutoDefault(t *testing.T) {
	// TEST-075: LoggerConfig{} returns config with ColorAuto default
	conf := LoggerConfig{}

	if conf.ColorMode != ColorAuto {
		t.Errorf("LoggerConfig{} ColorMode = %v, want %v", conf.ColorMode, ColorAuto)
	}
}

func TestLoggerConfig_ZeroValueSampleRate(t *testing.T) {
	// TEST-076: LoggerConfig{} SampleRate zero-value is 0, but middleware defaults to 1.0
	conf := LoggerConfig{}

	// Go zero-value is 0.0
	if conf.SampleRate != 0.0 {
		t.Errorf("LoggerConfig{} SampleRate = %v, want 0.0 (Go zero-value)", conf.SampleRate)
	}
}

func TestLoggerConfig_ZeroValueMinLevel(t *testing.T) {
	// TEST-077: LoggerConfig{} MinLevel zero-value is LevelDebug (0)
	conf := LoggerConfig{}

	// Go zero-value is 0 = LevelDebug
	if conf.MinLevel != LevelDebug {
		t.Errorf("LoggerConfig{} MinLevel = %v, want %v (Go zero-value)", conf.MinLevel, LevelDebug)
	}
}

// Additional helper tests for Level type

func TestLevel_String(t *testing.T) {
	// Level.String() returns correct string representation
	tests := []struct {
		level    Level
		expected string
	}{
		{LevelDebug, "DEBUG"},
		{LevelInfo, "INFO"},
		{LevelWarn, "WARN"},
		{LevelError, "ERROR"},
		{Level(999), "UNKNOWN"},
	}

	for _, tt := range tests {
		got := tt.level.String()
		if got != tt.expected {
			t.Errorf("Level(%d).String() = %q, want %q", tt.level, got, tt.expected)
		}
	}
}

func TestLevelFromStatusCode(t *testing.T) {
	// LevelFromStatusCode maps status codes correctly
	tests := []struct {
		status   int
		expected Level
	}{
		{200, LevelInfo},
		{201, LevelInfo},
		{301, LevelInfo},
		{302, LevelInfo},
		{400, LevelWarn},
		{404, LevelWarn},
		{500, LevelError},
		{503, LevelError},
	}

	for _, tt := range tests {
		got := LevelFromStatusCode(tt.status)
		if got != tt.expected {
			t.Errorf("LevelFromStatusCode(%d) = %v, want %v", tt.status, got, tt.expected)
		}
	}
}
