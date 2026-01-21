package httplog

import (
	"log/slog"
)

// Copyright 2022 Jack Rudenko. MadAppGang. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

// SlogLogger returns a LogFormatter that logs to slog
// Parameters:
//   - logger: *slog.Logger instance
//   - level: slog.Level (LevelDebug, LevelInfo, LevelWarn, LevelError)
//   - message: Log message prefix (e.g., "HTTP Request")
//
// The formatter extracts trace ID from context (if present) and logs
// structured attributes: method, path, status, latency, client_ip, body_size
//
// Example:
//
//	slogger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
//	conf := httplog.LoggerConfig{
//	    Formatter: httplog.SlogLogger(slogger, slog.LevelInfo, "HTTP"),
//	}
//	middleware, _ := httplog.LoggerWithConfig(conf)
func SlogLogger(logger *slog.Logger, level slog.Level, message string) LogFormatter {
	return func(param LogFormatterParams) string {
		// Map httplog.Level to slog.Level
		var slogLevel slog.Level
		switch param.Level {
		case LevelDebug:
			slogLevel = slog.LevelDebug
		case LevelInfo:
			slogLevel = slog.LevelInfo
		case LevelWarn:
			slogLevel = slog.LevelWarn
		case LevelError:
			slogLevel = slog.LevelError
		default:
			slogLevel = level // Use provided default
		}

		// Extract context for trace ID
		ctx := param.Context
		if ctx == nil {
			ctx = param.Request.Context()
		}

		// Build structured log attributes
		attrs := []slog.Attr{
			slog.String("method", param.Method),
			slog.String("path", param.Path),
			slog.Int("status", param.StatusCode),
			slog.Duration("latency", param.Latency),
			slog.String("client_ip", param.ClientIP),
			slog.Int("body_size", param.BodySize),
		}

		// Add router name if present
		if param.RouterName != "" {
			attrs = append(attrs, slog.String("router", param.RouterName))
		}

		// Log with context (for trace ID extraction if middleware is present)
		logger.LogAttrs(ctx, slogLevel, message, attrs...)

		// Return empty string as slog handles output
		return ""
	}
}

// DefaultSlogLogger creates a default slog formatter with Info level
func DefaultSlogLogger(logger *slog.Logger) LogFormatter {
	return SlogLogger(logger, slog.LevelInfo, "HTTP Request")
}

// DefaultSlogLoggerWithHeaders creates a slog formatter that also logs headers
func DefaultSlogLoggerWithHeaders(logger *slog.Logger) LogFormatter {
	return ChainLogFormatter(
		SlogLogger(logger, slog.LevelInfo, "HTTP Request"),
		RequestHeaderLogFormatter,
		ResponseHeaderLogFormatter,
	)
}
