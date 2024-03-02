package grpc_proxy_middleware

import (
	"errors"

	"github.com/afex/hystrix-go/hystrix"
	"github.com/hugokung/micro_gateway/internal/dao"
	"google.golang.org/grpc"
)

func GrpcCircuitBreakMiddleware(serviceDetail *dao.ServiceDetail) func(interface{}, grpc.ServerStream, *grpc.StreamServerInfo, grpc.StreamHandler) error{
	return func(svr interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		if serviceDetail.CircuitConfig.NeedCircuit == 1 {
			err := hystrix.Do(serviceDetail.Info.ServiceName, func () error {
				if err := handler(svr, ss); err != nil {
					return err
				}
				return nil
			}, func (err error) error {
				if err != nil {
					return errors.New(serviceDetail.CircuitConfig.FallBackMsg)
				}
				return nil
			})
			return err
		}
		return nil
	}
}