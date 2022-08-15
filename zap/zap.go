package zap

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/MadAppGang/httplog"
)

// ZapLogger log everything to zap logger with specific log level, if message is empty, URL is used instead
func ZapLogger(zl *zap.Logger, level zapcore.Level, message string) httplog.FormatterFunction {
	return func(params httplog.LogFormatterParams) string {
		if zl == nil {
			return ""
		}
		if len(message) == 0 {
			message = fmt.Sprintf("[%s] response %s", params.RouterName, params.Path)
		}

		zl.Log(level, message,
			zap.String("RouterName", params.RouterName),
			zap.Time("TimeStamp", params.TimeStamp),
			zap.Int("StatusCode", params.StatusCode),
			zap.Duration("Latency", params.Latency),
			zap.String("ClientIP", params.ClientIP),
			zap.String("Method", params.Method),
			zap.String("Path", params.Path),
			zap.Int("BodySize", params.BodySize),
		)
		return ""
	}
}

// Combines default logger and Zap logger
func DefaultZapLogger(zl *zap.Logger, level zapcore.Level, message string) httplog.FormatterFunction {
	return httplog.ChainLogFormatter(httplog.DefaultLogFormatter, ZapLogger(zl, level, message))
}

// DefaultZapLoggerWithHeaders combine default formatter, headers output and logger
func DefaultZapLoggerWithHeaders(zl *zap.Logger, level zapcore.Level, message string) httplog.FormatterFunction {
	return httplog.ChainLogFormatter(
		httplog.DefaultLogFormatter,
		httplog.HeadersLogFormatter,
		ZapLogger(zl, level, message),
	)
}

// DefaultZapLoggerWithHeaders combine default formatter, headers output, body output and logger
func DefaultZapLoggerWithHeadersAndBody(zl *zap.Logger, level zapcore.Level, message string) httplog.FormatterFunction {
	return httplog.ChainLogFormatter(
		httplog.DefaultLogFormatter,
		httplog.HeadersLogFormatter,
		httplog.BodyLogFormatter,
		ZapLogger(zl, level, message),
	)
}
