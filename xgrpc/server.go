package xgrpc

import (
	"context"
	"sync"
)

// 初始grpc上下文每个服务携带的方法名标识，即context.WithValue()中设置值使用的key
const CtxWithMethodKey = "_XGrpc_Method"

var ctxPool sync.Pool

func init() {
	ctxPool.New = func() interface{} {
		return new(Context)
	}
}

type Server struct {
	// 方法名，用于服务唯一标识等
	method   string
	handlers []CtxHandlerFn
}

func NewServer(handlers ...CtxHandlerFn) *Server {
	s := &Server{handlers: handlers}
	return s
}

// WithMethod 设置服务方法名
func (s *Server) WithMethod(method string) *Server {
	s.method = method
	return s
}

// Serve 执行grpc服务，返回响应数据+错误
func (s *Server) ServeGRPC(ctx context.Context, req interface{}) (interface{}, error) {
	// 获取服务标识，若果有设置
	var method string
	if s.method != "" {
		// 有设置，优先
		method = s.method
	} else {
		method, _ = ctx.Value(CtxWithMethodKey).(string)
	}
	c := ctxPool.Get().(*Context)
	defer ctxPool.Put(c)
	c.reset(method, s.handlers...)
	return c.Next(req)
}

type ServerBuilder struct {
	// heads 保存头部处理器，最终使用Build构造Server实例
	heads []CtxHandlerFn
	// tails 保存尾部处理器，最终使用Build构造Server实例
	tails []CtxHandlerFn
}

// Group 返回一个新的ServerBuilder实例，它的heads和tails从现有的Group复制
func (b *ServerBuilder) Group() *ServerBuilder {
	newS := new(ServerBuilder)
	newS.heads = make([]CtxHandlerFn, len(b.heads))
	copy(newS.heads, b.heads)
	newS.tails = make([]CtxHandlerFn, len(b.tails))
	copy(newS.tails, b.tails)
	return newS
}

// Use 添加指定处理器handlers到heads中
func (b *ServerBuilder) Use(handlers ...CtxHandlerFn) *ServerBuilder {
	b.heads = append(b.heads, handlers...)
	return b
}

// UseAfter 添加指定处理器handlers到tails中
func (b *ServerBuilder) UseAfter(handlers ...CtxHandlerFn) *ServerBuilder {
	b.tails = append(b.tails, handlers...)
	return b
}

// Build 构造一个Server实例并返回，Server实例的处理器包括：b.heads + handlers +b.tails
func (b *ServerBuilder) Build(handlers ...CtxHandlerFn) *Server {
	chain := make([]CtxHandlerFn, len(b.heads)+len(handlers)+len(b.tails))
	copy(chain, b.heads)
	copy(chain[len(b.heads):], handlers)
	copy(chain[len(b.heads)+len(handlers):], b.tails)
	if len(chain) > abortIndex {
		panic("too many context handlers")
	}
	return NewServer(chain...)
}

// 默认全局ServerBuilder实例
var serverBuilder = new(ServerBuilder)

// Group 返回一个新的ServerBuilder实例，它的heads和tails从现有的serverBuilder复制
func Group() *ServerBuilder {
	return serverBuilder.Group()
}

// Use 添加指定处理器handlers到serverBuilder的heads中
func Use(handlers ...CtxHandlerFn) *ServerBuilder {
	return serverBuilder.Use(handlers...)
}

// UseAfter 添加指定处理器handlers到serverBuilder的tails中
func UseAfter(handlers ...CtxHandlerFn) *ServerBuilder {
	return serverBuilder.UseAfter(handlers...)
}

// Build 构造一个Server实例并返回，Server实例的处理器包括：serverBuilder.heads + handlers +serverBuilder.tails
func Build(handlers ...CtxHandlerFn) *Server {
	return serverBuilder.Build(handlers...)
}
