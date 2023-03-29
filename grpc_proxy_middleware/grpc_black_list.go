package grpc_proxy_middleware

import (
	"fmt"
	"log"
	"strings"

	"github.com/hugokung/micro_gateway/dao"
	"github.com/hugokung/micro_gateway/public"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
)

func GrpcBlacListMiddleware(serviceDetail *dao.ServiceDetail) func (interface{}, grpc.ServerStream, *grpc.StreamServerInfo, grpc.StreamHandler) error  {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {

		whileList := []string{}
		blackList := []string{}

		if serviceDetail.AccessControl.BlackList != "" {
			blackList = strings.Split(serviceDetail.AccessControl.BlackList, ",")
		}
		if serviceDetail.AccessControl.WhiteList != "" {
			whileList = strings.Split(serviceDetail.AccessControl.WhiteList, ",")
		}
		peerCtx, ok := peer.FromContext(ss.Context())
		if !ok {
			return errors.New("peer not found with context")
		}
		peerAddr:=peerCtx.Addr.String()
		addrPos:=strings.LastIndex(peerAddr,":")
		clientIP:=peerAddr[0: addrPos]
		if serviceDetail.AccessControl.OpenAuth == 1 && len(whileList) == 0 && len(blackList) > 0 {
			if public.InStringSlice(blackList, clientIP) {
				return errors.New(fmt.Sprintf("%s in black ip list", clientIP))
			}
		}
		err := handler(srv, ss)
		if err != nil {
			log.Printf("RPC failed with error %v\n", err)
			return err
		}
		return nil
	}
}