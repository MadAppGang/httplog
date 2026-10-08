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
	"net/http/httptrace"
	"net/textproto"
	"strconv"
	"strings"
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

func TestResponseWriterInformationalStatus(t *testing.T) {
	for _, status := range []int{http.StatusContinue, http.StatusProcessing, http.StatusEarlyHints, 199} {
		for _, mode := range []string{"header", "write", "readFrom", "flush"} {
			t.Run(strconv.Itoa(status)+"/"+mode, func(t *testing.T) {
				type result struct {
					interimStatus  int
					interimWritten bool
					interimBefore  int
					status         int
					written        bool
					size           int
					before         []int
					err            error
				}
				results := make(chan result, 1)
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					rw := NewResponseWriter(w)
					var got result
					rw.Before(func(w ResponseWriter) {
						got.before = append(got.before, w.Status())
						w.Header().Set("X-Before", "final")
					})
					rw.WriteHeader(status)
					got.interimStatus = rw.Status()
					got.interimWritten = rw.Written()
					got.interimBefore = len(got.before)
					switch mode {
					case "header":
						rw.WriteHeader(http.StatusCreated)
						_, got.err = rw.Write([]byte("created"))
					case "write":
						_, got.err = rw.Write([]byte("created"))
					case "readFrom":
						_, got.err = io.Copy(rw, struct{ io.Reader }{strings.NewReader("created")})
					case "flush":
						rw.Flush()
						_, got.err = rw.Write([]byte("created"))
					}
					rw.WriteHeader(http.StatusNotFound)
					rw.WriteHeader(http.StatusEarlyHints)
					got.status, got.written, got.size = rw.Status(), rw.Written(), rw.Size()
					results <- got
				}))
				defer server.Close()

				var informational []int
				var informationalHeaders []http.Header
				trace := &httptrace.ClientTrace{Got1xxResponse: func(code int, header textproto.MIMEHeader) error {
					informational = append(informational, code)
					informationalHeaders = append(informationalHeaders, http.Header(header).Clone())
					return nil
				}}
				req, err := http.NewRequest(http.MethodGet, server.URL, nil)
				if !assert.NoError(t, err) {
					return
				}
				req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace))
				res, err := server.Client().Do(req)
				if !assert.NoError(t, err) {
					return
				}
				defer res.Body.Close()
				body, err := io.ReadAll(res.Body)
				assert.NoError(t, err)
				got := <-results
				wantStatus := http.StatusOK
				if mode == "header" {
					wantStatus = http.StatusCreated
				}
				assert.NoError(t, got.err)
				assert.Equal(t, []int{status}, informational)
				for _, header := range informationalHeaders {
					assert.Empty(t, header.Get("X-Before"))
				}
				assert.Zero(t, got.interimStatus)
				assert.False(t, got.interimWritten)
				assert.Zero(t, got.interimBefore)
				assert.Equal(t, wantStatus, got.status)
				assert.True(t, got.written)
				assert.Equal(t, len("created"), got.size)
				assert.Equal(t, []int{wantStatus}, got.before)
				assert.Equal(t, wantStatus, res.StatusCode)
				assert.Equal(t, "final", res.Header.Get("X-Before"))
				assert.Equal(t, "created", string(body))
			})
		}
	}
}

func TestResponseWriterSwitchingProtocolsIsFinal(t *testing.T) {
	rec := httptest.NewRecorder()
	rw := NewResponseWriter(rec)
	var before []int
	rw.Before(func(w ResponseWriter) { before = append(before, w.Status()) })
	rw.WriteHeader(http.StatusSwitchingProtocols)
	rw.WriteHeader(http.StatusOK)
	rw.WriteHeader(http.StatusEarlyHints)
	assert.Equal(t, http.StatusSwitchingProtocols, rw.Status())
	assert.Equal(t, http.StatusSwitchingProtocols, rec.Code)
	assert.True(t, rw.Written())
	assert.Equal(t, []int{http.StatusSwitchingProtocols}, before)
}

func TestLoggerMultipleInformationalResponses(t *testing.T) {
	logged := make(chan LogFormatterParams, 1)
	logger, err := LoggerWithConfig(LoggerConfig{
		Output: io.Discard,
		Formatter: func(p LogFormatterParams) string {
			logged <- p
			return ""
		},
	})
	if !assert.NoError(t, err) {
		return
	}
	server := httptest.NewServer(logger.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Link", "</first.css>; rel=preload")
		w.WriteHeader(http.StatusEarlyHints)
		w.Header().Set("Link", "</second.css>; rel=preload")
		w.WriteHeader(http.StatusEarlyHints)
		w.Header().Del("Link")
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte("unavailable"))
	})))
	defer server.Close()
	var links []string
	trace := &httptrace.ClientTrace{Got1xxResponse: func(code int, header textproto.MIMEHeader) error {
		assert.Equal(t, http.StatusEarlyHints, code)
		links = append(links, header.Get("Link"))
		return nil
	}}
	req, err := http.NewRequest(http.MethodGet, server.URL, nil)
	if !assert.NoError(t, err) {
		return
	}
	req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace))
	res, err := server.Client().Do(req)
	if !assert.NoError(t, err) {
		return
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	assert.NoError(t, err)
	param := <-logged
	assert.Equal(t, []string{"</first.css>; rel=preload", "</second.css>; rel=preload"}, links)
	assert.Equal(t, http.StatusServiceUnavailable, res.StatusCode)
	assert.Equal(t, http.StatusServiceUnavailable, param.StatusCode)
	assert.Equal(t, LevelError, param.Level)
	assert.Equal(t, len("unavailable"), param.BodySize)
	assert.Equal(t, "unavailable", string(body))
}
