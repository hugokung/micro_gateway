package tcp_proxy_router

import (
	"context"
	"fmt"
	"log"

	"github.com/hugokung/micro_gateway/dao"
	"github.com/hugokung/micro_gateway/reverse_proxy"
	"github.com/hugokung/micro_gateway/tcp_proxy_middleware"
	"github.com/hugokung/micro_gateway/tcp_server"
)

var tcpServerList []*tcp_server.TcpServer

func TcpServerRun() {
	//tcp代理程序启动前，需要获取已有的tcp服务信息。
	serviceList := dao.ServiceManagerHandler.GetTcpServiceList()

	for _, serviceItem := range serviceList {
		tmpItem := serviceItem
		log.Printf(" [INFO] Tcp_Proxy_Run:%v\n", tmpItem.TCPRule.Port)
		go func(serviceDetail *dao.ServiceDetail) {
			addr := fmt.Sprintf(":%d", serviceDetail.TCPRule.Port)

			//对于每个tcp服务，获取一个负载均衡器
			rb, err := dao.LoadBalancerHandler.GetLoadBalancer(serviceDetail)
			if err != nil {
				log.Fatalf("[INFO] GetTcpLoadBalancer err: %v %v\n", addr, err)
				return
			}

			//构建路由及设置中间件
			// counter, _ := public.NewFlowCountService("local_app", time.Second)
			router := tcp_proxy_middleware.NewTcpSliceRouter()
			router.Group("/").Use(
				tcp_proxy_middleware.TCPFlowCountMiddleware(),
				tcp_proxy_middleware.TCPFlowLimitMiddleware(),
				tcp_proxy_middleware.TCPWhileListMiddleware(),
				tcp_proxy_middleware.TCPBlacListMiddleware(),
			)

			//构建回调handler
			routerHandler := tcp_proxy_middleware.NewTcpSliceRouterHandler(
				func(c *tcp_proxy_middleware.TcpSliceRouterContext) tcp_server.TCPHandler {
					return reverse_proxy.NewTcpLoadBalanceReverseProxy(c, rb)
				}, router)

			baseCtx := context.WithValue(context.Background(), "service", serviceDetail)
			tcpServer := &tcp_server.TcpServer{
				Addr:    addr,
				Handler: routerHandler,
				BaseCtx: baseCtx,
			}
			if err := tcpServer.ListenAndServe(); err != nil && err != tcp_server.ErrServerClosed {
				log.Fatalf("[INFO] tcp_proxy_run err: %v\n", err)
			}
			tcpServerList = append(tcpServerList, tcpServer)
		}(tmpItem)
	}
}

func TcpServerStop() {
	for _, tcpServer := range tcpServerList {
		tcpServer.Close()
		log.Printf("[INFO] tcp_proxy_stop %v stopped\n", tcpServer.Addr)
	}
	log.Printf(" [INFO] Tcp_Proxy stopped\n")
}
