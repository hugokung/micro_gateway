package grpc_proxy_router

import (
	"fmt"
	"log"
	"net"

	"github.com/e421083458/grpc-proxy/proxy"
	"github.com/hugokung/micro_gateway/dao"
	"github.com/hugokung/micro_gateway/grpc_proxy_middleware"
	"github.com/hugokung/micro_gateway/reverse_proxy"
	"google.golang.org/grpc"
)

var grpcServerList = []*warpGrpcServer{}

type warpGrpcServer struct {
	Addr string
	*grpc.Server
}

func GrpcServerRun() {
	serviceList := dao.ServiceManagerHandler.GetGrpcServiceList()
	for _, serviceItem := range serviceList {
		tmpItem := serviceItem
		log.Printf(" [INFO] Grpc_Proxy_Run:%v\n", tmpItem.GRPCRule.Port)
		go func(serviceDetail *dao.ServiceDetail) {
			addr := fmt.Sprintf(":%d", serviceDetail.GRPCRule.Port)
			lis, err := net.Listen("tcp", addr)
			if err != nil {
				log.Fatalf("[INFO] GrpcListen %v err:%v\n: ", addr, err)
				return
			}
			rb, err := dao.LoadBalancerHandler.GetLoadBalancer(serviceDetail)
			if err != nil {
				log.Fatalf("[INFO] GetGrpcLoadBalancer err: %v %v\n", addr, err)
				return
			}
			grpcHandler := reverse_proxy.NewGrpcLoadBalanceHandler(rb)
			s := grpc.NewServer(
				grpc.ChainStreamInterceptor(
					grpc_proxy_middleware.GrpcFlowCountMiddleware(serviceDetail),
					grpc_proxy_middleware.GrpcFlowLimitMiddleware(serviceDetail),
					grpc_proxy_middleware.GrpcJwtAuthTokenMiddleware(serviceDetail),
					grpc_proxy_middleware.GrpcJwtFlowCountMiddleware(serviceDetail),
					grpc_proxy_middleware.GrpcJwtFlowLimitMiddleware(serviceDetail),
					grpc_proxy_middleware.GrpcWhileListMiddleware(serviceDetail),
					grpc_proxy_middleware.GrpcBlacListMiddleware(serviceDetail),
					grpc_proxy_middleware.GrpcHeaderTransferMiddleware(serviceDetail),
				),
				grpc.CustomCodec(proxy.Codec()),
				grpc.UnknownServiceHandler(grpcHandler)) //自定义全局回调
			grpcServerList = append(grpcServerList, &warpGrpcServer{
				Addr: addr,
				Server: s,
			})
			fmt.Printf("server listening at %v\n", lis.Addr())
			if err := s.Serve(lis); err != nil {
				log.Fatalf("[INFO] grpc_proxy_run err : %v, port: %v", err, addr)
			}
			
		}(tmpItem)
	}
}

func GrpcServerStop() {
	for _, grpcServer := range grpcServerList {
		grpcServer.Server.GracefulStop()
		log.Printf("[INFO] grpc_proxy_stop %v stopped\n", grpcServer.Addr)
	}
	log.Printf(" [INFO] Grpc_Proxy stopped\n")
}
