package xgrpc

import (
	"context"
	"fmt"
)

// abortIndex 标识CtxHandlerFn的最大个数
const abortIndex = 200

// CtxHandlerFn 是处理上下文的方法，req为请求参数，返回响应数据+错误
type CtxHandlerFn func(ctx *Context, req interface{}) (interface{}, error)

// 服务端操作链路上下文
type Context struct {
	// Context 是链路上下文，它的初始值是从grpc服务接口接收到的context.Context参数
	context.Context
	// 服务方法标识，不关注时可为""
	Method string
	// 处理链路过程，上一个处理器的响应resp，作为下一个处理器的请求参数req
	handlers []CtxHandlerFn
	// index 是上下文处理链路handlers当前的索引.
	index int
	// isErr 标识处理完后是否遇到了错误
	isErr bool
	// values 是贯穿整个上下文处理链路handlers的键值对信息
	// 常规用法是：在当前上下文处理器使用Context.Set方法设置键值对，
	// 然后在后续的上下文处理器中使用Context.Get或者Context.MustGet方法根据key获取设置的值
	values map[string]interface{}
}

// reset 重置上下文信息为初始状态.
func (c *Context) reset(method string, handlers ...CtxHandlerFn) {
	c.values = nil
	c.index = -1
	c.isErr = false
	c.handlers = handlers
	c.Method = method
}

// IsErr 返回是否上下文处理完，遇到了错误返回
func (c *Context) IsErr() bool {
	return c.isErr
}

// Next 执行上下文处理链路handlers中待处理方法
// 注意：它也可被当做处理器中间件使用，例如：
//
// func(c *xgrpc.Context, req interface{}) (interface{}, error)
// 	 startTime:=time.Now()
//   defer func() {
//		delay := time.Since(start)
//   }
// 	 return c.Next(req)
// }
//
func (c *Context) Next(req interface{}) (resp interface{}, err error) {
	c.index++
	for n := len(c.handlers); c.index < n; c.index++ {
		resp, err = c.handlers[c.index](c, req)
		if err != nil {
			c.isErr = true
			return nil, err
		}
		req = resp
	}
	return resp, nil
}

// Set 为上下文链路存储一个键值对信息
func (c *Context) Set(key string, v interface{}) {
	if c.values == nil {
		c.values = make(map[string]interface{})
	}
	c.values[key] = v
}

// Get 根据指定的key返回上下文中已存储的value，如果对于的key不存在则返回nil
func (c *Context) Get(key string) (interface{}, bool) {
	v, ok := c.values[key]
	return v, ok
}

// MustGet 根据指定的key返回上下文中已存储的value，如果对于的key不存在panic
func (c *Context) MustGet(key string) interface{} {
	if v, ok := c.Get(key); ok {
		return v
	}
	panic(fmt.Sprintf("key '%v' does not exist", key))
}
