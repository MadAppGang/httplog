package httplog

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResponseBodyLogFormatterEmpty(t *testing.T) {
	bodyParams := LogFormatterParams{
		RouterName:   "TEST",
		StatusCode:   200,
		ClientIP:     "20.20.20.20",
		Method:       "GET",
		ResponseBody: []byte(""),
		Path:         "/",
		colorMode:    ColorDisable,
	}

	assert.Equal(t,
		"===\n EMPTY BODY \n===\n",
		ResponseBodyLogFormatter(bodyParams),
	)
}

func TestResponseBodyLogFormatterText(t *testing.T) {
	bodyParams := LogFormatterParams{
		RouterName:   "TEST",
		StatusCode:   200,
		ClientIP:     "20.20.20.20",
		Method:       "GET",
		ResponseBody: []byte("I am text body!"),
		Path:         "/",
		colorMode:    ColorDisable,
	}

	assert.Equal(t,
		"===\n TEXT BODY:\nI am text body!\n===\n",
		ResponseBodyLogFormatter(bodyParams),
	)
}

func TestResponseBodyLogFormatterTextColor(t *testing.T) {
	bodyParams := LogFormatterParams{
		RouterName:   "TEST",
		StatusCode:   200,
		ClientIP:     "20.20.20.20",
		Method:       "GET",
		ResponseBody: []byte("I am text body!"),
		Path:         "/",
		colorMode:    ColorForce,
	}

	assert.Equal(t,
		"===\n\x1b[97;44m TEXT BODY:\x1b[0m\nI am text body!\n===\n",
		ResponseBodyLogFormatter(bodyParams),
	)
}

func TestResponseBodyLogFormatterJSON(t *testing.T) {
	bodyParams := LogFormatterParams{
		RouterName:   "TEST",
		StatusCode:   200,
		ClientIP:     "20.20.20.20",
		Method:       "GET",
		ResponseBody: []byte(`{"name":"John", "age":30, "car":null}`),
		Path:         "/",
		colorMode:    ColorDisable,
	}

	assert.Equal(t,
		"===\n JSON BODY:\n{\n  \"age\": 30,\n  \"car\": null,\n  \"name\": \"John\"\n}\n===\n",
		ResponseBodyLogFormatter(bodyParams),
	)
}
