package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/MadAppGang/httplog"
	lzap "github.com/MadAppGang/httplog/zap"
	"go.uber.org/zap"
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
func logger(h http.HandlerFunc, l *zap.Logger) http.Handler {
	return httplog.LoggerWithConfig(
		httplog.LoggerConfig{
			RouterName: "FillBodyFormatter",
			Formatter:  lzap.DefaultZapLogger(l, zap.InfoLevel, ""),
		},
		http.HandlerFunc(h),
	)
}

func main() {
	// setup routes
	z, _ := zap.NewDevelopment()
	defer z.Sync() // flushes buffer, if any

	http.Handle("/happy", logger(happyHandler, z))

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
	_, err := http.Get("http://localhost:3333/happy")
	if err != nil {
		fmt.Printf("Error: %+v", err)
	}
	fmt.Println("All done, thank you and see you soon 👋")
}
