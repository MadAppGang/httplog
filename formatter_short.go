package httplog

import (
	"fmt"
	"time"
)

// Copyright 2022 Jack Rudenko. MadAppGang. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

func ShortLogFormatter(param LogFormatterParams) string {
	var statusColor, resetColor string
	if param.IsOutputColor() {
		statusColor = param.StatusCodeColor()
		resetColor = param.ResetColor()
	}

	if param.Latency > time.Minute {
		param.Latency = param.Latency.Truncate(time.Second)
	}
	return fmt.Sprintf("[%s] %s %3d %s| %#v\n",
		param.RouterName,
		statusColor, param.StatusCode, resetColor,
		param.Path,
	)
}
