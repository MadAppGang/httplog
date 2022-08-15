package httplog

type FormatterFunction = func(param LogFormatterParams) string

// ChainLogFormatter chain a list of log formatters
var ChainLogFormatter = func(formatters ...FormatterFunction) FormatterFunction {
	return func(params LogFormatterParams) string {
		output := ""
		for _, f := range formatters {
			output += f(params)
		}
		return output
	}
}
