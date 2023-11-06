package xhttp

import (
	"net"
	"time"
)

// 拨号选项方法
type DialOptFn func(d *net.Dialer)

// NewDialer 新建Dialer实例，
// 默认：拨号超时30秒；连接存活时间30秒
// 使用指定的optFs来设置不同的自定义拨号选项
func NewDialer(optFns ...DialOptFn) *net.Dialer {
	dialer := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}
	for _, optFn := range optFns {
		optFn(dialer)
	}
	return dialer
}

// DialTimeout 返回对指定d设置Timeout的选项方法
func DialWithTimeout(timeout time.Duration) DialOptFn {
	return func(d *net.Dialer) {
		d.Timeout = timeout
	}
}

// DialTimeout 返回对指定d设置Timeout的选项方法
func DialWithKeepAlive(keepAlive time.Duration) DialOptFn {
	return func(d *net.Dialer) {
		d.KeepAlive = keepAlive
	}
}
