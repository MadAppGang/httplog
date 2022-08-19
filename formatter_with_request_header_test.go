package httplog

import (
	"bytes"
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func setHeader(r *http.Request) {
	r.Header.Set("X-Real-IP", " 10.10.10.10  ")
	r.Header.Set("X-Forwarded-For", "  20.20.20.20, 30.30.30.30")
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Token", "Bearer ABCDEFG")
	r.RemoteAddr = "  40.40.40.40:42123 "
}

func TestRequestHeaderLogFormatter(t *testing.T) {
	timeStamp := time.Unix(1544173902, 0).UTC()
	request, _ := http.NewRequestWithContext(context.Background(), "POST", "/", bytes.NewBufferString("I am just a text body!"))
	setHeader(request)
	textBodyParams := LogFormatterParams{
		Request:    request,
		RouterName: "TEST",
		TimeStamp:  timeStamp,
		StatusCode: 200,
		Latency:    time.Second * 5,
		ClientIP:   "20.20.20.20",
		Method:     "GET",
		Path:       "/",
		isTerm:     false,
	}
	result := RequestHeaderLogFormatter(textBodyParams)
	assert.Contains(t, result, "X-Forwarded-For :  [  20.20.20.20, 30.30.30.30] ")
	assert.Contains(t, result, "X-Real-Ip :  [ 10.10.10.10  ] ")
	assert.Contains(t, result, "Content-Type :  [application/json] ")
	assert.Contains(t, result, "Token :  [Bearer ABCDEFG] ")
}

func TestRequestHeaderLogFormatterColor(t *testing.T) {
	timeStamp := time.Unix(1544173902, 0).UTC()
	request, _ := http.NewRequestWithContext(context.Background(), "POST", "/", bytes.NewBufferString("I am just a text body!"))
	setHeader(request)
	textBodyParams := LogFormatterParams{
		Request:    request,
		RouterName: "TEST",
		TimeStamp:  timeStamp,
		StatusCode: 200,
		Latency:    time.Second * 5,
		ClientIP:   "20.20.20.20",
		Method:     "GET",
		Path:       "/",
		isTerm:     true,
	}

	result := RequestHeaderLogFormatter(textBodyParams)
	assert.Contains(t, result, "  \x1b[1;34m X-Real-Ip \x1b[0m: \x1b[;32m [ 10.10.10.10  ] \x1b[0m")
	assert.Contains(t, result, "\x1b[1;34m X-Forwarded-For \x1b[0m: \x1b[;32m [  20.20.20.20, 30.30.30.30] \x1b[0m")
	assert.Contains(t, result, "\x1b[1;34m Content-Type \x1b[0m: \x1b[;32m [application/json] \x1b[0m")
	assert.Contains(t, result, "  \x1b[1;34m Token \x1b[0m: \x1b[;32m [Bearer ABCDEFG] \x1b[0m\n")
}
