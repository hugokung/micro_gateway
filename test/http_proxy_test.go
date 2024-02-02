package test

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hugokung/micro_gateway/internal/load_balancer"
	"github.com/hugokung/micro_gateway/pkg/load_balance"
)

func TestHttpProxy(t *testing.T) {
	//模拟下游服务
	backendServer := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			// w.Write([]byte("Hello from backend"))
			fmt.Fprintf(w, "Hello from backend")
		},
	))
	defer backendServer.Close()
	ipConf := map[string]string{}
	//设置下游节点
	ipConf[backendServer.URL[7:]] = "50"
	mConf, err := load_balance.NewLoadBalanceCheckConf("TestHttpProxy",
		fmt.Sprintf("%s%s", "http://", "%s"), ipConf)
	if err != nil {
		t.Errorf("load_balance.NewLoadBalanceCheckConf error: %v", err)
	}
	trans := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   time.Duration(30) * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConnsPerHost:   15,
		MaxIdleConns:          4000,
		WriteBufferSize:       1 << 18, //200m
		ReadBufferSize:        1 << 18, //200m
		IdleConnTimeout:       time.Duration(90) * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: time.Duration(30) * time.Second,
	}
	lb := load_balance.LoadBanlanceFactorWithConf(load_balance.LbType(0), mConf)
	//设置模拟请求
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/", nil)
	
	http_proxy_handler := load_balancer.NewLoadBalanceReverseProxy(c, lb, trans)
	//http代理的方法
	http_proxy_handler.ServeHTTP(w, c.Request)
	
	if w.Result().StatusCode != http.StatusOK {
		t.Errorf("Expected status OK, got %v", c.Writer.Status())
	}
	body, err := io.ReadAll(w.Result().Body)
	if err != nil {
		t.Errorf("io.ReadAll error: %v", err)
	}
	if string(body) != "Hello from backend" {
		t.Errorf("got %v, %v", len(string(body)), string(body))
	}
}
