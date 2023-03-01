package main

import (
	"bytes"
	"fmt"
	"net/http"
	"time"

	"github.com/MadAppGang/httplog/echolog"
	"github.com/labstack/echo/v4"
)

func happyHandler(c echo.Context) error {
	c.Response().Writer.Write([]byte("I am happy!"))
	return nil
}

func main() {
	// setup routes
	e := echo.New()

	// Middleware
	e.Use(echolog.LoggerWithName("ECHO NATIVE"))
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
