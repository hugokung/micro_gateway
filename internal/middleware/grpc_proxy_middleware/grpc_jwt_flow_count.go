package grpc_proxy_middleware

import (
	"encoding/json"
	"fmt"
	"log"
	"github.com/hugokung/micro_gateway/internal/dao"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"github.com/hugokung/micro_gateway/pkg/public"
)

func GrpcJwtFlowCountMiddleware(serviceDetail *dao.ServiceDetail) func(interface{}, grpc.ServerStream, *grpc.StreamServerInfo, grpc.StreamHandler) error  {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {

		md, ok := metadata.FromIncomingContext(ss.Context())
		if !ok {
			if err := handler(srv, ss); err != nil {
				log.Printf("RPC failed with error %v\n", err)
				return err
			}
			return nil
		}
		appInfoStr := md.Get("app")
		appInfo := &dao.App{}
		if err := json.Unmarshal([]byte(appInfoStr[0]), appInfo); err != nil {
			return err
		}
		appCounter, err := public.FlowCounterHandler.GetCounter(public.FlowCountAppPrefix + appInfo.AppID)
		if err != nil {
			return err
		}
		appCounter.Increase()
		if appInfo.Qpd > 0 && appCounter.TotalCount > appInfo.Qpd {

			return errors.New(fmt.Sprintf("租户日请求量限流 limit:%v current:%v",appInfo.Qpd,appCounter.TotalCount))
		}
		if err := handler(srv, ss); err != nil {
			log.Printf("RPC failed with error %v\n", err)
			return err
		}
		return nil
	}
}