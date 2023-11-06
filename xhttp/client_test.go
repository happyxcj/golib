package xhttp

import (
	. "github.com/smartystreets/goconvey/convey"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

var (
	_testBody      = "test body"
	_testHeaderKey = "test_h_key"
	_testHeaderVal = "test_h_val"
)

func _do(cli IClient, url string) (string, error) {
	req, err := NewReq(http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	resp, err := cli.Do(req)
	if err != nil {
		return "", err
	}
	respMsg := new(string)
	err = DecodeStrResp(resp.Body, respMsg)
	if err != nil {
		return "", err
	}
	return *respMsg, nil
}

func TestBaseClient(t *testing.T) {
	Convey("TestBaseClient", t, func() {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(_testBody))
		}))
		cli := NewDefaultBaseClient()
		resp, _ := _do(cli, server.URL)
		So(resp, ShouldEqual, _testBody)
	})
}

func TestHeaderClient(t *testing.T) {
	Convey("TestHeaderClient", t, func() {
		var hVal string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			hVal = r.Header.Get(_testHeaderKey)
			w.Write([]byte(_testBody))
		}))
		inner := NewDefaultBaseClient()
		cli := NewHeaderClient(inner, nil).AddKV(_testHeaderKey, _testHeaderVal)
		resp, _ := _do(cli, server.URL)
		So(resp, ShouldEqual, _testBody)
		So(hVal, ShouldEqual, _testHeaderVal)
	})
}

func TestTimeoutClient(t *testing.T) {
	Convey("TestTimeoutClient", t, func() {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(time.Millisecond * 30)
			w.Write([]byte(_testBody))
		}))
		inner := NewDefaultBaseClient()
		cli := NewTimeoutClient(inner, time.Millisecond*10).AddPathTimeouts(time.Millisecond*50, "/test_path")
		resp, _ := _do(cli, server.URL)
		So(resp, ShouldEqual, "")
		resp, _ = _do(cli, server.URL+"/test_path?a=3")
		So(resp, ShouldEqual, _testBody)
	})
}

func TestDurClient(t *testing.T) {
	Convey("TestDurClient", t, func() {
		var times int64
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			nowTimes := atomic.AddInt64(&times, 1)
			if nowTimes <= 2 {
				// 超时
				time.Sleep(time.Second)
				w.Write([]byte(_testBody))
				return
			}
			if nowTimes%2 == 0 {
				time.Sleep(time.Millisecond * 500)
			}
			w.Write([]byte(_testBody))
		}))
		inner := NewDefaultBaseClient()
		innerCli := NewTimeoutClient(inner, time.Second)
		var slowReqCount int64
		cli := NewDurClient(innerCli, func(dur time.Duration) {
			if dur >= time.Millisecond*500 {
				atomic.AddInt64(&slowReqCount, 1)
			}
		})
		var wg sync.WaitGroup
		var errCount int64
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_, err := _do(cli, server.URL)
				if err != nil {
					atomic.AddInt64(&errCount, 1)
				}
			}()
		}
		wg.Wait()
		So(atomic.LoadInt64(&slowReqCount), ShouldEqual, 6)
		So(atomic.LoadInt64(&errCount), ShouldEqual, 2)
	})
}

type _TestReq struct {
	Name string `json:"name"`
}

type _TestResp struct {
	Age int `json:"age"`
}

func TestMethodClient(t *testing.T) {
	Convey("TestMethodClient", t, func() {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`{"age":5}`))
		}))
		inner := NewDefaultBaseClient()
		innerCli := NewHeaderClient(inner, nil).AddKV("Content-Type", "application/json")
		cli := NewMethodJsonClient(innerCli).WithBaseUrl(server.URL)
		_, err := cli.Get("/test_get")
		if err != nil {
			t.Fatalf("Get err: %v", err)
		}
		req := &_TestReq{Name: "happyxcj"}
		resp := new(_TestResp)
		if err := cli.PostAndDecode("/test_post", req, resp); err != nil {
			t.Fatalf("PostAndDecode err: %v", err)
		}
		So(resp.Age, ShouldEqual, 5)
	})
}
