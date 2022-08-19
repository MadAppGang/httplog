package httplog

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type brokenBody int

func (brokenBody) Read(p []byte) (n int, err error) {
	return 0, errors.New("I am always broken, no worries at all :-)")
}

func TestRequestBodyLogFormatterBroken(t *testing.T) {
	timeStamp := time.Unix(1544173902, 0).UTC()

	request, _ := http.NewRequestWithContext(context.Background(), "POST", "/", brokenBody(1))
	brokenBodyParams := LogFormatterParams{
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
	assert.Equal(t,
		"===\n ERROR READING BODY: I am always broken, no worries at all :-) \n===\n",
		RequestBodyLogFormatter(brokenBodyParams),
	)
}

func TestRequestBodyLogFormatterEmpty(t *testing.T) {
	timeStamp := time.Unix(1544173902, 0).UTC()

	request, _ := http.NewRequestWithContext(context.Background(), "POST", "/", nil)
	emptyBodyParams := LogFormatterParams{
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
	assert.Equal(t,
		"===\n EMPTY BODY \n===\n",
		RequestBodyLogFormatter(emptyBodyParams),
	)
}

func TestRequestBodyLogFormatterText(t *testing.T) {
	timeStamp := time.Unix(1544173902, 0).UTC()

	request, _ := http.NewRequestWithContext(context.Background(), "POST", "/", bytes.NewBufferString("I am just a text body!"))
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
	assert.Equal(t,
		"===\n TEXT BODY:\nI am just a text body!\n===\n",
		RequestBodyLogFormatter(textBodyParams),
	)
}

func TestRequestBodyLogFormatterJSON(t *testing.T) {
	timeStamp := time.Unix(1544173902, 0).UTC()

	jsonbody := `{"name":"John", "age":30, "car":null}`

	request, _ := http.NewRequestWithContext(context.Background(), "POST", "/", bytes.NewBufferString(jsonbody))
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
	assert.Equal(t,
		"===\n JSON BODY:\n{\n  \"age\": 30,\n  \"car\": null,\n  \"name\": \"John\"\n}\n===\n",
		RequestBodyLogFormatter(textBodyParams),
	)
}
