package ginlog

import (
	"io"
	"net/http"

	"github.com/MadAppGang/httplog/v2"
	"github.com/gin-gonic/gin"
)

// Logger returns default logger echo middleware
func Logger() (gin.HandlerFunc, error) {
	logger, err := httplog.LoggerWithConfig(httplog.LoggerConfig{
		ProxyHandler: httplog.NewProxy(),
	})
	if err != nil {
		return nil, err
	}
	return ginLogger(logger.Handler), nil
}

// LoggerWithName instance a Logger middleware with the specified name prefix.
func LoggerWithName(routerName string) (gin.HandlerFunc, error) {
	logger, err := httplog.LoggerWithName(routerName)
	if err != nil {
		return nil, err
	}
	return ginLogger(logger.Handler), nil
}

// LoggerWithFormatter instance a Logger middleware with the specified log format function.
func LoggerWithFormatter(f httplog.LogFormatter) (gin.HandlerFunc, error) {
	logger, err := httplog.LoggerWithFormatter(f)
	if err != nil {
		return nil, err
	}
	return ginLogger(logger.Handler), nil
}

// LoggerWithFormatterAndName instance a Logger middleware with the specified log format function.
func LoggerWithFormatterAndName(routerName string, f httplog.LogFormatter) (gin.HandlerFunc, error) {
	logger, err := httplog.LoggerWithFormatterAndName(routerName, f)
	if err != nil {
		return nil, err
	}
	return ginLogger(logger.Handler), nil
}

// LoggerWithWriter instance a Logger middleware with the specified writer buffer.
// Example: os.Stdout, a file opened in write mode, a socket...
func LoggerWithWriter(out io.Writer, notlogged ...string) (gin.HandlerFunc, error) {
	logger, err := httplog.LoggerWithWriter(out, notlogged...)
	if err != nil {
		return nil, err
	}
	return ginLogger(logger.Handler), nil
}

// LoggerWithConfig instance a Logger middleware with config.
func LoggerWithConfig(conf httplog.LoggerConfig) (gin.HandlerFunc, error) {
	logger, err := httplog.LoggerWithConfig(conf)
	if err != nil {
		return nil, err
	}
	return ginLogger(logger.Handler), nil
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
