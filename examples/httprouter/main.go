package main

import (
	"bytes"
	"fmt"
	"net/http"
	"time"

	"github.com/MadAppGang/httplog/v2"
	"github.com/julienschmidt/httprouter"
)

func LoggerMiddleware(h httprouter.Handle) httprouter.Handle {
	logger, err := httplog.LoggerWithName("ME")
	if err != nil {
		panic(err)
	}
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h(w, r, ps)
		})
		logger.Handler(handler).ServeHTTP(w, r)
	}
}

var happyHandler httprouter.Handle = func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	w.Write([]byte("I am happy!"))
}

var notFoundHandler httprouter.Handle = func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	http.NotFound(w, r)
}

func main() {
	// setup routes

	router := httprouter.New()
	router.GET("/happy", LoggerMiddleware(happyHandler))
	router.POST("/happy", LoggerMiddleware(happyHandler))
	router.GET("/not_found", LoggerMiddleware(notFoundHandler))

	go func() {
		fmt.Println("Server started at port 3333")
		err := http.ListenAndServe(":3333", router)
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
