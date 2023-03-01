package ginlog

import (
	"io"
	"net/http"

	"github.com/MadAppGang/httplog"
	"github.com/gin-gonic/gin"
)

// Logger returns default logger echo middleware
func Logger() gin.HandlerFunc {
	return ginLogger(httplog.Logger)
}

// LoggerWithName instance a Logger middleware with the specified name prefix.
func LoggerWithName(routerName string) gin.HandlerFunc {
	logger := httplog.LoggerWithName(routerName)
	return ginLogger(logger)
}

// LoggerWithFormatter instance a Logger middleware with the specified log format function.
func LoggerWithFormatter(f httplog.LogFormatter) gin.HandlerFunc {
	logger := httplog.LoggerWithFormatter(f)
	return ginLogger(logger)
}

// LoggerWithFormatterAndName instance a Logger middleware with the specified log format function.
func LoggerWithFormatterAndName(routerName string, f httplog.LogFormatter) gin.HandlerFunc {
	logger := httplog.LoggerWithFormatterAndName(routerName, f)
	return ginLogger(logger)
}

// LoggerWithWriter instance a Logger middleware with the specified writer buffer.
// Example: os.Stdout, a file opened in write mode, a socket...
func LoggerWithWriter(out io.Writer, notlogged ...string) gin.HandlerFunc {
	logger := httplog.LoggerWithWriter(out, notlogged...)
	return ginLogger(logger)
}

// LoggerWithConfig instance a Logger middleware with config.
func LoggerWithConfig(conf httplog.LoggerConfig) gin.HandlerFunc {
	logger := httplog.LoggerWithConfig(conf)
	return ginLogger(logger)
}

// embed canonical logger to gin middleware format.
func ginLogger(logger func(next http.Handler) http.Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		handler := http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			// gin uses wrapped ResponseWriter, we don't want to replace it
			// as gin use custom middleware approach with Next() method and context
			c.Next()
			// set result for ResponseWriter manually
			rwr, _ := rw.(httplog.ResponseWriter)
			rwr.Set(c.Writer.Status(), c.Writer.Size())
		})
		logger(handler).ServeHTTP(c.Writer, c.Request)
	}
}
