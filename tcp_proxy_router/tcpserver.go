package tcp_proxy_router

import (
	"context"
	"fmt"
	"github.com/hugokung/micro_gateway/public"
	"log"
	"sync"

	"github.com/hugokung/micro_gateway/dao"
	"github.com/hugokung/micro_gateway/reverse_proxy"
	"github.com/hugokung/micro_gateway/tcp_proxy_middleware"
	"github.com/hugokung/micro_gateway/tcp_server"
)

//var tcpServerList []*tcp_server.TcpServer

const (
	typeOfUpdate int = iota + 1
	typeOfOther
)

type TcpManager struct {
	ServerList []*tcp_server.TcpServer
}

func init() {
	TcpManagerHandler = NewTcpManager()
}

func NewTcpManager() *TcpManager {
	return &TcpManager{}
}

var TcpManagerHandler *TcpManager

func (t *TcpManager) tcpServerRunOnce(service *dao.ServiceDetail, tp int) {
	addr := fmt.Sprintf(":%d", service.TCPRule.Port)
	rb, err := dao.LoadBalancerHandler.GetLoadBalancer(service)
	if err != nil {
		log.Fatalf(" [INFO] GetTcpLoadBalancer %v err:%v\n", addr, err)
		return
	}

	//构建路由及设置中间件
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

	baseCtx := context.WithValue(context.Background(), "service", service)
	tcpServer := &tcp_server.TcpServer{
		Addr:        addr,
		Handler:     routerHandler,
		BaseCtx:     baseCtx,
		UpdateAt:    service.Info.UpdatedAt,
		ServiceName: service.Info.ServiceName,
	}

	if tp != typeOfUpdate {
		t.ServerList = append(t.ServerList, tcpServer)
	} else {
		for _, sl := range t.ServerList {
			if sl.ServiceName == service.Info.ServiceName {
				sl = tcpServer
				break
			}
		}
	}

	log.Printf(" [INFO] tcp_proxy_run %v\n", addr)
	if err := tcpServer.ListenAndServe(); err != nil && err != tcp_server.ErrServerClosed {
		log.Printf(" [INFO] tcp_proxy_run %v err:%v\n", addr, err)
	}
}

func (t *TcpManager) TcpServerRun() {
	//tcp代理程序启动前，需要获取已有的tcp服务信息。
	serviceList := dao.ServiceManagerHandler.GetTcpServiceList()

	for _, serviceItem := range serviceList {
		tmpItem := serviceItem
		log.Printf(" [INFO] Tcp_Proxy_Run:%v\n", tmpItem.TCPRule.Port)
		go func(serviceDetail *dao.ServiceDetail) {
			t.tcpServerRunOnce(serviceDetail, typeOfOther)
		}(tmpItem)
	}
	dao.ServiceManagerHandler.Register(t)
}

func (t *TcpManager) Update(e *dao.ServiceEvent) {
	log.Printf("TcpManager.Update")
	delList := e.DeleteService
	for _, delService := range delList {
		if delService.Info.LoadType == public.LoadTypeTCP {
			continue
		}
		for _, tcpServer := range TcpManagerHandler.ServerList {
			if delService.Info.ServiceName != tcpServer.ServiceName {
				continue
			}
			tcpServer.Close()
			log.Printf(" [INFO] tcp_proxy_stop %v stopped\n", tcpServer.Addr)
		}
	}
	addList := e.AddService
	for _, addService := range addList {
		if addService.Info.LoadType != public.LoadTypeTCP {
			continue
		}
		go t.tcpServerRunOnce(addService, typeOfOther)
	}
	updateList := e.UpdateService
	for _, updateService := range updateList {
		if updateService.Info.LoadType != public.LoadTypeTCP {
			continue
		}
		for _, tcpServer := range TcpManagerHandler.ServerList {
			if updateService.Info.ServiceName != tcpServer.ServiceName {
				continue
			}
			tcpServer.Close()
			log.Printf(" [INFO] tcp_proxy_stop %v stopped\n", tcpServer.Addr)
			break
		}
	}
	for _, updateService := range updateList {
		if updateService.Info.LoadType != public.LoadTypeTCP {
			continue
		}
		go t.tcpServerRunOnce(updateService, typeOfUpdate)
	}
}

func (t *TcpManager) TcpServerStop() {
	for _, tcpServer := range t.ServerList {
		wait := sync.WaitGroup{}
		wait.Add(1)
		go func() {
			defer func() {
				wait.Done()
				if err := recover(); err != nil {
					log.Println(err)
				}
			}()
			tcpServer.Close()
		}()
		log.Printf(" [INFO] tcp_proxy_stop %v stopped\n", tcpServer.Addr)
	}
}
