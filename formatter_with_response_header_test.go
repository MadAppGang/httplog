package httplog

import (
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResponseHeaderLogFormatter(t *testing.T) {
	rec := httptest.NewRecorder()
	rec.Header().Set("X-Real-IP", " 10.10.10.10  ")
	rec.Header().Set("X-Forwarded-For", "  20.20.20.20, 30.30.30.30")
	rec.Header().Set("Content-Type", "application/json")
	rec.Header().Set("Token", "Bearer ABCDEFG")
	textBodyParams := LogFormatterParams{
		ResponseHeader: rec.Header(),
	}
	result := ResponseHeaderLogFormatter(textBodyParams)
	assert.Contains(t, result, "X-Forwarded-For :  [  20.20.20.20, 30.30.30.30] ")
	assert.Contains(t, result, "X-Real-Ip :  [ 10.10.10.10  ] ")
	assert.Contains(t, result, "Content-Type :  [application/json] ")
	assert.Contains(t, result, "Token :  [Bearer ABCDEFG] ")
}

func TestResponseHeaderLogFormatterColor(t *testing.T) {
	rec := httptest.NewRecorder()
	rec.Header().Set("X-Real-IP", " 10.10.10.10  ")
	rec.Header().Set("X-Forwarded-For", "  20.20.20.20, 30.30.30.30")
	rec.Header().Set("Content-Type", "application/json")
	rec.Header().Set("Token", "Bearer ABCDEFG")
	textBodyParams := LogFormatterParams{
		ResponseHeader: rec.Header(),
		isTerm:         true,
	}
	result := ResponseHeaderLogFormatter(textBodyParams)
	assert.Contains(t, result, "  \x1b[1;34m X-Real-Ip \x1b[0m: \x1b[;32m [ 10.10.10.10  ] \x1b[0m")
	assert.Contains(t, result, "\x1b[1;34m X-Forwarded-For \x1b[0m: \x1b[;32m [  20.20.20.20, 30.30.30.30] \x1b[0m")
	assert.Contains(t, result, "\x1b[1;34m Content-Type \x1b[0m: \x1b[;32m [application/json] \x1b[0m")
	assert.Contains(t, result, "  \x1b[1;34m Token \x1b[0m: \x1b[;32m [Bearer ABCDEFG] \x1b[0m\n")
}
