package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/MadAppGang/httplog/v2"
)

var happyHandler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Happy handler log.")
	w.Write([]byte("I am happy!"))
})

func main() {
	// setup routes
	// Default log formatter
	defaultLogger, err := httplog.LoggerWithName("Default logger")
	if err != nil {
		panic(err)
	}
	http.Handle("/happy", defaultLogger.Handler(happyHandler))

	// Short log formatter
	shortLogged, err := httplog.LoggerWithFormatterAndName(
		"Short logger",
		httplog.ShortLogFormatter, // log formatter implemented in httplog lib
	)
	if err != nil {
		panic(err)
	}
	http.Handle("/happy_short", shortLogged.Handler(happyHandler))

	// Custom log formatter
	customLogger, err := httplog.LoggerWithFormatter(func(param httplog.LogFormatterParams) string {
		// formatter is a function, you can define your own
		statusColor := param.StatusCodeColor()
		resetColor := param.ResetColor()
		boldRedText := "\033[1;31m"

		return fmt.Sprintf("🥑[I am custom router!!!] %s %3d %s| size: %10d bytes | %s %#v %s 🔮👨🏻‍💻\n",
			statusColor, param.StatusCode, resetColor,
			param.BodySize,
			boldRedText, param.Path, resetColor,
		)
	})
	if err != nil {
		panic(err)
	}
	http.Handle("/happy_custom", customLogger.Handler(happyHandler))

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
	_, err = http.Get("http://localhost:3333/happy")
	if err != nil {
		fmt.Printf("Error: %+v", err)
	}
	_, _ = http.Get("http://localhost:3333/happy_short")
	_, _ = http.Get("http://localhost:3333/happy_custom")
	fmt.Println("All done, thank you and see you soon 👋")
}
