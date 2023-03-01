package httplog

// Copyright 2022 Jack Rudenko. MadAppGang. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

import (
	"io"
	"net/http"
	"regexp"
	"time"
)

type consoleColorModeValue int

const (
	autoColor consoleColorModeValue = iota
	disableColor
	forceColor
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

var consoleColorMode = autoColor

// LoggerConfig defines the config for Logger middleware.
type LoggerConfig struct {
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

	// CaptureBody saves response body copy for debug  purposes
	CaptureBody bool
}

// LogFormatter gives the signature of the formatter function passed to LoggerWithFormatter
// you can use predefined, like httplog.DefaultLogFormatter or httplog.ShortLogFormatter
// or you can create your custom
type LogFormatter func(params LogFormatterParams) string

// LogFormatterParams is the structure any formatter will be handed when time to log comes
type LogFormatterParams struct {
	Request *http.Request

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
	// isTerm shows whether output descriptor refers to a terminal.
	isTerm bool
	// BodySize is the size of the Response Body
	BodySize int
	// Body is a body content, if body is copied
	Body []byte
	// Response header
	ResponseHeader http.Header
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
	return consoleColorMode == forceColor || (consoleColorMode == autoColor && p.isTerm)
}

// DisableConsoleColor disables color output in the console.
func DisableConsoleColor() {
	consoleColorMode = disableColor
}

// ForceConsoleColor force color output in the console.
func ForceConsoleColor() {
	consoleColorMode = forceColor
}

// Logger instances a Logger middleware that will write the logs to console.
// By default, gin.DefaultWriter = os.Stdout.
func Logger(next http.Handler) http.Handler {
	return HandlerWithConfig(LoggerConfig{
		ProxyHandler: NewProxy(),
	}, next)
}

// HandlerWithName instance a Logger handler with the specified prefix name and next handler.
func HandlerWithName(routerName string, next http.Handler) http.Handler {
	return HandlerWithConfig(LoggerConfig{
		ProxyHandler: NewProxy(),
		RouterName:   routerName,
	}, next)
}

// HandlerWithFormatter instance a Logger handler with the specified log format function and next handler.
func HandlerWithFormatter(f LogFormatter, next http.Handler) http.Handler {
	return HandlerWithConfig(LoggerConfig{
		Formatter:    f,
		ProxyHandler: NewProxy(),
	}, next)
}

// HandlerWithFormatterAndName instance a Logger handler with the specified log format function and next handler.
func HandlerWithFormatterAndName(routerName string, f LogFormatter, next http.Handler) http.Handler {
	return HandlerWithConfig(LoggerConfig{
		Formatter:    f,
		ProxyHandler: NewProxy(),
		RouterName:   routerName,
	}, next)
}

// HandlerWithWriter instance a Logger handler with the specified writer buffer and next handler.
// Example: os.Stdout, a file opened in write mode, a socket...
func HandlerWithWriter(out io.Writer, next http.Handler, notlogged ...string) http.Handler {
	return HandlerWithConfig(LoggerConfig{
		Output:       out,
		SkipPaths:    notlogged,
		ProxyHandler: NewProxy(),
	}, next)
}

// HandlerWithConfig instance a Logger handler with config and next handler.
func HandlerWithConfig(conf LoggerConfig, next http.Handler) http.Handler {
	return LoggerWithConfig(conf)(next)
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
