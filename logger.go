package httplog

// Copyright 2022 Jack Rudenko. MadAppGang. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"time"

	"github.com/mattn/go-isatty"
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

// Logger instances a Logger middleware that will write the logs to gin.DefaultWriter.
// By default, gin.DefaultWriter = os.Stdout.
func Logger(next http.Handler) http.Handler {
	return LoggerWithConfig(LoggerConfig{
		ProxyHandler: NewProxy(),
	}, next)
}

func LoggerWithName(routerName string, next http.Handler) http.Handler {
	return LoggerWithConfig(LoggerConfig{
		ProxyHandler: NewProxy(),
		RouterName:   routerName,
	}, next)
}

// LoggerWithFormatter instance a Logger middleware with the specified log format function.
func LoggerWithFormatter(f LogFormatter, next http.Handler) http.Handler {
	return LoggerWithConfig(LoggerConfig{
		Formatter:    f,
		ProxyHandler: NewProxy(),
	}, next)
}

// LoggerWithFormatter instance a Logger middleware with the specified log format function.
func LoggerWithFormatterAndName(routerName string, f LogFormatter, next http.Handler) http.Handler {
	return LoggerWithConfig(LoggerConfig{
		Formatter:    f,
		ProxyHandler: NewProxy(),
		RouterName:   routerName,
	}, next)
}

// LoggerWithWriter instance a Logger middleware with the specified writer buffer.
// Example: os.Stdout, a file opened in write mode, a socket...
func LoggerWithWriter(out io.Writer, next http.Handler, notlogged ...string) http.Handler {
	return LoggerWithConfig(LoggerConfig{
		Output:       out,
		SkipPaths:    notlogged,
		ProxyHandler: NewProxy(),
	}, next)
}

// LoggerWithConfig instance a Logger middleware with config.
func LoggerWithConfig(conf LoggerConfig, next http.Handler) http.Handler {
	formatter := conf.Formatter
	if formatter == nil {
		formatter = DefaultLogFormatter
	}

	if conf.ProxyHandler == nil {
		conf.ProxyHandler = NewProxy()
	}

	out := conf.Output
	if out == nil {
		out = DefaultWriter
	}

	isTerm := true

	if w, ok := out.(*os.File); !ok || os.Getenv("TERM") == "dumb" ||
		(!isatty.IsTerminal(w.Fd()) && !isatty.IsCygwinTerminal(w.Fd())) {
		isTerm = false
	}

	var skipPath []*regexp.Regexp
	for _, p := range conf.SkipPaths {
		re, err := regexp.Compile(p)
		if err == nil {
			skipPath = append(skipPath, re)
		} else {
			fmt.Fprint(out, fmt.Sprintf("error parsing skip path regex, ignoring: %s", p))
		}
	}

	var hideHeaderKeys []*regexp.Regexp
	for _, p := range conf.HideHeaderKeys {
		re, err := regexp.Compile(p)
		if err == nil {
			hideHeaderKeys = append(hideHeaderKeys, re)
		} else {
			fmt.Fprint(out, fmt.Sprintf("error parsing header key regexp to hide, ignoring: %s", p))
		}
	}

	// http.StripPrefix()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Start timer
		start := time.Now()
		path := r.URL.Path
		raw := r.URL.RawQuery

		// Process request
		// Wrap response writer with Recorded response writer
		wr := NewWriter(w, conf.CaptureBody)
		next.ServeHTTP(wr, r)

		var skip bool
		// check path for skip regexp set
		for _, r := range skipPath {
			if r.MatchString(path) {
				skip = true
				break
			}
		}

		// Log only when path is not being skipped
		if skip == false {
			r.Header = maskHeaderKeys(r.Header.Clone(), hideHeaderKeys)

			param := LogFormatterParams{
				Request: r,
				isTerm:  isTerm,
			}

			// Stop timer
			param.TimeStamp = time.Now()
			param.Latency = param.TimeStamp.Sub(start)

			param.ClientIP = conf.ProxyHandler.ClientIP(r)
			param.Method = r.Method
			param.StatusCode = wr.Status()

			param.BodySize = wr.Size()
			param.Body = wr.Body()
			param.ResponseHeader = maskHeaderKeys(wr.Header().Clone(), hideHeaderKeys)

			param.RouterName = conf.RouterName

			if raw != "" {
				path = path + "?" + raw
			}

			param.Path = path

			fmt.Fprint(out, formatter(param))
		}
	})
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
