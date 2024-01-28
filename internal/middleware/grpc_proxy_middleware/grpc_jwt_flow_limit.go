package grpc_proxy_middleware

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"github.com/hugokung/micro_gateway/internal/dao"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"github.com/hugokung/micro_gateway/pkg/public"
)

func GrpcJwtFlowLimitMiddleware (serviceDetail *dao.ServiceDetail) func(interface{}, grpc.ServerStream, *grpc.StreamServerInfo, grpc.StreamHandler) error  {
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
		peerCtx, ok := peer.FromContext(ss.Context())
		if !ok {
			return errors.New("peer not found with context")
		}
		peerAddr := peerCtx.Addr.String()
		addrPos := strings.LastIndex(peerAddr,":")
		clientIP := peerAddr[0:addrPos]
		if appInfo.Qps > 0 {
			clientLimiter, err := public.FlowLimiterHandler.GetLimiter(public.FlowAppPrefix + appInfo.AppID + "_" + clientIP, float64(appInfo.Qps))
			if err != nil {
				return err
			}
			if !clientLimiter.Allow() {
				return errors.New(fmt.Sprintf("%v flow limit %v", clientIP, appInfo.Qps))
			}
		}
		if err := handler(srv, ss); err != nil {
			log.Printf("RPC failed with error %v\n", err)
			return err
		}
		return nil
	}
}