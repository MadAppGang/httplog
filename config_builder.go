package httplog

import "io"

// Copyright 2022 Jack Rudenko. MadAppGang. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

// ConfigBuilder provides fluent API for building LoggerConfig
type ConfigBuilder struct {
	config LoggerConfig
}

// NewConfigBuilder creates a new builder with defaults
func NewConfigBuilder() *ConfigBuilder {
	return &ConfigBuilder{
		config: LoggerConfig{
			ColorMode:       ColorAuto,
			Formatter:       nil, // Will use DefaultLogFormatter if nil
			Output:          nil, // Will use DefaultWriter if nil
			ProxyHandler:    nil, // Will use NewProxy() if nil
			SampleRate:      1.0, // 100% sampling by default
			MinLevel:        LevelInfo,
			AsyncBufferSize: 1000,
		},
	}
}

// WithFormatter sets the log formatter
func (b *ConfigBuilder) WithFormatter(f LogFormatter) *ConfigBuilder {
	b.config.Formatter = f
	return b
}

// WithOutput sets the output writer
func (b *ConfigBuilder) WithOutput(w io.Writer) *ConfigBuilder {
	b.config.Output = w
	return b
}

// WithColorMode sets the color mode
func (b *ConfigBuilder) WithColorMode(mode ColorMode) *ConfigBuilder {
	b.config.ColorMode = mode
	return b
}

// WithRouterName sets the router name prefix
func (b *ConfigBuilder) WithRouterName(name string) *ConfigBuilder {
	b.config.RouterName = name
	return b
}

// WithSkipPaths adds paths to skip logging
func (b *ConfigBuilder) WithSkipPaths(paths ...string) *ConfigBuilder {
	b.config.SkipPaths = append(b.config.SkipPaths, paths...)
	return b
}

// WithHideHeaderKeys adds header keys to mask
func (b *ConfigBuilder) WithHideHeaderKeys(keys ...string) *ConfigBuilder {
	b.config.HideHeaderKeys = append(b.config.HideHeaderKeys, keys...)
	return b
}

// WithProxyHandler sets the proxy handler
func (b *ConfigBuilder) WithProxyHandler(proxy *Proxy) *ConfigBuilder {
	b.config.ProxyHandler = proxy
	return b
}

// WithCaptureResponseBody enables response body capture
func (b *ConfigBuilder) WithCaptureResponseBody(capture bool) *ConfigBuilder {
	b.config.CaptureResponseBody = capture
	return b
}

// WithCaptureRequestBody enables request body capture
func (b *ConfigBuilder) WithCaptureRequestBody(capture bool) *ConfigBuilder {
	b.config.CaptureRequestBody = capture
	return b
}

// WithSampleRate sets the sample rate (0.0-1.0)
func (b *ConfigBuilder) WithSampleRate(rate float64) *ConfigBuilder {
	b.config.SampleRate = rate
	return b
}

// WithAsyncLogging enables async logging with buffer size
func (b *ConfigBuilder) WithAsyncLogging(enabled bool, bufferSize int) *ConfigBuilder {
	b.config.AsyncLogging = enabled
	b.config.AsyncBufferSize = bufferSize
	return b
}

// WithMinLevel sets the minimum log level
func (b *ConfigBuilder) WithMinLevel(level Level) *ConfigBuilder {
	b.config.MinLevel = level
	return b
}

// Build creates LoggerConfig and validates it
// Returns error if configuration is invalid
func (b *ConfigBuilder) Build() (LoggerConfig, error) {
	if err := ValidateConfig(b.config); err != nil {
		return LoggerConfig{}, err
	}
	return b.config, nil
}
