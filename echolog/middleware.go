package echolog

import (
	"io"
	"net/http"

	"github.com/MadAppGang/httplog/v2"
	"github.com/labstack/echo/v4"
)

// Logger returns default logger echo middleware
func Logger() (func(next echo.HandlerFunc) echo.HandlerFunc, error) {
	logger, err := httplog.LoggerWithConfig(httplog.LoggerConfig{
		ProxyHandler: httplog.NewProxy(),
	})
	if err != nil {
		return nil, err
	}
	return echoLogger(logger.Handler), nil
}

// LoggerWithName instance a Logger middleware with the specified name prefix.
func LoggerWithName(routerName string) (func(next echo.HandlerFunc) echo.HandlerFunc, error) {
	logger, err := httplog.LoggerWithName(routerName)
	if err != nil {
		return nil, err
	}
	return echoLogger(logger.Handler), nil
}

// LoggerWithFormatter instance a Logger middleware with the specified log format function.
func LoggerWithFormatter(f httplog.LogFormatter) (func(next echo.HandlerFunc) echo.HandlerFunc, error) {
	logger, err := httplog.LoggerWithFormatter(f)
	if err != nil {
		return nil, err
	}
	return echoLogger(logger.Handler), nil
}

// LoggerWithFormatterAndName instance a Logger middleware with the specified log format function.
func LoggerWithFormatterAndName(routerName string, f httplog.LogFormatter) (func(next echo.HandlerFunc) echo.HandlerFunc, error) {
	logger, err := httplog.LoggerWithFormatterAndName(routerName, f)
	if err != nil {
		return nil, err
	}
	return echoLogger(logger.Handler), nil
}

// LoggerWithWriter instance a Logger middleware with the specified writer buffer.
// Example: os.Stdout, a file opened in write mode, a socket...
func LoggerWithWriter(out io.Writer, notlogged ...string) (func(next echo.HandlerFunc) echo.HandlerFunc, error) {
	logger, err := httplog.LoggerWithWriter(out, notlogged...)
	if err != nil {
		return nil, err
	}
	return echoLogger(logger.Handler), nil
}

// LoggerWithConfig instance a Logger middleware with config.
func LoggerWithConfig(conf httplog.LoggerConfig) (func(next echo.HandlerFunc) echo.HandlerFunc, error) {
	logger, err := httplog.LoggerWithConfig(conf)
	if err != nil {
		return nil, err
	}
	return echoLogger(logger.Handler), nil
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
