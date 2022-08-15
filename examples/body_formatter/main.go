package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/MadAppGang/httplog"
)

const jsonBody = `{"str": "foo","num": 100,"bool": false,"null": null,"array": ["foo", "bar", "baz"],"obj": { "a": 1, "b": 2 }}`

func happyHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(jsonBody))
}

func notJSONHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("I am not a JSON"))
}

func emptyHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(""))
}

// let's create logger
func logger(h http.HandlerFunc) http.Handler {
	return httplog.LoggerWithConfig(
		httplog.LoggerConfig{
			RouterName:  "FillBodyFormatter",
			Formatter:   httplog.DefaultLogFormatterWithHeadersAndBody,
			CaptureBody: true,
		},
		http.HandlerFunc(h),
	)
}

func main() {
	// setup routes
	http.Handle("/empty", logger(emptyHandler))
	http.Handle("/wrong", logger(notJSONHandler))
	http.Handle("/happy", logger(happyHandler))

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
	// empty body
	_, err := http.Get("http://localhost:3333/wrong")
	if err != nil {
		fmt.Printf("Error: %+v", err)
	}
	_, _ = http.Get("http://localhost:3333/empty")
	_, _ = http.Get("http://localhost:3333/happy")
	fmt.Println("All done, thank you and see you soon 👋")
}
