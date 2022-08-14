package main

import (
	"bytes"
	"fmt"
	"net/http"
	"time"

	"github.com/MadAppGang/httplog"
	"github.com/labstack/echo/v4"
)

func loggerMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		logger := httplog.Logger(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			if err := next(c); err != nil {
				c.Error(err)
			}
			// echo is overwriting http.ResponseWriter to proxy data write (c.Response() as echo.Response type)
			// we just need manually bring this data back to our log's response writer (rw variable)
			lrw, _ := rw.(httplog.ResponseWriter)
			lrw.Set(c.Response().Status, int(c.Response().Size))
		}))
		logger.ServeHTTP(c.Response().Writer, c.Request())
		return nil
	}
}

func happyHandler(c echo.Context) error {
	c.Response().Writer.Write([]byte("I am happy!"))
	return nil
}

func main() {
	// setup routes
	e := echo.New()

	// Middleware
	e.Use(loggerMiddleware)
	e.GET("/happy", happyHandler)
	e.POST("/happy", happyHandler)
	e.GET("/not_found", echo.NotFoundHandler)

	go func() {
		fmt.Println("Server started at port 3333")
		err := e.Start(":3333")
		if err != nil {
			fmt.Printf("Server stopped because of error %s\n", err.Error())
		}
	}()

	// let server to start
	time.Sleep(time.Second * 2)

	// let's make couple of request
	_, err := http.Get("http://localhost:3333/happy")
	if err != nil {
		fmt.Printf("Error: %+v", err)
	}
	_, _ = http.Post("http://localhost:3333/happy", "text/plain", bytes.NewBuffer([]byte("I am not ")))
	_, _ = http.Get("http://localhost:3333/not_found")

	fmt.Println("All done, thank you and see you soon 👋")
}
