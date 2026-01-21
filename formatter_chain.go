package httplog

import "strings"

// ChainLogFormatter chain a list of log formatters
func ChainLogFormatter(formatters ...LogFormatter) LogFormatter {
	return func(params LogFormatterParams) string {
		var output strings.Builder
		for _, f := range formatters {
			output.WriteString(f(params))
		}
		return output.String()
	}
}
