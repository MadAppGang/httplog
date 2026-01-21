package main

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/MadAppGang/httplog/v2"
	"github.com/go-mojito/mojito"
)

func happyHandler(ctx mojito.Context) error {
	return ctx.String("I am happy!")
}

func main() {
	notFoundHandler, err := httplog.Logger(http.NotFoundHandler())
	if err != nil {
		panic(err)
	}
	mojito.WithNotFoundHandler(notFoundHandler)

	logger, err := httplog.LoggerWithName("🍸🍸🍸🍸")
	if err != nil {
		panic(err)
	}
	mojito.WithMiddleware(logger)
	mojito.POST("/happy", happyHandler)
	mojito.GET("/happy", happyHandler)

	go func() {
		fmt.Println("Server started at port 3333")
		err := mojito.ListenAndServe((":3333"))
		if err != nil {
			fmt.Printf("Server stopped because of error %s\n", err.Error())
		}
	}()

	// let's make couple of request
	_, err = http.Get("http://localhost:3333/happy")
	if err != nil {
		fmt.Printf("Error: %+v", err)
	}
	_, _ = http.Post("http://localhost:3333/happy", "text/plain", bytes.NewBuffer([]byte("I am not ")))
	_, _ = http.Get("http://localhost:3333/not_found")

	fmt.Println("All done, thank you and see you soon 👋")
}
