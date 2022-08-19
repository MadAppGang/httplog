package httplog

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResponseBodyLogFormatterEmpty(t *testing.T) {
	bodyParams := LogFormatterParams{
		RouterName: "TEST",
		StatusCode: 200,
		ClientIP:   "20.20.20.20",
		Method:     "GET",
		Body:       []byte(""),
		Path:       "/",
	}

	assert.Equal(t,
		"===\n EMPTY BODY \n===\n",
		ResponseBodyLogFormatter(bodyParams),
	)
}

func TestResponseBodyLogFormatterText(t *testing.T) {
	bodyParams := LogFormatterParams{
		RouterName: "TEST",
		StatusCode: 200,
		ClientIP:   "20.20.20.20",
		Method:     "GET",
		Body:       []byte("I am text body!"),
		Path:       "/",
	}

	assert.Equal(t,
		"===\n TEXT BODY:\nI am text body!\n===\n",
		ResponseBodyLogFormatter(bodyParams),
	)
}

func TestResponseBodyLogFormatterJSON(t *testing.T) {
	bodyParams := LogFormatterParams{
		RouterName: "TEST",
		StatusCode: 200,
		ClientIP:   "20.20.20.20",
		Method:     "GET",
		Body:       []byte(`{"name":"John", "age":30, "car":null}`),
		Path:       "/",
	}

	assert.Equal(t,
		"===\n JSON BODY:\n{\n  \"age\": 30,\n  \"car\": null,\n  \"name\": \"John\"\n}\n===\n",
		ResponseBodyLogFormatter(bodyParams),
	)
}
