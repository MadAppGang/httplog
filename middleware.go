package httplog

import (
	"bytes"
	"fmt"
	"hash/fnv"
	"io"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"time"

	"github.com/mattn/go-isatty"
)

// Copyright 2022 Jack Rudenko. MadAppGang. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

// LoggingMiddleware wraps the logging middleware with lifecycle methods
type LoggingMiddleware struct {
	Handler func(next http.Handler) http.Handler
	logChan chan LogFormatterParams
}

// Close cleanly shuts down async logging goroutine
func (m *LoggingMiddleware) Close() {
	if m.logChan != nil {
		close(m.logChan)
	}
}

// LoggerWithName instance a Logger middleware with the specified name prefix.
func LoggerWithName(routerName string) (*LoggingMiddleware, error) {
	return LoggerWithConfig(LoggerConfig{
		ProxyHandler: NewProxy(),
		RouterName:   routerName,
	})
}

// LoggerWithFormatter instance a Logger middleware with the specified log format function.
func LoggerWithFormatter(f LogFormatter) (*LoggingMiddleware, error) {
	return LoggerWithConfig(LoggerConfig{
		Formatter:    f,
		ProxyHandler: NewProxy(),
	})
}

// LoggerWithFormatterAndName instance a Logger middleware with the specified log format function.
func LoggerWithFormatterAndName(routerName string, f LogFormatter) (*LoggingMiddleware, error) {
	return LoggerWithConfig(LoggerConfig{
		Formatter:    f,
		ProxyHandler: NewProxy(),
		RouterName:   routerName,
	})
}

// LoggerWithWriter instance a Logger middleware with the specified writer buffer.
// Example: os.Stdout, a file opened in write mode, a socket...
func LoggerWithWriter(out io.Writer, notlogged ...string) (*LoggingMiddleware, error) {
	return LoggerWithConfig(LoggerConfig{
		Output:       out,
		SkipPaths:    notlogged,
		ProxyHandler: NewProxy(),
	})
}

// LoggerWithConfig instance a Logger middleware with config.
func LoggerWithConfig(conf LoggerConfig) (*LoggingMiddleware, error) {
	// Validate configuration first
	if err := ValidateConfig(conf); err != nil {
		return nil, err
	}

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

	colorMode := conf.ColorMode
	// Auto-detect terminal if ColorAuto
	if colorMode == ColorAuto {
		if w, ok := out.(*os.File); ok && os.Getenv("TERM") != "dumb" &&
			(isatty.IsTerminal(w.Fd()) || isatty.IsCygwinTerminal(w.Fd())) {
			colorMode = ColorForce // Resolve to ColorForce when terminal detected
		} else {
			// Not a terminal, disable colors
			colorMode = ColorDisable
		}
	}

	var skipPath []*regexp.Regexp
	for _, p := range conf.SkipPaths {
		re, _ := regexp.Compile(p) // Already validated
		skipPath = append(skipPath, re)
	}

	var hideHeaderKeys []*regexp.Regexp
	for _, p := range conf.HideHeaderKeys {
		re, _ := regexp.Compile(p) // Already validated
		hideHeaderKeys = append(hideHeaderKeys, re)
	}

	// Set defaults for new config fields
	sampleRate := conf.SampleRate
	if sampleRate < 0 {
		sampleRate = 1.0 // Default to 100% sampling when unset (-1 or negative)
	}

	minLevel := conf.MinLevel
	// minLevel defaults to LevelDebug (0), which is fine

	asyncBufferSize := conf.AsyncBufferSize
	if conf.AsyncLogging && asyncBufferSize == 0 {
		asyncBufferSize = 1000 // Default buffer size
	}

	// Create async logging channel if enabled
	var logChan chan LogFormatterParams
	if conf.AsyncLogging {
		logChan = make(chan LogFormatterParams, asyncBufferSize)
		// Start background goroutine for async logging
		go func() {
			for param := range logChan {
				fmt.Fprint(out, formatter(param))
			}
		}()
	}

	middleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Start timer
			start := time.Now()
			path := r.URL.Path
			raw := r.URL.RawQuery

			// Capture request body if enabled
			var requestBody []byte
			if conf.CaptureRequestBody && r.Body != nil {
				var err error
				requestBody, err = io.ReadAll(r.Body)
				if err != nil {
					// Log error but don't fail - body capture is optional
					requestBody = []byte(fmt.Sprintf("[body read error: %v]", err))
				}
				r.Body = io.NopCloser(bytes.NewBuffer(requestBody))
			}

			// Process request
			// Wrap response writer with Recorded response writer
			wr := NewWriter(w, conf.CaptureResponseBody)
			next.ServeHTTP(wr, r)

			var skip bool
			// check path for skip regexp set
			for _, r := range skipPath {
				if r.MatchString(path) {
					skip = true
					break
				}
			}

			// Apply sampling (skip if sample rate check fails)
			if !skip && sampleRate > 0 && sampleRate < 1.0 {
				if conf.DeterministicSampling {
					// Hash-based sampling using path + method
					h := fnv.New32a()
					h.Write([]byte(r.Method + path))
					hashVal := float64(h.Sum32()) / float64(^uint32(0))
					if hashVal > sampleRate {
						skip = true
					}
				} else {
					if rand.Float64() > sampleRate {
						skip = true
					}
				}
			}

			// Log only when path is not being skipped
			if !skip {
				maskedReqHeader := maskHeaderKeys(r.Header.Clone(), hideHeaderKeys)

				param := LogFormatterParams{
					Request:       r,
					Context:       r.Context(),
					colorMode:     colorMode,
					RequestHeader: maskedReqHeader,
					RequestBody:   requestBody,
				}

				// Stop timer
				param.TimeStamp = time.Now()
				param.Latency = param.TimeStamp.Sub(start)

				param.ClientIP = conf.ProxyHandler.ClientIP(r)
				param.Method = r.Method
				param.StatusCode = wr.Status()

				// Set level based on status code
				param.Level = LevelFromStatusCode(param.StatusCode)

				// Apply level filtering
				if param.Level < minLevel {
					return
				}

				param.BodySize = wr.Size()
				param.ResponseBody = wr.Body()
				param.ResponseHeader = maskHeaderKeys(wr.Header().Clone(), hideHeaderKeys)

				param.RouterName = conf.RouterName

				if raw != "" {
					path = path + "?" + raw
				}

				param.Path = path

				// Write log (sync or async)
				if conf.AsyncLogging {
					select {
					case logChan <- param:
						// Successfully queued
					default:
						// Buffer full, drop log (or could block here)
					}
				} else {
					fmt.Fprint(out, formatter(param))
				}
			}
		})
	}

	return &LoggingMiddleware{
		Handler: middleware,
		logChan: logChan,
	}, nil
}
