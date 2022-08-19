package httplog

import "fmt"

// Copyright 2022 Jack Rudenko. MadAppGang. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

// ResponseHeaderLogFormatter format function with headers output.
func ResponseHeaderLogFormatter(param LogFormatterParams) string {
	var blueColor, greenColor, resetColor, output string

	if param.IsOutputColor() {
		blueColor = "\033[1;34m"
		greenColor = "\033[;32m"
		resetColor = param.ResetColor()
	}
	for key, value := range param.ResponseHeader {
		output += fmt.Sprintf("  %s %s %s: %s %s %s\n",
			blueColor, key, resetColor,
			greenColor, value, resetColor,
		)
	}

	return output
}

// DefaultLogFormatterWithHeader is a combination of default log formatter and header log formatter
var DefaultLogFormatterWithResponseHeader = ChainLogFormatter(DefaultLogFormatter, ResponseHeaderLogFormatter)
