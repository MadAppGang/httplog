package httplog

// Copyright 2022 Jack Rudenko. MadAppGang. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

type header struct {
	Key   string
	Value string
}

// PerformRequest for testing http logger.
func PerformRequest(r http.Handler, method, path string, headers ...header) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, nil)
	for _, h := range headers {
		req.Header.Add(h.Key, h.Value)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// PerformRequest for testing http logger.
func PerformRequestWithRequest(h http.Handler, r *http.Request) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}

// testHandler200 returns sample data with 200 status code
func testHandler200(returnString string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(returnString))
	})
}

// testHandlerHeaders returns sample data with 200 status code
// and additional headers
func testHandlerHeaders(returnString string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Add("Bearer", "Token qwertyuupofkfmnvjdfkglkgjfhnfjgjkgklgfkjfjf")
		w.Header().Add("Secret2", "протестуємо юнікод літери та такст")
		w.Header().Add("Secret2", "Слава Україні")
		w.Header().Add("Secret2", "test multiple values for header")
		w.Header().Add("Something", "whatever you want")
		w.Write([]byte(returnString))
	})
}

func TestLogger(t *testing.T) {
	buffer := new(bytes.Buffer)

	logger, _ := HandlerWithWriter(buffer, testHandler200("Hello world!"))

	PerformRequest(logger, "GET", "/example?a=100")
	assert.Contains(t, buffer.String(), "200")
	assert.Contains(t, buffer.String(), "GET")
	assert.Contains(t, buffer.String(), "/example")
	assert.Contains(t, buffer.String(), "a=100")

	// I wrote these first (extending the above) but then realized they are more
	// like integration tests because they test the whole logging process rather
	// than individual functions.  Im not sure where these should go.
	buffer.Reset()
	PerformRequest(logger, "POST", "/example")
	assert.Contains(t, buffer.String(), "200")
	assert.Contains(t, buffer.String(), "POST")
	assert.Contains(t, buffer.String(), "/example")

	buffer.Reset()
	PerformRequest(logger, "PUT", "/example")
	assert.Contains(t, buffer.String(), "200")
	assert.Contains(t, buffer.String(), "PUT")
	assert.Contains(t, buffer.String(), "/example")

	buffer.Reset()
	PerformRequest(logger, "DELETE", "/example")
	assert.Contains(t, buffer.String(), "200")
	assert.Contains(t, buffer.String(), "DELETE")
	assert.Contains(t, buffer.String(), "/example")

	buffer.Reset()
	PerformRequest(logger, "PATCH", "/example")
	assert.Contains(t, buffer.String(), "200")
	assert.Contains(t, buffer.String(), "PATCH")
	assert.Contains(t, buffer.String(), "/example")

	buffer.Reset()
	PerformRequest(logger, "HEAD", "/example")
	assert.Contains(t, buffer.String(), "200")
	assert.Contains(t, buffer.String(), "HEAD")
	assert.Contains(t, buffer.String(), "/example")

	buffer.Reset()
	PerformRequest(logger, "OPTIONS", "/example")
	assert.Contains(t, buffer.String(), "200")
	assert.Contains(t, buffer.String(), "OPTIONS")
	assert.Contains(t, buffer.String(), "/example")

	buffer.Reset()

	logger, _ = HandlerWithWriter(buffer, http.NotFoundHandler())
	PerformRequest(logger, "GET", "/notfound")
	assert.Contains(t, buffer.String(), "404")
	assert.Contains(t, buffer.String(), "GET")
	assert.Contains(t, buffer.String(), "/notfound")
}

func TestLoggerWithConfig(t *testing.T) {
	buffer := new(bytes.Buffer)

	loggerConfig := LoggerConfig{
		Output: buffer,
	}
	logger, _ := HandlerWithConfig(loggerConfig, testHandler200("Hello world!"))
	PerformRequest(logger, "GET", "/example?a=100")
	assert.Contains(t, buffer.String(), "200")
	assert.Contains(t, buffer.String(), "GET")
	assert.Contains(t, buffer.String(), "/example")
	assert.Contains(t, buffer.String(), "a=100")

	buffer.Reset()
	PerformRequest(logger, "POST", "/example")
	assert.Contains(t, buffer.String(), "200")
	assert.Contains(t, buffer.String(), "POST")
	assert.Contains(t, buffer.String(), "/example")

	buffer.Reset()
	PerformRequest(logger, "PUT", "/example")
	assert.Contains(t, buffer.String(), "200")
	assert.Contains(t, buffer.String(), "PUT")
	assert.Contains(t, buffer.String(), "/example")

	buffer.Reset()
	PerformRequest(logger, "DELETE", "/example")
	assert.Contains(t, buffer.String(), "200")
	assert.Contains(t, buffer.String(), "DELETE")
	assert.Contains(t, buffer.String(), "/example")

	buffer.Reset()
	PerformRequest(logger, "PATCH", "/example")
	assert.Contains(t, buffer.String(), "200")
	assert.Contains(t, buffer.String(), "PATCH")
	assert.Contains(t, buffer.String(), "/example")

	buffer.Reset()
	PerformRequest(logger, "HEAD", "/example")
	assert.Contains(t, buffer.String(), "200")
	assert.Contains(t, buffer.String(), "HEAD")
	assert.Contains(t, buffer.String(), "/example")

	buffer.Reset()
	PerformRequest(logger, "OPTIONS", "/example")
	assert.Contains(t, buffer.String(), "200")
	assert.Contains(t, buffer.String(), "OPTIONS")
	assert.Contains(t, buffer.String(), "/example")

	buffer.Reset()
	logger, _ = HandlerWithWriter(buffer, http.NotFoundHandler())
	PerformRequest(logger, "GET", "/notfound")
	assert.Contains(t, buffer.String(), "404")
	assert.Contains(t, buffer.String(), "GET")
	assert.Contains(t, buffer.String(), "/notfound")
}

func TestLoggerWithFormatter(t *testing.T) {
	buffer := new(bytes.Buffer)

	d := DefaultWriter
	DefaultWriter = buffer
	defer func() {
		DefaultWriter = d
	}()

	logger, _ := HandlerWithFormatter(func(param LogFormatterParams) string {
		return fmt.Sprintf("[%s] %v | %3d | %13v | %15s | %-7s %#v\n",
			"TEST",
			param.TimeStamp.Format("2006/01/02 - 15:04:05"),
			param.StatusCode,
			param.Latency,
			param.ClientIP,
			param.Method,
			param.Path,
		)
	}, testHandler200("Hello world!"))
	PerformRequest(logger, "GET", "/example?a=100")

	// output test
	assert.Contains(t, buffer.String(), "[TEST]")
	assert.Contains(t, buffer.String(), "200")
	assert.Contains(t, buffer.String(), "GET")
	assert.Contains(t, buffer.String(), "/example")
	assert.Contains(t, buffer.String(), "a=100")
}

func TestLoggerWithConfigFormatting(t *testing.T) {
	var gotParam LogFormatterParams
	buffer := new(bytes.Buffer)

	logger, _ := HandlerWithConfig(LoggerConfig{
		Output:     buffer,
		RouterName: "TEST ROUTER",
		Formatter: func(param LogFormatterParams) string {
			// for assert test
			gotParam = param

			return fmt.Sprintf("[%s] %v | %3d | %13v | %15s | %-7s %s\n",
				param.RouterName,
				param.TimeStamp.Format("2006/01/02 - 15:04:05"),
				param.StatusCode,
				param.Latency,
				param.ClientIP,
				param.Method,
				param.Path,
			)
		},
	}, testHandler200("Some test data"))
	req := httptest.NewRequest("GET", "/example?a=100", nil)
	req.Header.Set("X-Forwarded-For", "20.20.20.20")
	PerformRequestWithRequest(logger, req)

	// output test
	assert.Contains(t, buffer.String(), "[TEST ROUTER]")
	assert.Contains(t, buffer.String(), "200")
	assert.Contains(t, buffer.String(), "GET")
	assert.Contains(t, buffer.String(), "/example")
	assert.Contains(t, buffer.String(), "a=100")

	// LogFormatterParams test
	assert.NotNil(t, gotParam.Request)
	assert.NotEmpty(t, gotParam.TimeStamp)
	assert.Equal(t, 200, gotParam.StatusCode)
	assert.NotEmpty(t, gotParam.Latency)
	assert.Equal(t, "20.20.20.20", gotParam.ClientIP)
	assert.Equal(t, "GET", gotParam.Method)
	assert.Equal(t, "/example?a=100", gotParam.Path)
}

func TestColorForMethod(t *testing.T) {
	colorForMethod := func(method string) string {
		p := LogFormatterParams{
			Method: method,
		}
		return p.MethodColor()
	}

	assert.Equal(t, blue, colorForMethod("GET"), "get should be blue")
	assert.Equal(t, cyan, colorForMethod("POST"), "post should be cyan")
	assert.Equal(t, yellow, colorForMethod("PUT"), "put should be yellow")
	assert.Equal(t, red, colorForMethod("DELETE"), "delete should be red")
	assert.Equal(t, green, colorForMethod("PATCH"), "patch should be green")
	assert.Equal(t, magenta, colorForMethod("HEAD"), "head should be magenta")
	assert.Equal(t, white, colorForMethod("OPTIONS"), "options should be white")
	assert.Equal(t, reset, colorForMethod("TRACE"), "trace is not defined and should be the reset color")
}

func TestColorForStatus(t *testing.T) {
	colorForStatus := func(code int) string {
		p := LogFormatterParams{
			StatusCode: code,
		}
		return p.StatusCodeColor()
	}

	assert.Equal(t, green, colorForStatus(http.StatusOK), "2xx should be green")
	assert.Equal(t, white, colorForStatus(http.StatusMovedPermanently), "3xx should be white")
	assert.Equal(t, yellow, colorForStatus(http.StatusNotFound), "4xx should be yellow")
	assert.Equal(t, red, colorForStatus(2), "other things should be red")
}

func TestResetColor(t *testing.T) {
	p := LogFormatterParams{}
	assert.Equal(t, string([]byte{27, 91, 48, 109}), p.ResetColor())
}

func TestIsOutputColor(t *testing.T) {
	// test with ColorForce
	p := LogFormatterParams{
		colorMode: ColorForce,
	}
	assert.Equal(t, true, p.IsOutputColor())

	// test with ColorAuto (should not appear in params - resolved at middleware level)
	// ColorAuto is resolved to ColorForce or ColorDisable in middleware
	p = LogFormatterParams{
		colorMode: ColorAuto,
	}
	assert.Equal(t, false, p.IsOutputColor())

	// test with ColorDisable
	p = LogFormatterParams{
		colorMode: ColorDisable,
	}
	assert.Equal(t, false, p.IsOutputColor())
}

func TestLoggerWithWriterSkippingPaths(t *testing.T) {
	buffer := new(bytes.Buffer)
	logger, _ := HandlerWithWriter(buffer, testHandler200("I am good!"), "/skipped")

	PerformRequest(logger, "GET", "/logged")
	assert.Contains(t, buffer.String(), "200")

	buffer.Reset()
	PerformRequest(logger, "GET", "/skipped")
	assert.Contains(t, buffer.String(), "")
}

func TestLoggerWithConfigSkippingPaths(t *testing.T) {
	buffer := new(bytes.Buffer)
	logger, _ := HandlerWithConfig(LoggerConfig{
		Output: buffer,
		SkipPaths: []string{
			"/skipped",
			"/payments/\\w+",
			"/user/[0-9]+",
		},
	}, testHandler200("I am good!"))

	PerformRequest(logger, "GET", "/logged")
	assert.Contains(t, buffer.String(), "200")

	buffer.Reset()
	PerformRequest(logger, "GET", "/skipped")
	assert.Contains(t, buffer.String(), "")

	buffer.Reset()
	PerformRequest(logger, "GET", "/payments")
	assert.Contains(t, buffer.String(), "200")

	buffer.Reset()
	PerformRequest(logger, "GET", "/payments/")
	assert.Contains(t, buffer.String(), "200")

	buffer.Reset()
	PerformRequest(logger, "GET", "/payments/1")
	assert.Contains(t, buffer.String(), "")

	buffer.Reset()
	PerformRequest(logger, "GET", "/payments/abcd")
	assert.Contains(t, buffer.String(), "")

	buffer.Reset()
	PerformRequest(logger, "GET", "/PaYments/2Uf")
	assert.Contains(t, buffer.String(), "")

	buffer.Reset()
	PerformRequest(logger, "GET", "/PaYments/2Uf()")
	assert.Contains(t, buffer.String(), "200")
}

func TestLoggerWithConfigMaskingHeaders(t *testing.T) {
	buffer := new(bytes.Buffer)
	logger, _ := HandlerWithConfig(LoggerConfig{
		Output: buffer,
		HideHeaderKeys: []string{
			"^Bearer",
			"^Cookie",
			"^Secret\\w*",
		},
		Formatter: FullFormatterWithRequestAndResponseHeadersAndBody,
	}, testHandlerHeaders("I am good!"))

	PerformRequest(logger, "GET", "/logged",
		header{
			Key:   "Bearer",
			Value: "token ABCDSABCDSABCDSABCDSABCDSABCDSABCDSABCDSABCDS",
		},
		header{
			Key:   "Not-Secret",
			Value: "not a secret value",
		},
		header{
			Key:   "Secret_1",
			Value: "Secret1",
		},
		header{
			Key:   "Secret",
			Value: "SingleSecret",
		},
	)
	assert.Contains(t, buffer.String(), "200")
	assert.NotContains(t, buffer.String(), "ABCDSABCDSABCDSABCDSABCDSABCDSABCDSABCDSABCDS")
	assert.NotContains(t, buffer.String(), "Secret1")
	assert.NotContains(t, buffer.String(), "SingleSecret")
	assert.Contains(t, buffer.String(), "Secret_1")
	assert.Contains(t, buffer.String(), "Secret")
	assert.Contains(t, buffer.String(), "Not-Secret")
	assert.Contains(t, buffer.String(), "not a secret value")

	// response header
	assert.Contains(t, buffer.String(), "T**********f")
	assert.Contains(t, buffer.String(), "п**********т")
	assert.Contains(t, buffer.String(), "С**********і")
	assert.Contains(t, buffer.String(), "whatever you want")
}
