package test

import (
	"context"
	"fmt"
	"github.com/happyxcj/golib/xgrpc"
	"github.com/happyxcj/golib/xgrpc/_test/pb"
	"time"
)

type testServer struct {
	testS   *xgrpc.Server
	testV2S *xgrpc.Server
	sb1     *xgrpc.ServerBuilder
}

type testService struct{}

func (testService) Test(req *pb.TestReq) (*pb.TestResp, error) {
	if req.B <= 1000 {
		return &pb.TestResp{V: "value 1"}, nil
	}
	return &pb.TestResp{V: "value 2"}, nil
}

func (testService) TestV2(req *pb.TestReqV2) (*pb.TestRespV2, error) {
	time.Sleep(time.Second)
	if req.B <= 1000 {
		return &pb.TestRespV2{V: req.A + " value 1"}, nil
	}
	return &pb.TestRespV2{V: req.A + " value 2"}, nil
}

func NewTestServer() *testServer {
	ss := new(testService)
	sb := xgrpc.Use(logSlowReqDelay)
	// 记录请求参数组
	sb1 := sb.Group().Use(logReqParam)
	// 记录请求参数和响应数据组
	sb2 := sb1.Group().UseAfter(logRespParam)
	s := &testServer{
		testS: sb1.Build(func(ctx *xgrpc.Context, req interface{}) (interface{}, error) {
			return ss.Test(req.(*pb.TestReq))
		}).WithMethod("Special Test"),
		testV2S: sb2.Build(func(ctx *xgrpc.Context, req interface{}) (interface{}, error) {
			return ss.TestV2(req.(*pb.TestReqV2))
		}),
	}
	return s
}

// 记录慢请求中间件
func logSlowReqDelay(c *xgrpc.Context, req interface{}) (interface{}, error) {
	start := time.Now()
	defer func() {
		delay := time.Since(start)
		if delay >= time.Second*1 {
			fmt.Printf("service '%v' slow req: %v\n", c.Method, delay)
		}
	}()
	return c.Next(req)
}

// 记录请求参数
func logReqParam(c *xgrpc.Context, req interface{}) (interface{}, error) {
	fmt.Printf("service '%v' req param: %v\n", c.Method, req)
	return req, nil
}

func logRespParam(c *xgrpc.Context, req interface{}) (interface{}, error) {
	fmt.Printf("service '%v' resp data: %v, is err: %v\n", c.Method, req, c.IsErr())
	return req, nil
}

func (s *testServer) Test(ctx context.Context, req *pb.TestReq) (*pb.TestResp, error) {
	response, err := s.testS.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return response.(*pb.TestResp), nil
}

func (s *testServer) TestV2(ctx context.Context, req *pb.TestReqV2) (*pb.TestRespV2, error) {
	response, err := s.testV2S.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return response.(*pb.TestRespV2), nil
}
