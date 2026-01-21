package httplog

// Copyright 2022 Jack Rudenko. MadAppGang. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"time"
)

const (
	green   = "\033[97;42m"
	white   = "\033[90;47m"
	yellow  = "\033[90;43m"
	red     = "\033[97;41m"
	blue    = "\033[97;44m"
	magenta = "\033[97;45m"
	cyan    = "\033[97;46m"
	reset   = "\033[0m"
)

// ColorMode controls color output behavior
type ColorMode int

const (
	// ColorAuto detects if output is a terminal
	ColorAuto ColorMode = iota

	// ColorDisable forces colors off
	ColorDisable

	// ColorForce forces colors on (even for non-terminals)
	ColorForce
)

// Level represents log severity level
type Level int

const (
	// LevelDebug is for debug-level logs (not used by default)
	LevelDebug Level = iota

	// LevelInfo is for informational logs (2xx, 3xx responses)
	LevelInfo

	// LevelWarn is for warning logs (4xx responses)
	LevelWarn

	// LevelError is for error logs (5xx responses)
	LevelError
)

// LevelFromStatusCode determines log level from HTTP status code
func LevelFromStatusCode(status int) Level {
	switch {
	case status >= 500:
		return LevelError
	case status >= 400:
		return LevelWarn
	default:
		return LevelInfo
	}
}

// String returns the string representation of the level
func (l Level) String() string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// LoggerConfig defines the config for Logger middleware.
type LoggerConfig struct {
	// ColorMode controls color output behavior
	// Default: ColorAuto (detect terminal)
	ColorMode ColorMode

	// Optional. Default value is httplog.DefaultLogFormatter
	Formatter LogFormatter

	// Output is a writer where logs are written.
	// Optional. Default value is httplog.DefaultWriter.
	Output io.Writer

	// SkipPaths is an url path array which logs are not written.
	// Could be a regexp like: /user/payment/*
	// Optional.
	SkipPaths []string

	// HideHeader is a header keys array which value should be masked with ****.
	// Optional.
	HideHeaderKeys []string

	// ProxyHandler is a instance of Proxy struct with could get remote IP using proxy data
	// Default is default httplog.NewLogger()
	// If you run you instance on Google App engine or Cloudflare,
	// you need to create explicit Proxy instance with httplog.NewLoggerWithType(...)
	ProxyHandler *Proxy

	// Router prints router name in the log.
	// If you have more than one router it is useful to get one's name in a console output.
	RouterName string

	// CaptureResponseBody saves response body copy for debug purposes
	// WARNING: Increases memory usage, use for debugging only
	// Default: false
	CaptureResponseBody bool

	// CaptureRequestBody enables request body capture
	// Captured in middleware (not formatter) to prevent mutation
	// WARNING: Increases memory usage, use for debugging only
	// Default: false
	CaptureRequestBody bool

	// SampleRate controls what percentage of requests to log
	// Range: 0.0 (0% - log nothing) to 1.0 (100% - log all)
	// Use -1 or leave unset to use default (100%)
	// Example: 0.01 = 1%, 0.1 = 10%
	// Default: -1 (100% - log all requests)
	SampleRate float64

	// DeterministicSampling uses hash-based sampling instead of random
	// When true, same request path/method will consistently be sampled or not
	// Useful for reproducible behavior and debugging
	// Default: false (random sampling)
	DeterministicSampling bool

	// AsyncLogging enables asynchronous log writing
	// Reduces request latency, but may lose logs on crash
	// Default: false (synchronous)
	AsyncLogging bool

	// AsyncBufferSize is the channel buffer size for async logging
	// Only used if AsyncLogging is true
	// Default: 1000
	AsyncBufferSize int

	// MinLevel is the minimum log level to output
	// Requests below this level are not logged
	// Level determined by status code:
	//   - 2xx, 3xx: Info
	//   - 4xx: Warn
	//   - 5xx: Error
	// Default: LevelInfo
	MinLevel Level
}

// LogFormatter gives the signature of the formatter function passed to LoggerWithFormatter
// you can use predefined, like httplog.DefaultLogFormatter or httplog.ShortLogFormatter
// or you can create your custom
type LogFormatter func(params LogFormatterParams) string

// ValidateConfig validates LoggerConfig before middleware creation
// Returns detailed error if invalid, nil if valid
func ValidateConfig(conf LoggerConfig) error {
	// Validate SkipPaths regexes
	for i, pattern := range conf.SkipPaths {
		if _, err := regexp.Compile(pattern); err != nil {
			return fmt.Errorf("invalid SkipPaths[%d] regex pattern '%s': %w", i, pattern, err)
		}
	}

	// Validate HideHeaderKeys regexes
	for i, pattern := range conf.HideHeaderKeys {
		if _, err := regexp.Compile(pattern); err != nil {
			return fmt.Errorf("invalid HideHeaderKeys[%d] regex pattern '%s': %w", i, pattern, err)
		}
	}

	// Validate SampleRate
	if conf.SampleRate != -1 && (conf.SampleRate < 0.0 || conf.SampleRate > 1.0) {
		return fmt.Errorf("invalid SampleRate: %f (must be -1 for default, or between 0.0 and 1.0)", conf.SampleRate)
	}

	// Validate AsyncBufferSize - only reject negative values
	if conf.AsyncBufferSize < 0 {
		return fmt.Errorf("invalid AsyncBufferSize: %d (cannot be negative)", conf.AsyncBufferSize)
	}

	return nil
}

// LogFormatterParams is the structure any formatter will be handed when time to log comes
type LogFormatterParams struct {
	Request *http.Request

	// Context from http.Request for trace ID extraction, etc.
	Context context.Context

	// Router prints router name in the log.
	// If you have more than one router it is useful to get one's name in a console output.
	RouterName string
	// TimeStamp shows the time after the server returns a response.
	TimeStamp time.Time
	// StatusCode is HTTP response code.
	StatusCode int
	// Latency is how much time the server cost to process a certain request.
	Latency time.Duration
	// ClientIP calculated real IP of requester, see Proxy for details.
	ClientIP string
	// Method is the HTTP method given to the request.
	Method string
	// Path is a path the client requests.
	Path string
	// colorMode is the color mode for this logger (private)
	colorMode ColorMode
	// BodySize is the size of the Response Body
	BodySize int
	// ResponseBody is the response body content (if captured)
	ResponseBody []byte
	// RequestBody is the request body content (if captured)
	RequestBody []byte
	// Response header
	ResponseHeader http.Header
	// RequestHeader are the request headers (masked if configured)
	RequestHeader http.Header
	// Level is the log level for this request
	Level Level
}

// StatusCodeColor is the ANSI color for appropriately logging http status code to a terminal.
func (p *LogFormatterParams) StatusCodeColor() string {
	code := p.StatusCode

	switch {
	case code >= http.StatusOK && code < http.StatusMultipleChoices:
		return green
	case code >= http.StatusMultipleChoices && code < http.StatusBadRequest:
		return white
	case code >= http.StatusBadRequest && code < http.StatusInternalServerError:
		return yellow
	default:
		return red
	}
}

// MethodColor is the ANSI color for appropriately logging http method to a terminal.
func (p *LogFormatterParams) MethodColor() string {
	method := p.Method

	switch method {
	case http.MethodGet:
		return blue
	case http.MethodPost:
		return cyan
	case http.MethodPut:
		return yellow
	case http.MethodDelete:
		return red
	case http.MethodPatch:
		return green
	case http.MethodHead:
		return magenta
	case http.MethodOptions:
		return white
	default:
		return reset
	}
}

// ResetColor resets all escape attributes.
func (p *LogFormatterParams) ResetColor() string {
	return reset
}

// IsOutputColor indicates whether can colors be outputted to the log.
func (p *LogFormatterParams) IsOutputColor() bool {
	return p.colorMode == ColorForce
}

// Logger instances a Logger middleware that will write the logs to console.
// By default, gin.DefaultWriter = os.Stdout.
func Logger(next http.Handler) (http.Handler, error) {
	return HandlerWithConfig(LoggerConfig{
		ProxyHandler: NewProxy(),
	}, next)
}

// HandlerWithName instance a Logger handler with the specified prefix name and next handler.
func HandlerWithName(routerName string, next http.Handler) (http.Handler, error) {
	return HandlerWithConfig(LoggerConfig{
		ProxyHandler: NewProxy(),
		RouterName:   routerName,
	}, next)
}

// HandlerWithFormatter instance a Logger handler with the specified log format function and next handler.
func HandlerWithFormatter(f LogFormatter, next http.Handler) (http.Handler, error) {
	return HandlerWithConfig(LoggerConfig{
		Formatter:    f,
		ProxyHandler: NewProxy(),
	}, next)
}

// HandlerWithFormatterAndName instance a Logger handler with the specified log format function and next handler.
func HandlerWithFormatterAndName(routerName string, f LogFormatter, next http.Handler) (http.Handler, error) {
	return HandlerWithConfig(LoggerConfig{
		Formatter:    f,
		ProxyHandler: NewProxy(),
		RouterName:   routerName,
	}, next)
}

// HandlerWithWriter instance a Logger handler with the specified writer buffer and next handler.
// Example: os.Stdout, a file opened in write mode, a socket...
func HandlerWithWriter(out io.Writer, next http.Handler, notlogged ...string) (http.Handler, error) {
	return HandlerWithConfig(LoggerConfig{
		Output:       out,
		SkipPaths:    notlogged,
		ProxyHandler: NewProxy(),
	}, next)
}

// HandlerWithConfig instance a Logger handler with config and next handler.
func HandlerWithConfig(conf LoggerConfig, next http.Handler) (http.Handler, error) {
	loggingMiddleware, err := LoggerWithConfig(conf)
	if err != nil {
		return nil, err
	}
	return loggingMiddleware.Handler(next), nil
}

func maskHeaderKeys(h http.Header, keys []*regexp.Regexp) http.Header {
	for k, v := range h {
		for _, rx := range keys {
			if rx.MatchString(k) {
				for iv, vv := range v {
					h[k][iv] = masked(vv)
				}
			}
		}
	}
	return h
}

// returns ten asterisks for short string
// and first and last runes with ten asterisks between for long strings
func masked(s string) string {
	if len(s) < 10 {
		return "**********"
	} else {
		runes := []rune(s)
		return string(runes[0:1]) + "**********" + string(runes[len(runes)-1:])
	}
}
