package main

import (
	"bytes"
	"fmt"
	"net/http"
	"time"

	"github.com/MadAppGang/httplog/v2"
	"github.com/urfave/negroni"
)

var happyHandler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("I am happy!"))
})

// use function Curry pattern
var negroniLoggerMiddleware negroni.Handler = negroni.HandlerFunc(func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	logger, err := httplog.LoggerWithName("negroni") // setup your router here
	if err != nil {
		panic(err)
	}
	logger.Handler(next).ServeHTTP(rw, r)
})

func main() {
	// setup routes

	mux := http.NewServeMux()
	mux.Handle("/happy", happyHandler)
	mux.Handle("/not_found", http.NotFoundHandler())

	n := negroni.New()
	n.Use(negroniLoggerMiddleware)
	n.UseHandler(mux)

	go func() {
		fmt.Println("Server started at port 3333")
		err := http.ListenAndServe(":3333", n)
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
