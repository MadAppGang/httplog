package httplog

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDefaultLogFormatter(t *testing.T) {
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

	termTrueParam := LogFormatterParams{
		RouterName: "TEST",
		TimeStamp:  timeStamp,
		StatusCode: 200,
		Latency:    time.Second * 5,
		ClientIP:   "20.20.20.20",
		Method:     "GET",
		Path:       "/",
		isTerm:     true,
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

	termFalseLongDurationParam := LogFormatterParams{
		RouterName: "TEST",
		TimeStamp:  timeStamp,
		StatusCode: 200,
		Latency:    time.Millisecond * 9876543210,
		ClientIP:   "20.20.20.20",
		Method:     "GET",
		Path:       "/",
		isTerm:     false,
	}

	assert.Equal(t, "[TEST] 2018/12/07 - 09:11:42 | 200 |            5s |     20.20.20.20 | GET      \"/\"\n", DefaultLogFormatter(termFalseParam))
	assert.Equal(t, "[TEST] 2018/12/07 - 09:11:42 | 200 |    2743h29m3s |     20.20.20.20 | GET      \"/\"\n", DefaultLogFormatter(termFalseLongDurationParam))

	assert.Equal(t, "[TEST] 2018/12/07 - 09:11:42 |\x1b[97;42m 200 \x1b[0m|            5s |     20.20.20.20 |\x1b[97;44m GET     \x1b[0m \"/\"\n", DefaultLogFormatter(termTrueParam))
	assert.Equal(t, "[TEST] 2018/12/07 - 09:11:42 |\x1b[97;42m 200 \x1b[0m|    2743h29m3s |     20.20.20.20 |\x1b[97;44m GET     \x1b[0m \"/\"\n", DefaultLogFormatter(termTrueLongDurationParam))
}
