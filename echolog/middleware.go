package echolog

import (
	"net/http"

	"github.com/MadAppGang/httplog"
	"github.com/labstack/echo"
)

func LoggerWithConfig(conf httplog.LoggerConfig) func(next echo.HandlerFunc) echo.HandlerFunc {
	logger := httplog.LoggerWithConfig(conf)
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
