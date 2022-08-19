package httplog

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/TylerBrock/colorjson"
)

// Copyright 2022 Jack Rudenko. MadAppGang. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

// ResponseBodyLogFormatter format function with JSON body output or text
func ResponseBodyLogFormatter(param LogFormatterParams) string {
	var blueColor, yellowColor, greenColor, resetColor string
	if param.IsOutputColor() {
		blueColor = blue
		yellowColor = yellow
		greenColor = green
		resetColor = param.ResetColor()
	}

	if len(param.Body) == 0 {
		return fmt.Sprintf("===\n%s EMPTY BODY %s\n===\n", yellowColor, resetColor)
	}

	var body map[string]interface{}
	err := json.Unmarshal(param.Body, &body)
	if err != nil {
		// it is not a json
		text := bytes.ToValidUTF8(param.Body, nil)
		return fmt.Sprintf("===\n%s TEXT BODY:%s\n%s\n===\n", blueColor, resetColor, string(text))
	}

	f := colorjson.NewFormatter()
	f.Indent = 2
	s, _ := f.Marshal(body)
	return fmt.Sprintf("===\n%s JSON BODY:%s\n%s\n===\n", greenColor, resetColor, string(s))
}

// DefaultLogFormatterWithHeadersAndBody is a combination of default log formatter, header log formatter and json body
var DefaultLogFormatterWithResponseHeadersAndBody = ChainLogFormatter(DefaultLogFormatter, ResponseHeaderLogFormatter, ResponseBodyLogFormatter)
