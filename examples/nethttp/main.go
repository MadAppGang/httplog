package main

import (
	"bytes"
	"fmt"
	"net/http"
	"time"

	"github.com/MadAppGang/httplog"
)

var happyHandler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Happy handler log.")
	w.Write([]byte("I am happy!"))
})

func main() {
	// setup routes
	http.Handle("/happy", httplog.LoggerWithFormatter(httplog.DefaultLogFormatter, happyHandler))
	http.Handle("/not_found", httplog.Logger(http.NotFoundHandler()))

	go func() {
		fmt.Println("Server started at port 3333")
		err := http.ListenAndServe(":3333", nil)
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
