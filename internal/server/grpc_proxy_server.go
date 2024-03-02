package server

import (
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/e421083458/grpc-proxy/proxy"
	"github.com/hugokung/micro_gateway/internal/dao"
	"github.com/hugokung/micro_gateway/internal/load_balancer"
	"github.com/hugokung/micro_gateway/internal/middleware/grpc_proxy_middleware"
	"github.com/hugokung/micro_gateway/pkg/public"
	"google.golang.org/grpc"
)


type GrpcManager struct {
	ServerList []*warpGrpcServer
}

type warpGrpcServer struct {
	Addr        string
	ServiceName string
	UpdateAt    time.Time
	*grpc.Server
}

func NewGrpcManager() *GrpcManager {
	return &GrpcManager{}
}

var GrpcManagerHandler *GrpcManager

func init() {
	GrpcManagerHandler = NewGrpcManager()
}

func (g *GrpcManager) grpcServerRunOnce(service *dao.ServiceDetail, tp int) {
	addr := fmt.Sprintf(":%d", service.GRPCRule.Port)
	rb, err := dao.LoadBalancerHandler.GetLoadBalancer(service)
	if err != nil {
		log.Printf(" [ERROR] GetTcpLoadBalancer %v err:%v\n", addr, err)
		return
	}
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Printf(" [ERROR] GrpcListen %v err:%v\n", addr, err)
		return
	}
	grpcHandler := load_balancer.NewGrpcLoadBalanceHandler(rb)
	s := grpc.NewServer(
		grpc.ChainUnaryInterceptor(),
		grpc.ChainStreamInterceptor(
			grpc_proxy_middleware.GrpcFlowCountMiddleware(service),
			grpc_proxy_middleware.GrpcFlowLimitMiddleware(service),
			grpc_proxy_middleware.GrpcCircuitBreakMiddleware(service),
			grpc_proxy_middleware.GrpcJwtAuthTokenMiddleware(service),
			grpc_proxy_middleware.GrpcJwtFlowCountMiddleware(service),
			grpc_proxy_middleware.GrpcJwtFlowLimitMiddleware(service),
			grpc_proxy_middleware.GrpcWhileListMiddleware(service),
			grpc_proxy_middleware.GrpcBlacListMiddleware(service),
			grpc_proxy_middleware.GrpcHeaderTransferMiddleware(service),
		),
		grpc.CustomCodec(proxy.Codec()),
		grpc.UnknownServiceHandler(grpcHandler))
	if tp != typeOfUpdate {
		GrpcManagerHandler.ServerList = append(GrpcManagerHandler.ServerList, &warpGrpcServer{
			Addr:        addr,
			ServiceName: service.Info.ServiceName,
			UpdateAt:    service.Info.UpdatedAt,
			Server:      s,
		})
	} else {
		for i, sl := range g.ServerList {
			if sl.ServiceName == service.Info.ServiceName {
				g.ServerList[i] = &warpGrpcServer{
					Addr:        addr,
					ServiceName: service.Info.ServiceName,
					UpdateAt:    service.Info.UpdatedAt,
					Server:      s,
				}
				break
			}
		}
	}

	log.Printf(" [INFO] grpc_proxy_run %v\n", addr)
	if err := s.Serve(lis); err != nil {
		log.Printf(" [INFO] grpc_proxy_run %v err:%v\n", addr, err)
	}
}

func (g *GrpcManager) GrpcServerRun() {
	serviceList := dao.ServiceManagerHandler.GetGrpcServiceList()
	for _, serviceItem := range serviceList {
		tmpItem := serviceItem
		// g.grpcServerRunOnce(tmpItem, typeOfOther)
		log.Printf(" [INFO] Grpc_Proxy_Run:%v\n", tmpItem.GRPCRule.Port)
		go func(serviceDetail *dao.ServiceDetail) {
			g.grpcServerRunOnce(serviceDetail, typeOfOther)
		}(tmpItem)
	}
	dao.ServiceManagerHandler.Register(g)
}

func (g *GrpcManager) Update(e *dao.ServiceEvent) {
	log.Printf("GrpcManager.Update")
	delList := e.DeleteService
	for _, delService := range delList {
		if delService.Info.LoadType != public.LoadTypeGRPC {
			continue
		}
		for _, grpcServer := range GrpcManagerHandler.ServerList {
			if delService.Info.ServiceName != grpcServer.ServiceName {
				continue
			}
			grpcServer.GracefulStop()
			log.Printf(" [INFO] grpc_proxy_stop %v stopped\n", grpcServer.Addr)
		}
	}
	addList := e.AddService
	for _, addService := range addList {
		if addService.Info.LoadType != public.LoadTypeGRPC {
			continue
		}
		go g.grpcServerRunOnce(addService, typeOfOther)
	}
	updateList := e.UpdateService
	for _, updateService := range updateList {
		if updateService.Info.LoadType != public.LoadTypeGRPC {
			continue
		}
		for _, grpcServer := range GrpcManagerHandler.ServerList {
			if grpcServer.ServiceName != updateService.Info.ServiceName {
				continue
			}
			wait := sync.WaitGroup{}
			wait.Add(1)
			go func() {
				defer func() {
					wait.Done()
					if err := recover(); err != nil {
						log.Println(err)
					}
				}()
				grpcServer.GracefulStop()
			}()
			wait.Wait()
			log.Printf(" [INFO] grpc_proxy_stop %v stopped\n", grpcServer.Addr)
			// break
		}
	}
	for _, updateService := range updateList {
		if updateService.Info.LoadType != public.LoadTypeGRPC {
			continue
		}
		go g.grpcServerRunOnce(updateService, typeOfUpdate)
	}
}

func (g *GrpcManager) GrpcServerStop() {
	for _, grpcServer := range GrpcManagerHandler.ServerList {
		wait := sync.WaitGroup{}
		wait.Add(1)
		go func() {
			defer func() {
				wait.Done()
				if err := recover(); err != nil {
					log.Println(err)
				}
			}()
			grpcServer.GracefulStop()
		}()
		wait.Wait()
		log.Printf(" [INFO] grpc_proxy_stop %v stopped\n", grpcServer.Addr)
	}
}
