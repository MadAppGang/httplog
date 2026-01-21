package httplog

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestChainLogFormatter(t *testing.T) {
	timeStamp := time.Unix(1544173902, 0).UTC()

	termTrueLongDurationParam := LogFormatterParams{
		RouterName: "TEST",
		TimeStamp:  timeStamp,
		StatusCode: 200,
		Latency:    time.Millisecond * 9876543210,
		ClientIP:   "20.20.20.20",
		Method:     "GET",
		Path:       "/",
		colorMode:  ColorForce,
	}

	result := ChainLogFormatter(
		ShortLogFormatter,
		DefaultLogFormatter,
	)(termTrueLongDurationParam)

	assert.Equal(t,
		"[TEST] \x1b[97;42m 200 \x1b[0m| \"/\"\n[TEST] 2018/12/07 - 09:11:42 |\x1b[97;42m 200 \x1b[0m|    2743h29m3s |     20.20.20.20 |\x1b[97;44m GET     \x1b[0m \"/\"\n",
		result,
	)
}
