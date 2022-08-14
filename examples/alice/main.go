package main

import (
	"bytes"
	"fmt"
	"net/http"
	"time"

	"github.com/MadAppGang/httplog"
	"github.com/justinas/alice"
	"github.com/justinas/nosurf"
)

func happyHandler() http.Handler {
	h := func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("I am happy!"))
	}
	return http.HandlerFunc(h)
}

func LoggerMiddleware(h http.Handler) http.Handler {
	return httplog.Logger(h)
}

func main() {
	// setup routes
	mux := http.NewServeMux()

	// create reusable middleware chain
	chain := alice.New(LoggerMiddleware, nosurf.NewPure)

	mux.Handle("/happy", chain.Then(happyHandler()))
	mux.Handle("/not_found", chain.Then(http.NotFoundHandler()))

	go func() {
		fmt.Println("Server started at port 3333")
		err := http.ListenAndServe(":3333", mux)
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
