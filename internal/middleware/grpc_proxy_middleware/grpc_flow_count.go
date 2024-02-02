package grpc_proxy_middleware

import (
	"fmt"
	"log"
	"time"
	"github.com/hugokung/micro_gateway/internal/dao"
	"google.golang.org/grpc"
	"github.com/hugokung/micro_gateway/pkg/public"
)

func GrpcFlowCountMiddleware (serviceDetail *dao.ServiceDetail) func(interface{}, grpc.ServerStream, *grpc.StreamServerInfo, grpc.StreamHandler) error  {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {

		totalCounter, err := public.FlowCounterHandler.GetCounter(public.FlowTotal)
		if err != nil {
			return err
		}
		totalCounter.Increase()

		dayCount, _ := totalCounter.GetDayData(time.Now())
		fmt.Printf("totalCounter qps: %v, dayCount: %v", totalCounter.QPS, dayCount)

		serviceCounter, err := public.FlowCounterHandler.GetCounter(public.FlowCountServicePrefix + serviceDetail.Info.ServiceName)
		if err != nil {
			return err
		}
		serviceCounter.Increase()

		dayServiceCount, _ := serviceCounter.GetDayData(time.Now())
		fmt.Printf("totalCounter qps: %v, dayCount: %v", serviceCounter.QPS, dayServiceCount)
		
		if err := handler(srv, ss); err != nil {
			log.Printf("RPC failed with error %v\n", err)
			return err
		}
		return nil
	}
}