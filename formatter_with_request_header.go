package httplog

import "fmt"

// Copyright 2022 Jack Rudenko. MadAppGang. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

// HeaderLogFormatter format function with headers output.
func RequestHeaderLogFormatter(param LogFormatterParams) string {
	output := ""

	var blueColor, greenColor, resetColor string

	if param.IsOutputColor() {
		blueColor = "\033[1;34m"
		greenColor = "\033[;32m"
		resetColor = param.ResetColor()
	}
	for key, value := range param.Request.Header {
		output += fmt.Sprintf("  %s %s %s: %s %s %s\n",
			blueColor, key, resetColor,
			greenColor, value, resetColor,
		)
	}

	return output
}

// DefaultLogFormatterWithHeaders is a combination of default log formatter and header log formatter
var DefaultLogFormatterWithRequestHeader = ChainLogFormatter(DefaultLogFormatter, RequestHeaderLogFormatter)
