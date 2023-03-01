package httplog

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"time"

	"github.com/mattn/go-isatty"
)

// Copyright 2022 Jack Rudenko. MadAppGang. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

// LoggerWithName instance a Logger middleware with the specified name prefix.
func LoggerWithName(routerName string) func(next http.Handler) http.Handler {
	return LoggerWithConfig(LoggerConfig{
		ProxyHandler: NewProxy(),
		RouterName:   routerName,
	})
}

// LoggerWithFormatter instance a Logger middleware with the specified log format function.
func LoggerWithFormatter(f LogFormatter) func(next http.Handler) http.Handler {
	return LoggerWithConfig(LoggerConfig{
		Formatter:    f,
		ProxyHandler: NewProxy(),
	})
}

// LoggerWithFormatterAndName instance a Logger middleware with the specified log format function.
func LoggerWithFormatterAndName(routerName string, f LogFormatter) func(next http.Handler) http.Handler {
	return LoggerWithConfig(LoggerConfig{
		Formatter:    f,
		ProxyHandler: NewProxy(),
		RouterName:   routerName,
	})
}

// LoggerWithWriter instance a Logger middleware with the specified writer buffer.
// Example: os.Stdout, a file opened in write mode, a socket...
func LoggerWithWriter(out io.Writer, notlogged ...string) func(next http.Handler) http.Handler {
	return LoggerWithConfig(LoggerConfig{
		Output:       out,
		SkipPaths:    notlogged,
		ProxyHandler: NewProxy(),
	})
}

// LoggerWithConfig instance a Logger middleware with config.
func LoggerWithConfig(conf LoggerConfig) func(next http.Handler) http.Handler {
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

	middleware := func(next http.Handler) http.Handler {
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
	return middleware
}
