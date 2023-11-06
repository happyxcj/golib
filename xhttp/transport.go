package xhttp

import (
	"context"
	"net"
	"net/http"
	"time"
)

// 传输选项方法
type TransOptFn func(t *http.Transport)

// NewTransport 新建Transport实例，
// 默认：MaxIdleConns=500，IdleConnTimeout=90秒，TLSHandshakeTimeout=15秒
// 使用指定的optFns来设置不同的Transport选项
func NewTransport(optFns ...TransOptFn) *http.Transport {
	tmpT := &http.Transport{
		// 和http库DefaultTransport一样默认配置
		ForceAttemptHTTP2:   true,
		MaxIdleConns:        500,
		IdleConnTimeout:     90 * time.Second,
		TLSHandshakeTimeout: 15 * time.Second,
		DialContext:         NewDialer().DialContext,
	}
	for _, optFn := range optFns {
		optFn(tmpT)
	}
	return tmpT
}

// TransWithIdleConnTimeout 返回对指定t设置IdleConnTimeout的选项方法
func TransWithIdleConnTimeout(idleConnTimeout time.Duration) TransOptFn {
	return func(t *http.Transport) {
		t.IdleConnTimeout = idleConnTimeout
	}
}

// TransWithMaxIdleConns 返回对指定t设置MaxIdleConns的选项方法
func TransWithMaxIdleConns(maxIdleConns int) TransOptFn {
	return func(t *http.Transport) {
		t.MaxIdleConns = maxIdleConns
	}
}

// TransWithTLSHandshakeTimeout 返回对指定t设置TLSHandshakeTimeout的选项方法
func TransWithTLSHandshakeTimeout(tlsHandshakeTimeout time.Duration) TransOptFn {
	return func(t *http.Transport) {
		t.TLSHandshakeTimeout = tlsHandshakeTimeout
	}
}

// TransWithDisableKeepAlive 返回对指定t设置DisableKeepAlives的选项方法
func TransWithDisableKeepAlive(disableKeepAlive bool) TransOptFn {
	return func(t *http.Transport) {
		t.DisableKeepAlives = disableKeepAlive
	}
}

// TransWithDialContext 返回对指定t设置DialContext的选项方法
func TransWithDialContext(dialContext func(ctx context.Context, network, addr string) (net.Conn, error)) TransOptFn {
	return func(t *http.Transport) {
		t.DialContext = dialContext
	}
}

// TransWithDisableCompression 返回对指定t设置DisableCompression的选项方法
func TransWithDisableCompression(disableCompression bool) TransOptFn {
	return func(t *http.Transport) {
		t.DisableCompression = disableCompression
	}
}

// TransWithResponseHeaderTimeout 返回对指定t设置ResponseHeaderTimeout的选项方法
func TransWithResponseHeaderTimeout(responseHeaderTimeout time.Duration) TransOptFn {
	return func(t *http.Transport) {
		t.ResponseHeaderTimeout = responseHeaderTimeout
	}
}
