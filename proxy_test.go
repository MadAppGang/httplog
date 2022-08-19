package httplog_test

// Copyright 2022 Jack Rudenko. MadAppGang. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.
import (
	"context"
	"net/http"
	"testing"

	"github.com/MadAppGang/httplog"
	"github.com/stretchr/testify/assert"
)

func setRemoteIPHeaders(r *http.Request) {
	r.Header.Set("X-Real-IP", " 10.10.10.10  ")
	r.Header.Set("X-Forwarded-For", "  20.20.20.20, 30.30.30.30")
	r.Header.Set("X-Appengine-Remote-Addr", "50.50.50.50")
	r.Header.Set("CF-Connecting-IP", "60.60.60.60")
	r.RemoteAddr = "  40.40.40.40:42123 "
}

func TestContextClientIP(t *testing.T) {
	request, _ := http.NewRequestWithContext(context.Background(), "POST", "/", nil)
	setRemoteIPHeaders(request)
	// default proxy
	dp := httplog.NewProxy()

	assert.Equal(t, "20.20.20.20", dp.ClientIP(request))

	request.Header.Del("X-Forwarded-For")
	assert.Equal(t, "10.10.10.10", dp.ClientIP(request))

	request.Header.Set("X-Forwarded-For", "30.30.30.30  ")
	assert.Equal(t, "30.30.30.30", dp.ClientIP(request))

	request.Header.Del("X-Forwarded-For")
	request.Header.Del("X-Real-IP")

	googleProxy := httplog.NewProxyWithType(httplog.ProxyGoogleAppEngine)
	assert.Equal(t, "50.50.50.50", googleProxy.ClientIP(request))

	request.Header.Del("X-Appengine-Remote-Addr")
	assert.Equal(t, "40.40.40.40", googleProxy.ClientIP(request))

	// no port
	request.RemoteAddr = "50.50.50.50"
	assert.Empty(t, googleProxy.ClientIP(request))

	// X-Forwarded-For has a non-IP element
	request.Header.Set("X-Forwarded-For", " blah ")
	request.RemoteAddr = "40.40.40.40:1234"
	assert.Equal(t, "40.40.40.40", dp.ClientIP(request))

	// Use custom Proxy type
	customProxy := httplog.NewProxyWithType("X-CDN-IP")
	request.Header.Set("X-CDN-IP", "80.80.80.80")
	assert.Equal(t, "80.80.80.80", customProxy.ClientIP(request))

	// use custom headers instead of type
	customProxy = httplog.NewProxyWithTypeAndHeaders(httplog.ProxyDefaultType, []string{"X-CDN-IP"})
	assert.Equal(t, "80.80.80.80", customProxy.ClientIP(request))

	// wrong header
	customProxy = httplog.NewProxyWithType("X-Wrong-Header")
	assert.Equal(t, "40.40.40.40", customProxy.ClientIP(request))

	request.Header.Del("X-CDN-IP")

	// ProxyType is empty
	customProxy = httplog.NewProxyWithType("")
	assert.Equal(t, "40.40.40.40", customProxy.ClientIP(request))

	// Cloud Flare
	cfProxy := httplog.NewProxyWithType(httplog.ProxyCloudflare)
	assert.Equal(t, "60.60.60.60", cfProxy.ClientIP(request))

	request.Header.Del("CF-Connecting-IP")
	assert.Equal(t, "40.40.40.40", cfProxy.ClientIP(request))

	// no port
	request.RemoteAddr = "50.50.50.50"
	assert.Empty(t, dp.ClientIP(request))
}
