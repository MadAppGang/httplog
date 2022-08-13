package httplog

// Copyright 2022 Jack Rudenko. MadAppGang. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

import (
	"io"
	"os"
)

// DefaultWriter is the default io.Writer used by httplog for
// middleware output Logger() and Body().
// To support coloring in Windows use:
//
//	import "github.com/mattn/go-colorable"
//	httplog.DefaultWriter = colorable.NewColorableStdout()
var DefaultWriter io.Writer = os.Stdout

// DefaultErrorWriter is the default io.Writer used by httplog to debug errors
var DefaultErrorWriter io.Writer = os.Stderr
