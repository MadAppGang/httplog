package httplog

// Copyright 2022 Jack Rudenko. MadAppGang. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

import (
	"bufio"
	"bytes"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

type hijackableResponse struct {
	Hijacked bool
}

func newHijackableResponse() *hijackableResponse {
	return &hijackableResponse{}
}

func (h *hijackableResponse) Header() http.Header           { return nil }
func (h *hijackableResponse) Write(buf []byte) (int, error) { return 0, nil }
func (h *hijackableResponse) WriteHeader(code int)          {}
func (h *hijackableResponse) Flush()                        {}
func (h *hijackableResponse) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	h.Hijacked = true
	return nil, nil, nil
}

func TestResponseWriterBeforeWrite(t *testing.T) {
	rec := httptest.NewRecorder()
	rw := NewResponseWriter(rec)

	assert.Equal(t, rw.Status(), 0)
	assert.Equal(t, rw.Written(), false)
}

func TestResponseWriterBeforeFuncHasAccessToStatus(t *testing.T) {
	var status int

	rec := httptest.NewRecorder()
	rw := NewResponseWriter(rec)

	rw.Before(func(w ResponseWriter) {
		status = w.Status()
	})
	rw.WriteHeader(http.StatusCreated)

	assert.Equal(t, status, http.StatusCreated)
}

func TestResponseWriterWritingString(t *testing.T) {
	rec := httptest.NewRecorder()
	rw := NewResponseWriter(rec)

	_, _ = rw.Write([]byte("Hello world"))

	assert.Equal(t, rec.Code, rw.Status())
	assert.Equal(t, rec.Body.String(), "Hello world")
	assert.Equal(t, rw.Status(), http.StatusOK)
	assert.Equal(t, rw.Size(), 11)
	assert.Equal(t, rw.Written(), true)
}

func TestResponseWriterWritingStringShadowBody(t *testing.T) {
	rec := httptest.NewRecorder()
	rw := NewResponseWriterWithBody(rec)

	data := []byte("Hello world")
	_, _ = rw.Write(data)

	assert.Equal(t, rec.Code, rw.Status())
	assert.Equal(t, rec.Body.String(), "Hello world")
	assert.Equal(t, rw.Status(), http.StatusOK)
	assert.Equal(t, rw.Size(), 11)
	assert.Equal(t, rw.Written(), true)
	assert.Equal(t, bytes.Compare(data, rw.Body()), 0)
}

func TestResponseWriterWritingStrings(t *testing.T) {
	rec := httptest.NewRecorder()
	rw := NewResponseWriter(rec)

	_, _ = rw.Write([]byte("Hello world"))
	_, _ = rw.Write([]byte("foo bar bat baz"))

	assert.Equal(t, rec.Code, rw.Status())
	assert.Equal(t, rec.Body.String(), "Hello worldfoo bar bat baz")
	assert.Equal(t, rw.Status(), http.StatusOK)
	assert.Equal(t, rw.Size(), 26)
}

func TestResponseWriterWritingHeader(t *testing.T) {
	rec := httptest.NewRecorder()
	rw := NewResponseWriter(rec)

	rw.WriteHeader(http.StatusNotFound)

	assert.Equal(t, rec.Code, rw.Status())
	assert.Equal(t, rec.Body.String(), "")
	assert.Equal(t, rw.Status(), http.StatusNotFound)
	assert.Equal(t, rw.Size(), 0)
}

func TestResponseWriterWritingHeaderTwice(t *testing.T) {
	rec := httptest.NewRecorder()
	rw := NewResponseWriter(rec)

	rw.WriteHeader(http.StatusNotFound)
	rw.WriteHeader(http.StatusInternalServerError)

	assert.Equal(t, rec.Code, rw.Status())
	assert.Equal(t, rec.Body.String(), "")
	assert.Equal(t, rw.Status(), http.StatusNotFound)
	assert.Equal(t, rw.Size(), 0)
}

func TestResponseWriterBefore(t *testing.T) {
	rec := httptest.NewRecorder()
	rw := NewResponseWriter(rec)
	result := ""

	rw.Before(func(ResponseWriter) {
		result += "foo"
	})
	rw.Before(func(ResponseWriter) {
		result += "bar"
	})

	rw.WriteHeader(http.StatusNotFound)

	assert.Equal(t, rec.Code, rw.Status())
	assert.Equal(t, rec.Body.String(), "")
	assert.Equal(t, rw.Status(), http.StatusNotFound)
	assert.Equal(t, rw.Size(), 0)
	assert.Equal(t, result, "barfoo")
}

func TestResponseWriterHijack(t *testing.T) {
	hijackable := newHijackableResponse()
	rw := NewResponseWriter(hijackable)
	hijacker, ok := rw.(http.Hijacker)
	assert.Equal(t, ok, true)
	_, _, err := hijacker.Hijack()
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, hijackable.Hijacked, true)
}

func TestResponseWriteHijackNotOK(t *testing.T) {
	hijackable := new(http.ResponseWriter)
	rw := NewResponseWriter(*hijackable)
	hijacker, ok := rw.(http.Hijacker)
	assert.Equal(t, ok, true)
	_, _, err := hijacker.Hijack()

	assert.NotEqual(t, err, nil)
}

func TestResponseWriterFlusher(t *testing.T) {
	rec := httptest.NewRecorder()
	rw := NewResponseWriter(rec)

	_, ok := rw.(http.Flusher)
	assert.Equal(t, ok, true)
}

func TestResponseWriter_Flush_marksWritten(t *testing.T) {
	rec := httptest.NewRecorder()
	rw := NewResponseWriter(rec)

	rw.Flush()
	assert.Equal(t, rw.Status(), http.StatusOK)
	assert.Equal(t, rw.Written(), true)
}

// mockReader only implements io.Reader without other methods like WriterTo
type mockReader struct {
	readStr string
	eof     bool
}

func (r *mockReader) Read(p []byte) (n int, err error) {
	if r.eof {
		return 0, io.EOF
	}
	copy(p, []byte(r.readStr))
	r.eof = true
	return len(r.readStr), nil
}

func TestResponseWriterWithoutReadFrom(t *testing.T) {
	writeString := "Hello world"

	rec := httptest.NewRecorder()
	rw := NewResponseWriter(rec)

	n, err := io.Copy(rw, &mockReader{readStr: writeString})
	assert.Equal(t, err, nil)
	assert.Equal(t, rw.Status(), http.StatusOK)
	assert.Equal(t, rw.Written(), true)
	assert.Equal(t, rw.Size(), len(writeString))
	assert.Equal(t, int(n), len(writeString))
	assert.Equal(t, rec.Body.String(), writeString)
}

type mockResponseWriterWithReadFrom struct {
	*httptest.ResponseRecorder
	writtenStr string
}

func (rw *mockResponseWriterWithReadFrom) ReadFrom(r io.Reader) (n int64, err error) {
	bytes, err := io.ReadAll(r)
	if err != nil {
		return 0, err
	}
	rw.writtenStr = string(bytes)
	_, _ = rw.ResponseRecorder.Write(bytes)
	return int64(len(bytes)), nil
}

func TestResponseWriterWithReadFrom(t *testing.T) {
	writeString := "Hello world"
	mrw := &mockResponseWriterWithReadFrom{ResponseRecorder: httptest.NewRecorder()}
	rw := NewResponseWriter(mrw)
	n, err := io.Copy(rw, &mockReader{readStr: writeString})
	assert.Equal(t, err, nil)
	assert.Equal(t, rw.Status(), http.StatusOK)
	assert.Equal(t, rw.Written(), true)
	assert.Equal(t, rw.Size(), len(writeString))
	assert.Equal(t, int(n), len(writeString))
	assert.Equal(t, mrw.Body.String(), writeString)
	assert.Equal(t, mrw.writtenStr, writeString)
}
