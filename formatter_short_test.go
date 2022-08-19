package httplog

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestShortLogFormatter(t *testing.T) {
	timeStamp := time.Unix(1544173902, 0).UTC()

	termFalseParam := LogFormatterParams{
		RouterName: "TEST",
		TimeStamp:  timeStamp,
		StatusCode: 200,
		Latency:    time.Second * 5,
		ClientIP:   "20.20.20.20",
		Method:     "GET",
		Path:       "/",
		isTerm:     false,
	}

	termTrueLongDurationParam := LogFormatterParams{
		RouterName: "TEST",
		TimeStamp:  timeStamp,
		StatusCode: 200,
		Latency:    time.Millisecond * 9876543210,
		ClientIP:   "20.20.20.20",
		Method:     "GET",
		Path:       "/",
		isTerm:     true,
	}

	assert.Equal(t, "[TEST]  200 | \"/\"\n", ShortLogFormatter(termFalseParam))
	assert.Equal(t, "[TEST] \x1b[97;42m 200 \x1b[0m| \"/\"\n", ShortLogFormatter(termTrueLongDurationParam))
}
