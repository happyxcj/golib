package xhttp

import (
	"context"
	"io"
	"net/http"
)

// ReqOptFn 对指定的r进行一些自定义处理，
// 返回一个http.Request实例(可以是参数r，也可以是一个新的实例)
type ReqOptFn func(r *http.Request) (*http.Request, error)

func NewReq(method, url string, body io.Reader, optFns ...ReqOptFn) (*http.Request, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	return ReqUseOptFns(req, optFns...)
}

// NewReqWithCtx 返回一个http.Request实例，设置ctx和指定的选项optFns
func NewReqWithCtx(ctx context.Context, method, url string, body io.Reader, optFns ...ReqOptFn) (*http.Request, error) {
	optFns = append(optFns, ReqWithCtx(ctx))
	return NewReq(method, url, body, optFns...)
}

// ReqWithHeader 返回对指定r设置一个请求头参数的选项方法
func ReqWithHeader(key, value string) ReqOptFn {
	return func(r *http.Request) (*http.Request, error) {
		r.Header.Set(key, value)
		return r, nil
	}
}

// ReqWithHeaders 返回对指定r设置多个请求头参数的选项方法
func ReqWithHeaders(kvs map[string]string) ReqOptFn {
	return func(r *http.Request) (*http.Request, error) {
		for k, v := range kvs {
			r.Header.Set(k, v)
		}
		return r, nil
	}
}

// ReqWithHeaders 返回对指定r设置多个请求头参数的选项方法
func ReqWithCtx(ctx context.Context) ReqOptFn {
	return func(r *http.Request) (*http.Request, error) {
		newReq := r.WithContext(ctx)
		return newReq, nil
	}
}

// ReqUseOptFns 对指定r设置指定的选项optFns
// 返回http.Request实例
func ReqUseOptFns(r *http.Request, optFns ...ReqOptFn) (*http.Request, error) {
	var err error
	tmpReq := r
	for _, optFn := range optFns {
		tmpReq, err = optFn(tmpReq)
		if err != nil {
			return nil, err
		}
	}
	return tmpReq, nil
}
