package test

import (
	"fmt"
	"github.com/happyxcj/golib/xgrpc"
	"google.golang.org/grpc"
	"time"
)

func NewClient(cc *grpc.ClientConn) xgrpc.IClient {
	baseCli := xgrpc.NewBaseClient(cc)
	inner := xgrpc.NewServiceClient(baseCli, "pb.Test")
	timeoutCli := xgrpc.NewTimeoutClient(inner, time.Second).AddMethodTimeout("TestV2", time.Second*2)
	cli := xgrpc.NewDurClient(timeoutCli, func(method string, dur time.Duration) {
		fmt.Printf("method '%v' cost time: %v\n", method, dur)
	})
	return cli
}
