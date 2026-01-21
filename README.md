# Beautiful logger for http

Proudly created and supported by [MadAppGang](https://madappgang.com) company.

[![Go Reference](https://pkg.go.dev/badge/github.com/MadAppGang/httplog/v2.svg)](https://pkg.go.dev/github.com/MadAppGang/httplog/v2)
[![Go Report Card](https://goreportcard.com/badge/github.com/MadAppGang/httplog/v2)](https://goreportcard.com/report/github.com/MadAppGang/httplog/v2)
[![codecov](https://codecov.io/gh/MadAppGang/httplog/branch/main/graph/badge.svg?token=BJT0WRNJS1)](https://codecov.io/gh/MadAppGang/httplog)

## 🚀 v2.0.0 Released!

This is a major release with breaking changes. See the [Migration Guide](#migrating-from-v1-to-v2) below.

**Install:**
```bash
go get github.com/MadAppGang/httplog/v2@latest
```

## Why?

Every single web framework has a build-in logger already, why do we need on more?
The question is simple and the answer is not.

The best logger is structured logger, like [Uber Zap](https://github.com/uber-go/zap). Structure logs are essentials today to help collect and analyze logs with monitoring tools like ElasticSearch or DataDog.

But what about humans? Obviously you can go to your monitoring tool and parse your structured logs. But what about seeing human readable logs in place just in console? Although you can read JSON, it is extremely hard to find most critical information in the large JSON data flow.

Httplog package brings good for humans and log collection systems, you can use tools like zap to create structured logs and see colored beautiful human-friendly output in console in place. Win-win for human and machines 🤖❤️👩🏽‍💻

Nice and clean output is critical for any web framework. Than is why some people use go web frameworks just to get beautiful logs.

This library brings you fantastic http logs to any web framework, even if you use native `net/http` for that.

But it's better to see once, here the default output you will get with couple of lines of code:

![logs screenshot](docs/logs_screenshot.png)

And actual code looks like this:

```go
func main() {
  logger, err := httplog.Logger(happyHandler)
  if err != nil {
    panic(err)
  }
  http.Handle("/happy", logger)

  notFoundLogger, _ := httplog.Logger(http.NotFoundHandler())
  http.Handle("/not_found", notFoundLogger)

  _ = http.ListenAndServe(":3333", nil)
}
```


All you need is wrap you handler with `httplog.Logger` and the magic happens.

And this is how the structured logs looks in AWS CloudWatch. As you can see, the color output looks not as good here as in console. But JSON structured logs are awesome 😍

![structured logs](docs/structured_logs.com.png)

Here is a main features:

- framework agnostic (could be easily integrated with any web framework), you can find `examples` for:
  - [alice](https://github.com/MadAppGang/httplog/blob/main/examples/alice/main.go)
  - [chi](https://github.com/MadAppGang/httplog/blob/main/examples/chi/main.go)
  - [echo](https://github.com/MadAppGang/httplog/blob/main/examples/echo/main.go)
  - [gin](https://github.com/MadAppGang/httplog/blob/main/examples/gin/main.go)
  - [goji](https://github.com/MadAppGang/httplog/blob/main/examples/goji/main.go)
  - [gorilla mux](https://github.com/MadAppGang/httplog/blob/main/examples/gorilla/main.go)
  - [httprouter](https://github.com/MadAppGang/httplog/blob/main/examples/httprouter/main.go)
  - [mojito](https://github.com/MadAppGang/httplog/blob/main/examples/mojito/main.go)
  - [negroni](https://github.com/MadAppGang/httplog/blob/main/examples/negroni/main.go)
  - [native net/http](https://github.com/MadAppGang/httplog/blob/main/examples/nethttp/main.go)
  - not found yours? let us know and we will add it
- response code using special wrapper
- response length using special wrapper
- can copy response body
- get real user IP for Google App Engine
- get real user IP for CloudFront
- get real user IP for other reverse proxy which implements [RFC7239](https://www.rfc-editor.org/rfc/rfc7239.html)
- customize output format
- has the list of routes to ignore
- build in structure logger integration
- callback function to modify response before write back (add headers or do something)

This framework is highly inspired by [Gin logger](https://github.com/gin-gonic/gin/blob/master/logger.go) library, but has not Gin dependencies at all and has some improvements.
Httplog has only one dependency at all: `github.com/mattn/go-isatty`. So it's will not affect your codebase size.

## New in v2.0.0

### Request Sampling
Log only a percentage of requests to reduce log volume:
```go
logger, _ := httplog.LoggerWithConfig(httplog.LoggerConfig{
    SampleRate: 0.1, // Log 10% of requests
    DeterministicSampling: true, // Same request always sampled the same way
})
```

### Log Level Filtering
Filter logs by HTTP status code:
```go
logger, _ := httplog.LoggerWithConfig(httplog.LoggerConfig{
    MinLevel: httplog.LevelWarn, // Only log 4xx and 5xx responses
})
```

### Async Logging
Reduce request latency with async log writing:
```go
logger, _ := httplog.LoggerWithConfig(httplog.LoggerConfig{
    AsyncLogging: true,
    AsyncBufferSize: 5000,
})
defer logger.Close() // Important! Cleanly shuts down the async goroutine
```

### Request Body Capture
```go
logger, _ := httplog.LoggerWithConfig(httplog.LoggerConfig{
    CaptureRequestBody: true,
})
```

### Color Mode Control
```go
logger, _ := httplog.LoggerWithConfig(httplog.LoggerConfig{
    ColorMode: httplog.ColorForce, // ColorAuto, ColorDisable, ColorForce
})
```

### ConfigBuilder Fluent API
```go
config := httplog.NewConfigBuilder().
    WithName("API").
    WithColorMode(httplog.ColorForce).
    WithSampleRate(0.5).
    WithAsyncLogging(true, 1000).
    Build()

logger, _ := httplog.LoggerWithConfig(config)
```

## Custom format

You can modify formatter as you want. Now there are two formatter available:

- `DefaultLogFormatter`
- `ShortLogFormatter`
- `HeadersLogFormatter`
- `DefaultLogFormatterWithHeaders`
- `BodyLogFormatter`
- `DefaultLogFormatterWithHeadersAndBody`
- `RequestHeaderLogFormatter`
- `DefaultLogFormatterWithRequestHeader`
- `RequestBodyLogFormatter`
- `DefaultLogFormatterWithRequestHeadersAndBody`
- `FullFormatterWithRequestAndResponseHeadersAndBody`

And you can combine them using `ChainLogFormatter`.

Here is an example of formatter in code:

```go
// Short log formatter
shortLoggedHandler, _ := httplog.LoggerWithFormatter(
  httplog.ShortLogFormatter,
)
http.Handle("/short", shortLoggedHandler.Handler(wrappedHandler))
```

You can define your own log format. Log formatter is a function with a set of precalculated parameters:

```go
// Custom log formatter
customLoggedHandler, _ := httplog.LoggerWithFormatter(
  // formatter is a function, you can define your own
  func(param httplog.LogFormatterParams) string {
    statusColor := param.StatusCodeColor()
    resetColor := param.ResetColor()
    boldRedText := "\033[1;31m"

    return fmt.Sprintf("🥑[I am custom router!!!] %s %3d %s| size: %10d bytes | %s %#v %s 🔮👨🏻‍💻\n",
      statusColor, param.StatusCode, resetColor,
      param.BodySize,
      boldRedText, param.Path, resetColor,
    )
  },
)
http.Handle("/happy_custom", customLoggedHandler.Handler(happyHandler))
```

For more details and how to capture response body please look in the [example app](https://github.com/MadAppGang/httplog/blob/main/examples/custom_formatter/main.go).

params is a type of LogFormatterParams and the following params available for you:

| param | description |
| --- | --- |
| Request | `http.Request` instance |
| RouterName | when you create logger, you can specify router name |
| Timestamp | TimeStamp shows the time after the server returns a response |
| StatusCode | StatusCode is HTTP response code |
| Latency | Latency is how much time the server cost to process a certain request |
| ClientIP | ClientIP calculated real IP of requester, see Proxy for details |
| Method | Method is the HTTP method given to the request |
| Path | Path is a path the client requests |
| BodySize | BodySize is the size of the Response Body |
| Context | context.Context from the request |
| RequestBody | Request body content (if CaptureRequestBody enabled) |
| RequestHeader | Request headers (with masking applied) |
| ResponseBody | Response body content (if CaptureResponseBody enabled) |
| Level | Log level (Debug, Info, Warn, Error) |

## Integrate with structure logger

Good and nice output is good, but as soon as we have so much data about every response it is a good idea to pass it to our application log structured collector.

One of the most popular solution is [Uber zap](https://github.com/uber-go/zap).
You can use any structured logger you want, use zap's integration example as a reference.

All you need is create custom log formatter function with your logger integration.
This repository has this formatter for zap created and you can use it importing `github.com/MadAppGang/httplog/zap`:

```go
logger, err := httplog.LoggerWithConfig(httplog.LoggerConfig{
    Formatter: lzap.DefaultZapLogger(zapLogger, zap.InfoLevel, ""),
})
if err != nil {
    panic(err)
}
http.Handle("/happy", logger.Handler(handler))
```

You can find full-featured [example in zap integration folder](https://github.com/MadAppGang/httplog/blob/main/examples/zap/main.go).

## Customize log output destination

You can use any output you need, your output must support `io.Writer` protocol.
After that you need init logger:

```go
buffer := new(bytes.Buffer)
logger, _ := LoggerWithWriter(buffer) //all output is written to buffer
http.Handle("/", logger.Handler(handler))
```

## Care about secrets. Skipping path and masking headers

Some destinations should not be logged. For that purpose logger config has `SkipPaths` property with array of strings. Each string is a Regexp for path you want to skip. You can write exact path to skip, which would be valid regexp, or you can use regexp power:

```go
logger, _ := LoggerWithConfig(LoggerConfig{
  SkipPaths: []string{
    "/skipped",
    "/payments/\\w+",
    "/user/[0-9]+",
  },
})
http.Handle("/", logger.Handler(handler))


The other feature to safe your logs from leaking secrets is Header masking. For example you do not want to log Bearer token header, but it is  useful to see it is present and not empty.

For this purpose `LoggerConfig` has field `HideHeaderKeys` which works the same as `SkipPaths`. Just feed an array of case insensitive key names regexps like that:

```go
logger, _ := LoggerWithConfig(LoggerConfig{
  HideHeaderKeys: []string{
    "Bearer",
    "Secret-Key",
    "Cookie",
  },
})
http.Handle("/", logger.Handler(handler))

```

If regexp is failed to compile, the logger will skip and and it will write the message to the log output destination.

## Use GoogleApp Engine or CloudFlare

You application is operating behind load balancers and reverse proxies. That is why origination IP address is changing on every hop.
To save the first sender's IP (real user remote IP) reverse proxies should save original IP in request headers.
Default headers are `X-Forwarded-For` and `X-Real-IP`.

But some clouds have custom headers, like Cloudflare and Google Apps Engine.
If you are using those clouds or have custom headers in you environment, you can handle that by using custom `Proxy` init parameters:

```go
logger, _ := httplog.LoggerWithConfig(httplog.LoggerConfig{
  ProxyHandler: NewProxyWithType(httplog.ProxyGoogleAppEngine),
})
http.Handle("/", logger.Handler(h))
```

or if you have your custom headers:

```go
logger, _ := httplog.LoggerWithConfig(httplog.LoggerConfig{
  ProxyHandler: NewProxyWithTypeAndHeaders(
    httplog.ProxyDefaultType,
    []string{"My-HEADER-NAME", "OTHER-HEADER-NAME"}
  ),
})
http.Handle("/", logger.Handler(handler))
```

## How to save request body and headers

You can capture response data as well. But please use it in dev environments only, as it use extra resources and produce a lot of output in terminal. Example of body output [could be found here](https://github.com/MadAppGang/httplog/blob/main/examples/body_formatter/main.go).

![body output](docs/full_body_formatter.png)

You can use `DefaultLogFormatterWithHeaders` for headers output or `DefaultLogFormatterWithHeadersAndBody` to output response body. Don't forget to set `CaptureResponseBody` in LoggerParams.

You can combine your custom Formatter and `HeadersLogFormatter` or/and `BodyLogFormatter` using `ChainLogFormatter`:

```go
var myFormatter = httplog.ChainLogFormatter(
  MyLogFormatter,
  httplog.HeadersLogFormatter,
  httplog.BodyLogFormatter,
)
```

## Migrating from v1 to v2

### Import Path
```go
// Old
import "github.com/MadAppGang/httplog"

// New
import "github.com/MadAppGang/httplog/v2"
```

### Error Handling
All factory functions now return errors:
```go
// Old
logger := httplog.LoggerWithName("API")
http.Handle("/", logger(handler))

// New
logger, err := httplog.LoggerWithName("API")
if err != nil {
    panic(err)
}
http.Handle("/", logger.Handler(handler))
```

### LoggingMiddleware Type
The middleware is now wrapped in a struct with a `Close()` method:
```go
logger, _ := httplog.LoggerWithName("API")
defer logger.Close() // Important for async logging!

// Use logger.Handler instead of calling logger directly
http.Handle("/", logger.Handler(handler))
```

### Removed Global Functions
```go
// Old - these are removed
httplog.DisableConsoleColor()
httplog.ForceConsoleColor()

// New - use ColorMode in config
logger, _ := httplog.LoggerWithConfig(httplog.LoggerConfig{
    ColorMode: httplog.ColorDisable,
})
```

### Renamed Fields
| Old | New |
|-----|-----|
| `CaptureBody` | `CaptureResponseBody` |
| `LogFormatterParams.Body` | `LogFormatterParams.ResponseBody` |

## Integration examples

Please go to examples folder and see how it's work:

![Run demo](docs/demo_run.gif)

### Native `net/http` package

This package is suing canonical golang approach, and it easy to implement it with all `net/http` package.

```go
logger, _ := httplog.Logger(http.NotFoundHandler())
http.Handle("/not_found", logger)
```

Full example [could be found here](https://github.com/MadAppGang/httplog/blob/main/examples/nethttp/main.go).

### Alice middleware

Alice is a fantastic lightweight framework to chain and manage middlewares. As Alice is using canonical approach, it is working with httplog out-of-the-box.

You don't need any wrappers and you can user logger directly:

```golang
logger, _ := httplog.LoggerWithName("ALICE")
chain := alice.New(logger.Handler, nosurf.NewPure)
mux.Handle("/happy", chain.Then(happyHandler()))
```

Full example [could be found here](https://github.com/MadAppGang/httplog/blob/main/examples/alice/main.go).

### Chi

Chi is a router which uses the standard approach as `Alice` package.

You don't need any wrappers and you can user logger directly:

```go
r := chi.NewRouter()
logger, _ := httplog.LoggerWithName("CHI")
r.Use(logger.Handler)
r.Get("/happy", happyHandler)
```

Full example [could be found here](https://github.com/MadAppGang/httplog/blob/main/examples/chi/main.go).

### Echo

Echo middleware uses internal `echo.Context` to manage request/responser flow and middleware management.

As a result, `echo` creates it's own http.ResponseWriter wrapper to catch all written data.
That is why `httplog` could not handle it automatically.

To handle that there is a package `httplog/chilog` which has all the native echo middlewares to use:


```go
	e := echo.New()

	// Middleware
	e.Use(echolog.LoggerWithName("ECHO NATIVE"))
	e.GET("/happy", happyHandler)
	e.POST("/happy", happyHandler)
	e.GET("/not_found", echo.NotFoundHandler)

```

Full example [could be found here](https://github.com/MadAppGang/httplog/blob/main/examples/echo/main.go).

### Gin

Gin has the most beautiful log output. That is why this package has build highly inspired by `Gin's` logger.

If you missing some features of Gin's native logger and want to use this one, it is still possible.

Gin middleware uses internal `gin.Context` to manage request/responser flow and middleware management. The same approach as `Echo`.

That is why we created `httplog`, to bring this fantastic logger to native golang world :-)

`Gin` creates it's own http.ResponseWriter wrapper to catch all written data.
That is why `httplog` could not handle it automatically.

To handle that there is a package `httplog/ginlog` which has all the native gin middlewares to use:


```go
  r := gin.New()
	r.Use(ginlog.LoggerWithName("I AM GIN ROUTER"))
	r.GET("/happy", happyHandler)
	r.POST("/happy", happyHandler)
	r.GET("/not_found", gin.WrapF(http.NotFound))
```

Full example [could be found here](https://github.com/MadAppGang/httplog/blob/main/examples/gin/main.go).

### Goji

Goji is using canonical middleware approach, that is why it is working with httplog out-of-the-box.

You don't need wrappers or anything like that.

```go
mux := goji.NewMux()
logger, _ := httplog.Logger(happyHandler)
mux.Handle(pat.Get("/happy"), logger)
```

Full example [could be found here](https://github.com/MadAppGang/httplog/blob/main/examples/goji/main.go).

### Gorilla

Gorilla mux is one of the most loved muxer/router.
Gorilla is using canonical middleware approach, that is why it is working with httplog out-of-the-box.

You don't need to create any wrappers:

```go
r := mux.NewRouter()
r.HandleFunc("/happy", happyHandler)
logger, _ := httplog.LoggerWithName("GORILLA")
r.Use(logger.Handler)
```

Full example [could be found here](https://github.com/MadAppGang/httplog/blob/main/examples/gorilla/main.go).

### HTTPRouter

To use HTPPRouter you need to create simple wrapper. As HTTPRouter using custom `httprouter.Handler` and additional argument with params in handler function.

```go
func LoggerMiddleware(h httprouter.Handle) httprouter.Handle {
	logger, _ := httplog.LoggerWithName("HTTPROUTER")
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h(w, r, ps)
		})
		logger.Handler(handler).ServeHTTP(w, r)
	}
}

//use as native middleware
router := httprouter.New()
router.GET("/happy", LoggerMiddleware(happyHandler))
```

Full example [could be found here](https://github.com/MadAppGang/httplog/blob/main/examples/httprouter/main.go).

### Mojito
Because Go-Mojito uses dynamic handler functions, which include support for net/http types, httplog works out-of-the-box:  

```go
logger, _ := httplog.LoggerWithName("MOJITO")
mojito.WithMiddleware(logger.Handler)
mojito.GET("/happy", happyHandler)
```
Full example [could be found here](https://github.com/MadAppGang/httplog/blob/main/examples/mojito/main.go).

### Negroni

Negroni uses custom `negroni.Handler` as middleware. We need to create custom wrapper for that:

```go
var negroniLoggerMiddleware negroni.Handler = negroni.HandlerFunc(func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
  logger, _ := httplog.Logger(next)
  logger.ServeHTTP(rw, r)
})

//use it as native middleware:
n := negroni.New()
n.Use(negroniLoggerMiddleware)
n.UseHandler(mux)
```

Full example [could be found here](https://github.com/MadAppGang/httplog/blob/main/examples/negroni/main.go).
