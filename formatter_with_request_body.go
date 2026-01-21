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

// RequestBodyLogFormatter format function with JSON body output or text
// Note: Requires CaptureRequestBody to be enabled in LoggerConfig
func RequestBodyLogFormatter(param LogFormatterParams) string {
	var blueColor, yellowColor, greenColor, resetColor string
	if param.IsOutputColor() {
		blueColor = blue
		yellowColor = yellow
		greenColor = green
		resetColor = param.ResetColor()
	}

	if len(param.RequestBody) == 0 {
		return fmt.Sprintf("===\n%s EMPTY BODY %s\n===\n", yellowColor, resetColor)
	}

	var bodyJSON map[string]interface{}
	err := json.Unmarshal(param.RequestBody, &bodyJSON)
	if err != nil {
		// it is not a json
		text := bytes.ToValidUTF8(param.RequestBody, nil)
		return fmt.Sprintf("===\n%s TEXT BODY:%s\n%s\n===\n", blueColor, resetColor, string(text))
	}

	f := colorjson.NewFormatter()
	f.Indent = 2
	s, _ := f.Marshal(bodyJSON)
	return fmt.Sprintf("===\n%s JSON BODY:%s\n%s\n===\n", greenColor, resetColor, string(s))
}

// DefaultLogFormatterWithHeadersAndBody is a combination of default log formatter, header log formatter and json body
var DefaultLogFormatterWithRequestHeadersAndBody = ChainLogFormatter(DefaultLogFormatter, RequestHeaderLogFormatter, RequestBodyLogFormatter)
