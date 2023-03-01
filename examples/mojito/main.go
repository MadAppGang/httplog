package main

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/MadAppGang/httplog"
	"github.com/go-mojito/mojito"
)

func happyHandler(ctx mojito.Context) error {
	return ctx.String("I am happy!")
}

func main() {
	mojito.WithNotFoundHandler(httplog.Logger(http.NotFoundHandler()))
	mojito.WithMiddleware(httplog.LoggerWithName("🍸🍸🍸🍸"))
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
	_, err := http.Get("http://localhost:3333/happy")
	if err != nil {
		fmt.Printf("Error: %+v", err)
	}
	_, _ = http.Post("http://localhost:3333/happy", "text/plain", bytes.NewBuffer([]byte("I am not ")))
	_, _ = http.Get("http://localhost:3333/not_found")

	fmt.Println("All done, thank you and see you soon 👋")
}
