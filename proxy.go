package httplog

// Copyright 2022 Jack Rudenko. MadAppGang. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

import (
	"net"
	"net/http"
	"strings"
)

// ProxyType is a proxy type we support
type ProxyType string

// Supported proxies
const (
	// ProxyGoogleAppEngine when running on Google App Engine. Trust X-Appengine-Remote-Addr
	// for determining the client's IP
	ProxyGoogleAppEngine = "X-Appengine-Remote-Addr"
	// ProxyCloudflare when using Cloudflare's CDN. Trust CF-Connecting-IP for determining
	// the client's IP
	ProxyCloudflare = "CF-Connecting-IP"

	// ProxyDefaultType represents default proxy type
	ProxyDefaultType = ""
)

// Default proxy remote IP headers
var DefaultRemoteIPHeaders = []string{"X-Forwarded-For", "X-Real-IP"}

type Proxy struct {
	ptype           ProxyType
	RemoteIPHeaders []string
}

// NewProxy creates and returns default proxy with default params
func NewProxy() *Proxy {
	return &Proxy{
		ptype:           ProxyDefaultType,
		RemoteIPHeaders: DefaultRemoteIPHeaders,
	}
}

// NewProxy creates and returns proxy with specific type
func NewProxyWithType(t ProxyType) *Proxy {
	return &Proxy{
		ptype:           t,
		RemoteIPHeaders: DefaultRemoteIPHeaders,
	}
}

// NewProxy creates and returns proxy with specific type and proxy headers
func NewProxyWithTypeAndHeaders(t ProxyType, header []string) *Proxy {
	return &Proxy{
		ptype:           t,
		RemoteIPHeaders: header,
	}
}

// ClientIP implements one best effort algorithm to return the real client IP.
// It calls c.RemoteIP() under the hood, to check if the remote IP is a trusted proxy or not.
// If it is it will then try to parse the headers defined in Engine.RemoteIPHeaders (defaulting to [X-Forwarded-For, X-Real-Ip]).
// If the headers are not syntactically valid OR the remote IP does not correspond to a trusted proxy,
// the remote IP (coming from Request.RemoteAddr) is returned.
func (p *Proxy) ClientIP(r *http.Request) string {
	// Check if we're running on a trusted platform, continue running backwards if error
	if p.ptype != "" {
		// Developers can define their own header of Trusted Platform or use predefined constants
		if addr := r.Header.Get(p.ptype.String()); addr != "" {
			return addr
		}
	}

	remoteIP := net.ParseIP(RemoteIP(r))

	if p.RemoteIPHeaders != nil {
		for _, headerName := range p.RemoteIPHeaders {
			ip, valid := p.validateHeader(r.Header.Get(headerName))
			if valid {
				return ip
			}
		}
	}
	return remoteIP.String()
}

// validateHeader will parse X-Forwarded-For like header and return the trusted client IP address
func (p *Proxy) validateHeader(header string) (clientIP string, valid bool) {
	if header == "" {
		return "", false
	}
	header = strings.ReplaceAll(header, "[", "")
	header = strings.ReplaceAll(header, "]", "")
	items := strings.Split(header, ",")
	// we are getting the first ip in the list, which should refer to client's ip in proxy chain
	if len(items) > 0 {
		ipStr := strings.TrimSpace(items[0])
		ip := net.ParseIP(ipStr)
		if ip != nil {
			return ipStr, true
		}
	}
	return "", false
}

// RemoteIP parses the IP from Request.RemoteAddr, normalizes and returns the IP (without the port).
func RemoteIP(r *http.Request) string {
	ip, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr))
	if err != nil {
		return ""
	}
	return ip
}

func (pt ProxyType) String() string {
	return string(pt)
}
