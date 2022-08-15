# Beautiful logger for http

Proudly created and supported by [MadAppGang](https://madappgang.com) company.

## Why?

Every single web framework has a build-in logger already, why do we need on more?
The question is simple and the answer is not.

Nice and clean output is critical for any web framework. Than is why some people use go web frameworks just to get beautiful logs.

This library brings you fantastic http logs to any web framework, even if you use native `net/http` for that.

But it's better to see once, here the default output you will get with couple of lines of code:

![logs screenshot](docs/logs_screenshot.png)

And actual code looks like this:

```go
  func main() {
    // setup routes
    http.Handle("/happy", httplog.Logger(happyHandler))
    http.Handle("/not_found", httplog.Logger(http.NotFoundHandler()))

    //run server
    _ = http.ListenAndServe(":3333", nil)
  }
```

All you need is wrap you handler with `httplog.Logger` and the magic happens.

Here is a main features:

- framework agnostic (could be easily integrated with any web framework), you can find `examples` for:
  - [alice](https://github.com/MadAppGang/httplog/blob/main/examples/alice/main.go)
  - [chi](https://github.com/MadAppGang/httplog/blob/main/examples/chi/main.go)
  - [echo](https://github.com/MadAppGang/httplog/blob/main/examples/echo/main.go)
  - [gin](https://github.com/MadAppGang/httplog/blob/main/examples/gin/main.go)
  - [goji](https://github.com/MadAppGang/httplog/blob/main/examples/goji/main.go)
  - [gorilla mux](https://github.com/MadAppGang/httplog/blob/main/examples/gorilla/main.go)
  - [httprouter](https://github.com/MadAppGang/httplog/blob/main/examples/httprouter/main.go)
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

## Custom format

You can modify formatter as you want. Now there are two formatter available: 

- `DefaultLogFormatter`
- `ShortLogFormatter`
- `HeadersLogFormatter`
- `DefaultLogFormatterWithHeaders`
- `BodyLogFormatter`
- `DefaultLogFormatterWithHeadersAndBody`

And you can combine them using `ChainLogFormatter`.

Here is an example of formatter in code:

```go
// Short log formatter
shortLoggedHandler := httplog.LoggerWithFormatter(
  httplog.ShortLogFormatter,
  wrappedHandler,
)    
```

You can define your own log format. Log formatter is a function with a set of precalculated parameters:

```go
// Custom log formatter
customLoggedHandler := httplog.LoggerWithFormatter(
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
  happyHandler,
)
http.Handle("/happy_custom", customLoggedHandler)
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
| Body | Body is a body content, if body is copied |

## Integrate with structure logger

Good and nice output is good, but as soon as we have so much data about every response it is a good idea to pass it to our application log structured collector.

One of the most popular solution is [Uber zap](https://github.com/uber-go/zap).
You can use any structured logger you want, use zap's integration example as a reference.

All you need is create custom log formatter function with your logger integration.
This repository has this formatter for zap created and you can use it importing `github.com/MadAppGang/httplog/zap`:

```go
logger := httplog.LoggerWithConfig(
  httplog.LoggerConfig{
    Formatter:  lzap.DefaultZapLogger(zapLogger, zap.InfoLevel, ""),
  },
  http.HandlerFunc(handler),
)
http.Handle("/happy", logger)
```

You can find full-featured [example in zap integration folder](https://github.com/MadAppGang/httplog/blob/main/examples/zap/main.go).

## Customize log output destination

TODO: Jack

## Use GoogleApp Engine or CloudFlare

TODO: Jack

## Run custom logic before a response has been written

TODO: Jack

## How to save request body and headers


You can capture response data as well. But please use it in dev environments only, as it use extra resources and produce a lot of output in terminal. Example of body output [could be found here](https://github.com/MadAppGang/httplog/blob/main/examples/body_formatter/main.go).

![body output](docs/full_body_formatter.png)

You can use `DefaultLogFormatterWithHeaders` for headers output or `DefaultLogFormatterWithHeadersAndBody` to output response body. Don't forget to set `CaptureBody` in LoggerParams.

You can combine your custom Formatter and `HeadersLogFormatter` or/and `BodyLogFormatter` using `ChainLogFormatter`:

```go
var myFormatter = httplog.ChainLogFormatter(
  MyLogFormatter, 
  httplog.HeadersLogFormatter, 
  httplog.BodyLogFormatter,
)
```

## Integration examples

Please go to examples folder and see how it's work:

![Run demo](docs/demo_run.gif)

### Native `net/http` package

TODO: Jack

### Alice middleware

TODO: Jack

### Chi

TODO: Jack

### Echo

TODO: Jack

### Gin

TODO: Jack

### Goji

TODO: Jack

### Gorilla

TODO: Jack

### HTTPRouter

TODO: Jack

### Negroni

TODO: Jack
