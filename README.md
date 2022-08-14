# Beautiful logger for http

## Why?

Every single web framework has a build-in logger already, why do we need on more?
The question is simple and the answer is not.

Nice and clean output is critical for any web framework. Than is why come people use go web frameworks just because to get beautiful logs.

This library brings you fantastic http logs to any web framework, even if you use native `net/http` for that.

But it's better to see once, here the default output you will get with couple of lines of code:
![logs screenshot](docs/logs_screenshot.png)

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

## Examples

Please go to examples folder and see how it's work:

![Run demo](docs/demo_run.gif)

