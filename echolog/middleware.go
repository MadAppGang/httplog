package echolog

import (
	"io"
	"net/http"

	"github.com/MadAppGang/httplog"
	"github.com/labstack/echo/v4"
)

// Logger returns default logger echo middleware
func Logger() func(next echo.HandlerFunc) echo.HandlerFunc {
	return echoLogger(httplog.Logger)
}

// LoggerWithName instance a Logger middleware with the specified name prefix.
func LoggerWithName(routerName string) func(next echo.HandlerFunc) echo.HandlerFunc {
	logger := httplog.LoggerWithName(routerName)
	return echoLogger(logger)
}

// LoggerWithFormatter instance a Logger middleware with the specified log format function.
func LoggerWithFormatter(f httplog.LogFormatter) func(next echo.HandlerFunc) echo.HandlerFunc {
	logger := httplog.LoggerWithFormatter(f)
	return echoLogger(logger)
}

// LoggerWithFormatterAndName instance a Logger middleware with the specified log format function.
func LoggerWithFormatterAndName(routerName string, f httplog.LogFormatter) func(next echo.HandlerFunc) echo.HandlerFunc {
	logger := httplog.LoggerWithFormatterAndName(routerName, f)
	return echoLogger(logger)
}

// LoggerWithWriter instance a Logger middleware with the specified writer buffer.
// Example: os.Stdout, a file opened in write mode, a socket...
func LoggerWithWriter(out io.Writer, notlogged ...string) func(next echo.HandlerFunc) echo.HandlerFunc {
	logger := httplog.LoggerWithWriter(out, notlogged...)
	return echoLogger(logger)
}

// LoggerWithConfig instance a Logger middleware with config.
func LoggerWithConfig(conf httplog.LoggerConfig) func(next echo.HandlerFunc) echo.HandlerFunc {
	logger := httplog.LoggerWithConfig(conf)
	return echoLogger(logger)
}

func echoLogger(logger func(next http.Handler) http.Handler) func(next echo.HandlerFunc) echo.HandlerFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			handler := http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
				if err := next(c); err != nil {
					c.Error(err)
				}
				// echo is overwriting http.ResponseWriter to proxy data write (c.Response() as echo.Response type)
				// we just need manually bring this data back to our log's response writer (rw variable)
				lrw, _ := rw.(httplog.ResponseWriter)
				lrw.Set(c.Response().Status, int(c.Response().Size))
			})
			logger(handler).ServeHTTP(c.Response().Writer, c.Request())
			return nil
		}
	}
}
