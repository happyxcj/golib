package xhttp

import (
	"context"
	"io"
	"net/http"
	"time"
)

// 客户端接口
type IClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// 基础客户端
type BaseClient struct {
	cli *http.Client
}

func NewBaseClient(cli *http.Client) *BaseClient {
	return &BaseClient{
		cli: cli,
	}
}

// NewDefaultBaseClient 创建默认基础客户端实例，可自定义Transport选项
func NewDefaultBaseClient(opts ...TransOptFn) *BaseClient {
	transport := NewTransport(opts...)
	return NewBaseClientByTransport(transport, 0)
}

// NewBaseClientWithTimout 根据地址transport和客户端超时时间timeout创建默认实例
func NewBaseClientByTransport(transport *http.Transport, timeout time.Duration) *BaseClient {
	cli := &http.Client{
		Transport: transport,
		Timeout:   timeout,
	}
	return NewBaseClient(cli)
}

func (c *BaseClient) Do(req *http.Request) (*http.Response, error) {
	return c.cli.Do(req)
}

// 可方便调用Post，Get等方法的客户端
type MethodClient struct {
	Inner IClient
	// 请求编码方法
	EncodeFn EncodeReqFn
	// 响应解码方法
	DecodeFn DecodeRespFn
	// 基本url，可为空，有设置时http请求的最终url=baseUrl+paramUrl(请求参数传的url)
	baseUrl string
}

// NewMethodClient 返回MethodClient实例
// inner: 内部客户端，encodeFn和decodeFn可根据需要设置，不需要时可传nil
func NewMethodClient(inner IClient, encodeFn EncodeReqFn, decodeFn DecodeRespFn) *MethodClient {
	return &MethodClient{
		Inner:    inner,
		EncodeFn: encodeFn,
		DecodeFn: decodeFn,
	}
}

// NewMethodClient 返回MethodClient实例，使用json序列化req，反序列化resp
func NewMethodJsonClient(inner IClient) *MethodClient {
	return NewMethodClient(inner, EncodeJsonReq, DecodeJsonResp)
}

func (c *MethodClient) WithBaseUrl(baseUrl string) *MethodClient {
	c.baseUrl = baseUrl
	return c
}

func (c *MethodClient) Do(req *http.Request) (*http.Response, error) {
	return c.Inner.Do(req)
}

// DoMethod 执行指定方法method的http请求，请求body信息为reqMsg
// 注意：没有请求体时reqMsg可传nil
// 返回http响应
func (c *MethodClient) DoMethod(method, url string, reqMsg interface{}) (*http.Response, error) {
	var body io.Reader
	var err error
	if reqMsg != nil {
		// 带请求体body
		body, err = c.EncodeFn(reqMsg)
	}
	if err != nil {
		return nil, err
	}
	if c.baseUrl != "" {
		url = c.baseUrl + url
	}
	req, err := NewReq(method, url, body)
	if err != nil {
		return nil, err
	}
	return c.Inner.Do(req)
}

// DoMethod 执行Get方式的http请求
// 返回http响应
func (c *MethodClient) Get(url string) (*http.Response, error) {
	return c.DoMethod(http.MethodGet, url, nil)
}

// DoMethod 执行Post方式的http请求，请求body信息为reqMsg
// 注意：没有请求体时reqMsg可传nil
// 返回http响应
func (c *MethodClient) Post(url string, reqMsg interface{}) (*http.Response, error) {
	return c.DoMethod(http.MethodPost, url, reqMsg)
}

// DoMethod 执行Put方式的http请求，请求body信息为reqMsg
// 注意：没有请求体时reqMsg可传nil
// 返回http响应
func (c *MethodClient) Put(url string, reqMsg interface{}) (*http.Response, error) {
	return c.DoMethod(http.MethodPut, url, reqMsg)
}

// DoMethod 执行Delete方式的http请求，请求body信息为reqMsg
// 注意：没有请求体时reqMsg可传nil
// 返回http响应
func (c *MethodClient) Delete(url string, reqMsg interface{}) (*http.Response, error) {
	return c.DoMethod(http.MethodDelete, url, reqMsg)
}

// DoAndDecode 执行指定方法method的http请求，请求body信息为reqMsg，响应成功后反序列化响应到消息respMsg
// 注意：没有请求体时reqMsg可传nil
func (c *MethodClient) DoAndDecode(method, url string, reqMsg, respMsg interface{}) error {
	resp, err := c.DoMethod(method, url, reqMsg)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return c.DecodeFn(resp.Body, respMsg)
}

// DoAndDecode 执行Get方式的http请求，响应成功后反序列化响应到消息respMsg
func (c *MethodClient) GetAndDecode(url string, respMsg interface{}) error {
	return c.DoAndDecode(http.MethodGet, url, nil, respMsg)
}

// DoAndDecode 执行Post方式的http请求，请求body信息为reqMsg，响应成功后反序列化响应到消息respMsg
// 注意：没有请求体时reqMsg可传nil
func (c *MethodClient) PostAndDecode(url string, reqMsg, respMsg interface{}) error {
	return c.DoAndDecode(http.MethodPost, url, reqMsg, respMsg)
}

// DoAndDecode 执行Put方式的http请求，请求body信息为reqMsg，响应成功后反序列化响应到消息respMsg
// 注意：没有请求体时reqMsg可传nil
func (c *MethodClient) PutAndDecode(url string, reqMsg, respMsg interface{}) error {
	return c.DoAndDecode(http.MethodPut, url, reqMsg, respMsg)
}

// DoAndDecode 执行Delete方式的http请求，请求body信息为reqMsg，响应成功后反序列化响应到消息respMsg
// 注意：没有请求体时reqMsg可传nil
func (c *MethodClient) DeleteAndDecode(url string, reqMsg, respMsg interface{}) error {
	return c.DoAndDecode(http.MethodDelete, url, reqMsg, respMsg)
}

// 统计请求时长客户端
type DurClient struct {
	Inner IClient
	// 处理请求时长的方法
	HandleDurFn func(dur time.Duration)
}

func NewDurClient(inner IClient, handleDurFn func(dur time.Duration)) *DurClient {
	return &DurClient{
		Inner:       inner,
		HandleDurFn: handleDurFn,
	}
}

func (c *DurClient) Do(req *http.Request) (*http.Response, error) {
	startT := time.Now()
	defer func() {
		c.HandleDurFn(time.Since(startT))
	}()
	return c.Inner.Do(req)
}

// 每个请求加上超时时间客户端
type TimeoutClient struct {
	Inner IClient
	// 默认超时
	Timeout time.Duration
	// 指定path的超时时间
	PathTimeouts map[string]time.Duration
}

func NewTimeoutClient(inner IClient, timeout time.Duration) *TimeoutClient {
	return &TimeoutClient{
		Inner:        inner,
		Timeout:      timeout,
		PathTimeouts: map[string]time.Duration{},
	}
}

// AddMethodTimeout 添加指定paths使用相同超时时间timeout
func (c *TimeoutClient) AddPathTimeouts(timeout time.Duration, paths ...string) *TimeoutClient {
	for _, path := range paths {
		c.PathTimeouts[path] = timeout
	}
	return c
}

func (c *TimeoutClient) Do(req *http.Request) (*http.Response, error) {
	var timeout time.Duration
	if v, ok := c.PathTimeouts[req.URL.Path]; ok {
		timeout = v
	} else {
		timeout = c.Timeout
	}
	requestCtx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	req = req.WithContext(requestCtx)
	return c.Inner.Do(req)
}

// 每个请求加上header客户端
type HeaderClient struct {
	Inner   IClient
	Headers map[string]string
}

func NewHeaderClient(inner IClient, headers map[string]string) *HeaderClient {
	return &HeaderClient{
		Inner:   inner,
		Headers: headers,
	}
}

// AddKV 添加单个请求头，k为key, v为value
func (c *HeaderClient) AddKV(k, v string) *HeaderClient {
	if c.Headers == nil {
		c.Headers = make(map[string]string)
	}
	c.Headers[k] = v
	return c
}

func (c *HeaderClient) Do(req *http.Request) (*http.Response, error) {
	for k, v := range c.Headers {
		req.Header.Set(k, v)
	}
	return c.Inner.Do(req)
}
