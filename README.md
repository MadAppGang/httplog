# Beautiful logger for http

## Why?

Every single web framework has a build-in logger already, why do we need on more?
The question is simple and the answer is not.

Nice and clean output is critical for any web framework. Than is why come people use go web frameworks just because to get beautiful logs.

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
  - alice
  - chi
  - echo
  - gin
  - goji
  - gorilla mux
  - httprouter
  - negroni
  - native net/http
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

TODO: Jack

## Integrate with structure logger

TODO: Jack

## Customize log output

TODO: Jack

## Use GoogleApp Engine or CloudFlare

TODO: Jack

## Run custom logic before a response has been written

TODO: Jack

## How to save request body and headers

TODO: Jack

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
