package xgrpc

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
	"log"
	"time"
)

// 客户端接口
type IClient interface {
	Invoke(ctx context.Context, method string, req, resp interface{}, opts ...grpc.CallOption) error
}

// 基础客户端
type BaseClient struct {
	Conn *grpc.ClientConn
}

// NewBaseClientWithMaxDelay 根据地址target和拨号ConnectParams的最大重试延迟maxDelay创建默认实例
// 默认连接参数如下:
//
// grpc.ConnectParams{
//		Backoff: backoff.Config{
//			BaseDelay:  500 * time.Millisecond,
//			Multiplier: 1.6,
//			Jitter:     0.2,
//			MaxDelay:   maxDelay,
//		},
//		MinConnectTimeout: 15 * time.Second,
// }
func NewBaseClientWithMaxDelay(target string, maxDelay time.Duration) *BaseClient {
	connParams := grpc.ConnectParams{
		Backoff: backoff.Config{
			BaseDelay:  500 * time.Millisecond,
			Multiplier: 1.6,
			Jitter:     0.2,
			MaxDelay:   maxDelay,
		},
		MinConnectTimeout: 15 * time.Second,
	}
	return NewBaseClientWithOpts(target, grpc.WithInsecure(), grpc.WithConnectParams(connParams))
}

func NewBaseClientWithOpts(target string, opts ...grpc.DialOption) *BaseClient {
	conn, err := grpc.DialContext(context.Background(), target, opts...)
	if err != nil {
		log.Fatal(fmt.Sprintf("DialContext '%v' err: %v", target, err))
	}
	return NewBaseClient(conn)
}

func NewBaseClient(conn *grpc.ClientConn) *BaseClient {
	return &BaseClient{
		Conn: conn,
	}
}

func (c *BaseClient) Invoke(ctx context.Context, method string, req, resp interface{}, opts ...grpc.CallOption) error {
	return c.Conn.Invoke(ctx, method, req, resp, opts...)
}

// 使用相同服务名字的客户端，即最终的method=`/serviceName/method`
type ServiceClient struct {
	Inner       IClient
	ServiceName string
}

func NewServiceClient(inner IClient, serviceName string) *ServiceClient {
	return &ServiceClient{
		Inner:       inner,
		ServiceName: serviceName,
	}
}

func (c *ServiceClient) Invoke(ctx context.Context, method string, req, resp interface{}, opts ...grpc.CallOption) error {
	method = fmt.Sprintf("/%v/%v", c.ServiceName, method)
	return c.Inner.Invoke(ctx, method, req, resp, opts...)
}

// 统计请求时长客户端
type DurClient struct {
	Inner IClient
	// 处理请求时长的方法
	HandleDurFn func(method string, dur time.Duration)
}

func NewDurClient(inner IClient, handleDurFn func(method string, dur time.Duration)) *DurClient {
	return &DurClient{
		Inner:       inner,
		HandleDurFn: handleDurFn,
	}
}

func (c *DurClient) Invoke(ctx context.Context, method string, req, resp interface{}, opts ...grpc.CallOption) error {
	startT := time.Now()
	defer func() {
		c.HandleDurFn(method, time.Since(startT))
	}()
	return c.Inner.Invoke(ctx, method, req, resp, opts...)
}

// 每个请求加上超时时间客户端
type TimeoutClient struct {
	Inner IClient
	// 默认超时
	Timeout time.Duration
	// 指定方法的超时时间
	MethodTimeouts map[string]time.Duration
}

func NewTimeoutClient(inner IClient, timeout time.Duration) *TimeoutClient {
	return &TimeoutClient{
		Inner:          inner,
		Timeout:        timeout,
		MethodTimeouts: make(map[string]time.Duration),
	}
}

// AddMethodTimeout 添加一个指定方法的超时时间
func (c *TimeoutClient) AddMethodTimeout(method string, timeout time.Duration) *TimeoutClient {
	c.MethodTimeouts[method] = timeout
	return c
}

// AddMethodTimeouts 添加多个指定方法的超时时间，注意：不会覆盖c已有method
func (c *TimeoutClient) AddMethodTimeouts(methodTimeouts map[string]time.Duration) *TimeoutClient {
	for method, timeout := range methodTimeouts {
		c.AddMethodTimeout(method, timeout)
	}
	return c
}

func (c *TimeoutClient) Invoke(ctx context.Context, method string, req, resp interface{}, opts ...grpc.CallOption) error {
	var timeout time.Duration
	if v, ok := c.MethodTimeouts[method]; ok {
		timeout = v
	} else {
		timeout = c.Timeout
	}
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	return c.Inner.Invoke(timeoutCtx, method, req, resp, opts...)
}
