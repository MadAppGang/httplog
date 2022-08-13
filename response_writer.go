package httplog

// Copyright 2022 Jack Rudenko. MadAppGang. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

import (
	"bufio"
	"errors"
	"io"
	"net"
	"net/http"
)

// ResponseWriter is a wrapper around http.ResponseWriter that provides extra information about
// the response. It is recommended that middleware handlers use this construct to wrap a responsewriter
// if the functionality calls for it.
type ResponseWriter interface {
	http.ResponseWriter
	http.Flusher
	http.Pusher
	// Status returns the status code of the response or 0 if the response has
	// not been written
	Status() int
	// Written returns whether or not the ResponseWriter has been written.
	Written() bool
	// Size returns the size of the response body.
	Size() int
	// Before allows for a function to be called before the ResponseWriter has been written to. This is
	// useful for setting headers or any other operations that must happen before a response has been written.
	Before(func(ResponseWriter))
	// Optional copy of response body
	// If you need to log or save full response bodies - use it
	// But extra memory and CPU will be used for that
	Body() []byte
}

type beforeFunc func(ResponseWriter)

// NewResponseWriter creates a ResponseWriter that wraps an http.ResponseWriter
func NewResponseWriter(rw http.ResponseWriter) ResponseWriter {
	nrw := &responseWriter{
		ResponseWriter: rw,
		copyBody:       false,
	}

	return nrw
}

// NewResponseWriterWithBody creates a ResponseWriter that wraps an http.ResponseWriter
// and copy the body of response. The body is not copied if you use ReadFrom for obvious reason
func NewResponseWriterWithBody(rw http.ResponseWriter) ResponseWriter {
	nrw := &responseWriter{
		ResponseWriter: rw,
		copyBody:       true,
	}

	return nrw
}

type responseWriter struct {
	http.ResponseWriter
	status      int
	size        int
	beforeFuncs []beforeFunc
	body        []byte
	copyBody    bool
}

func (rw *responseWriter) WriteHeader(s int) {
	if rw.Written() {
		return
	}
	rw.status = s
	rw.callBefore()
	rw.ResponseWriter.WriteHeader(s)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if !rw.Written() {
		// The status will be StatusOK if WriteHeader has not been called yet
		rw.WriteHeader(http.StatusOK)
	}
	size, err := rw.ResponseWriter.Write(b)
	rw.size += size
	if rw.copyBody {
		rw.body = append(rw.body, b...)
	}
	return size, err
}

// ReadFrom exposes underlying http.ResponseWriter to io.Copy and if it implements
// io.ReaderFrom, it can take advantage of optimizations such as sendfile, io.Copy
// with sync.Pool's buffer which is in http.(*response).ReadFrom and so on.
func (rw *responseWriter) ReadFrom(r io.Reader) (n int64, err error) {
	if !rw.Written() {
		// The status will be StatusOK if WriteHeader has not been called yet
		rw.WriteHeader(http.StatusOK)
	}
	n, err = io.Copy(rw.ResponseWriter, r)
	rw.size += int(n)
	return
}

func (rw *responseWriter) Status() int {
	return rw.status
}

func (rw *responseWriter) Size() int {
	return rw.size
}

func (rw *responseWriter) Written() bool {
	return rw.status != 0
}

func (rw *responseWriter) Before(before func(ResponseWriter)) {
	rw.beforeFuncs = append(rw.beforeFuncs, before)
}

func (rw *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := rw.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, errors.New("the ResponseWriter doesn't support the Hijacker interface")
	}
	return hijacker.Hijack()
}

func (rw *responseWriter) callBefore() {
	for i := len(rw.beforeFuncs) - 1; i >= 0; i-- {
		rw.beforeFuncs[i](rw)
	}
}

func (rw *responseWriter) Body() []byte {
	return rw.body
}

func (rw *responseWriter) Flush() {
	flusher, ok := rw.ResponseWriter.(http.Flusher)
	if ok {
		if !rw.Written() {
			// The status will be StatusOK if WriteHeader has not been called yet
			rw.WriteHeader(http.StatusOK)
		}
		flusher.Flush()
	}
}

func (rw *responseWriter) Push(target string, opts *http.PushOptions) error {
	pusher, ok := rw.ResponseWriter.(http.Pusher)
	if ok {
		return pusher.Push(target, opts)
	}
	return errors.New("the ResponseWriter doesn't support the Pusher interface")
}
