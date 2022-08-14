package main

import (
	"bytes"
	"fmt"
	"net/http"
	"time"

	"github.com/MadAppGang/httplog"
	"github.com/gin-gonic/gin"
)

// httplog.ResponseWriter is not fully compatible with gin.ResponseWriter
func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		l := httplog.Logger(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			// gin uses wrapped ResponseWriter, we don't want to replace it
			// as gin use custom middleware approach with Next() method and context
			c.Next()
			// set result for ResponseWriter manually
			rwr, _ := rw.(httplog.ResponseWriter)
			rwr.Set(c.Writer.Status(), c.Writer.Size())
		}))
		l.ServeHTTP(c.Writer, c.Request)
	}
}

var happyHandler = func(c *gin.Context) {
	fmt.Println("I am happy handler")
	c.Writer.Write([]byte("I am happy!"))
}

func main() {
	// setup routes
	r := gin.New()
	r.Use(LoggerMiddleware())
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
