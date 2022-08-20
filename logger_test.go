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
		w.Write([]byte(returnString))
	})
}

func TestLogger(t *testing.T) {
	buffer := new(bytes.Buffer)

	logger := LoggerWithWriter(buffer, testHandler200("Hello world!"))

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

	logger = LoggerWithWriter(buffer, http.NotFoundHandler())
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
	logger := LoggerWithConfig(loggerConfig, testHandler200("Hello world!"))
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
	logger = LoggerWithWriter(buffer, http.NotFoundHandler())
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

	logger := LoggerWithFormatter(func(param LogFormatterParams) string {
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

	logger := LoggerWithConfig(LoggerConfig{
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
	// test with isTerm flag true.
	p := LogFormatterParams{
		isTerm: true,
	}

	consoleColorMode = autoColor
	assert.Equal(t, true, p.IsOutputColor())

	ForceConsoleColor()
	assert.Equal(t, true, p.IsOutputColor())

	DisableConsoleColor()
	assert.Equal(t, false, p.IsOutputColor())

	// test with isTerm flag false.
	p = LogFormatterParams{
		isTerm: false,
	}

	consoleColorMode = autoColor
	assert.Equal(t, false, p.IsOutputColor())

	ForceConsoleColor()
	assert.Equal(t, true, p.IsOutputColor())

	DisableConsoleColor()
	assert.Equal(t, false, p.IsOutputColor())

	// reset console color mode.
	consoleColorMode = autoColor
}

func TestLoggerWithWriterSkippingPaths(t *testing.T) {
	buffer := new(bytes.Buffer)
	logger := LoggerWithWriter(buffer, testHandler200("I am good!"), "/skipped")

	PerformRequest(logger, "GET", "/logged")
	assert.Contains(t, buffer.String(), "200")

	buffer.Reset()
	PerformRequest(logger, "GET", "/skipped")
	assert.Contains(t, buffer.String(), "")
}

func TestLoggerWithConfigSkippingPaths(t *testing.T) {
	buffer := new(bytes.Buffer)
	logger := LoggerWithConfig(LoggerConfig{
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

func TestDisableConsoleColor(t *testing.T) {
	assert.Equal(t, autoColor, consoleColorMode)
	DisableConsoleColor()
	assert.Equal(t, disableColor, consoleColorMode)

	// reset console color mode.
	consoleColorMode = autoColor
}

func TestForceConsoleColor(t *testing.T) {
	assert.Equal(t, autoColor, consoleColorMode)
	ForceConsoleColor()
	assert.Equal(t, forceColor, consoleColorMode)

	// reset console color mode.
	consoleColorMode = autoColor
}
