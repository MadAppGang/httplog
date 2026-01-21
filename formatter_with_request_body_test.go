package httplog

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRequestBodyLogFormatterBroken(t *testing.T) {
	// This test is no longer relevant since we don't read from Request.Body
	// RequestBodyLogFormatter now uses the RequestBody field directly
	// which is populated by the middleware
	t.Skip("Not applicable - RequestBodyLogFormatter now uses RequestBody field")
}

func TestRequestBodyLogFormatterEmpty(t *testing.T) {
	emptyBodyParams := LogFormatterParams{
		RequestBody: []byte{},
		colorMode:   ColorDisable,
	}
	assert.Equal(t,
		"===\n EMPTY BODY \n===\n",
		RequestBodyLogFormatter(emptyBodyParams),
	)
}

func TestRequestBodyLogFormatterEmptyColor(t *testing.T) {
	emptyBodyParams := LogFormatterParams{
		RequestBody: []byte{},
		colorMode:   ColorForce,
	}
	assert.Equal(t,
		"===\n\x1b[90;43m EMPTY BODY \x1b[0m\n===\n",
		RequestBodyLogFormatter(emptyBodyParams),
	)
}

func TestRequestBodyLogFormatterText(t *testing.T) {
	textBodyParams := LogFormatterParams{
		RequestBody: []byte("I am just a text body!"),
		colorMode:   ColorDisable,
	}
	assert.Equal(t,
		"===\n TEXT BODY:\nI am just a text body!\n===\n",
		RequestBodyLogFormatter(textBodyParams),
	)
}

func TestRequestBodyLogFormatterJSON(t *testing.T) {
	jsonbody := `{"name":"John", "age":30, "car":null}`

	textBodyParams := LogFormatterParams{
		RequestBody: []byte(jsonbody),
		colorMode:   ColorDisable,
	}
	assert.Equal(t,
		"===\n JSON BODY:\n{\n  \"age\": 30,\n  \"car\": null,\n  \"name\": \"John\"\n}\n===\n",
		RequestBodyLogFormatter(textBodyParams),
	)
}
