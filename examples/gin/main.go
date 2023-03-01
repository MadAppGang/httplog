package main

import (
	"bytes"
	"fmt"
	"net/http"
	"time"

	"github.com/MadAppGang/httplog/ginlog"
	"github.com/gin-gonic/gin"
)

var happyHandler = func(c *gin.Context) {
	fmt.Println("I am happy handler")
	c.Writer.Write([]byte("I am happy!"))
}

func main() {
	// setup routes
	r := gin.New()
	r.Use(ginlog.LoggerWithName("I AM GIN ROUTER"))
	r.GET("/happy", happyHandler)
	r.POST("/happy", happyHandler)
	r.GET("/not_found", gin.WrapF(http.NotFound))

	go func() {
		fmt.Println("Server started at port 3333")
		err := r.Run(":3333")
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
