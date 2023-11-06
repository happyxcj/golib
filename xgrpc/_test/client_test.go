package test

import (
	"context"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/happyxcj/golib/xgrpc"
	"github.com/happyxcj/golib/xgrpc/_test/pb"
	"google.golang.org/grpc"
	"net"
	"testing"
)

const (
	_hostPort string = "localhost:8002"
)

func TestGRPCClient(t *testing.T) {
	Convey("TestBaseClient", t, func() {
		sc, err := net.Listen("tcp", _hostPort)
		if err != nil {
			t.Fatalf("unable to listen: %+v", err)
		}
		opt := grpc.UnaryInterceptor(func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
			ctx = context.WithValue(ctx, xgrpc.CtxWithMethodKey, info.FullMethod)
			return handler(ctx, req)
		})
		server := grpc.NewServer(opt)
		defer server.GracefulStop()

		srv := NewTestServer()
		go func() {
			pb.RegisterTestServer(server, srv)
			_ = server.Serve(sc)
		}()

		cc, err := grpc.Dial(_hostPort, grpc.WithInsecure())
		if err != nil {
			t.Fatalf("failed to Dial: %+v", err)
		}

		client := NewClient(cc)

		req := &pb.TestReq{A: "22", B: 242}
		resp := new(pb.TestResp)
		if err := client.Invoke(context.Background(), "Test", req, resp); err != nil {
			t.Fatalf("client Invoke err: %v", err)
		}
		So(resp.V, ShouldEqual, "value 1")

		reqV2 := &pb.TestReqV2{A: "xcj", B: 242}
		respV2 := new(pb.TestRespV2)
		if err := client.Invoke(context.Background(), "TestV2", reqV2, respV2); err != nil {
			t.Fatalf("client Invoke err: %v", err)
		}
		So(respV2.V, ShouldEqual, "xcj value 1")
	})
}
