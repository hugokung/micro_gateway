package test

// import (
// 	"context"
// 	"testing"
// 	"time"

// 	"github.com/hugokung/micro_gateway/internal/dao"
// 	"github.com/hugokung/micro_gateway/internal/load_balancer"
// 	"github.com/hugokung/micro_gateway/internal/middleware/tcp_proxy_middleware"
// 	"github.com/hugokung/micro_gateway/internal/server/tcp_server"
// )

// func TestTcpProxy(t *testing.T) {
// 	dsService := dao.ServiceDetail{
// 		Info: &dao.ServiceInfo{
// 			ID: 1,
// 			LoadType: 0,
// 			ServiceName: "DownStream",
// 			ServiceDesc: "---",
// 			CreatedAt: time.Now(),
// 			UpdatedAt: time.Now(),
// 		},
// 		TCPRule: &dao.TcpRule{
// 			ID: 1,
// 			ServiceID: 1,
// 			Port: 2003,
// 		},
// 		LoadBalance: &dao.LoadBalance{
// 			ID: 1,
// 			ServiceID: 1,
// 			RoundType: 0,
// 			IpList: "127.0.0.1:2004",
// 			WeightList: "50",
// 		},
// 		AccessControl: &dao.AccessControl{
// 			ID: 1,
// 			ServiceID: 1,
// 		},
// 	}

// 	rb, err := dao.LoadBalancerHandler.GetLoadBalancer(&dsService)
// 	if err != nil {
// 		t.Errorf("GetLoadBalancer got error: %v", err)
// 	}
// 	router := tcp_proxy_middleware.NewTcpSliceRouter()
// 	router.Group("/").Use()
// 	routerHandler := tcp_proxy_middleware.NewTcpSliceRouterHandler(
// 		func(c *tcp_proxy_middleware.TcpSliceRouterContext) tcp_server.TCPHandler {
// 			return load_balancer.NewTcpLoadBalanceReverseProxy(c, rb)
// 		}, router)
// 	baseCtx := context.WithValue(context.Background(), "service", &dsService)
// 	tcpServer := &tcp_server.TcpServer{
// 		Addr:        ":2003",
// 		Handler:     routerHandler,
// 		BaseCtx:     baseCtx,
// 		UpdateAt:    dsService.Info.UpdatedAt,
// 		ServiceName: dsService.Info.ServiceName,
// 	}
// 	go runProxy(tcpServer)
// 	var ch = make(chan struct{})

// 	<- ch
// 	defer tcpServer.Close()
// }

// func runProxy(proxy *tcp_server.TcpServer) {
// 	if err := proxy.ListenAndServe(); err != nil {
// 		return
// 	}
// 	// defer proxy.Close()
// }