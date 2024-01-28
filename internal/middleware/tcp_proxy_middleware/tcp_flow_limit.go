package tcp_proxy_middleware

import (
	"fmt"
	"strings"

	"github.com/hugokung/micro_gateway/internal/dao"
	"github.com/hugokung/micro_gateway/pkg/public"
)

func TCPFlowLimitMiddleware() func(c *TcpSliceRouterContext) {
	return func(ctx *TcpSliceRouterContext) {
		serviceInterface:= ctx.Get("service")
		if serviceInterface == nil {
			ctx.conn.Write([]byte("service not found"))
			ctx.Abort()
			return
		}
		serviceDetail := serviceInterface.(*dao.ServiceDetail)
		if serviceDetail.AccessControl.ServiceFlowLimit > 0 {
			serviceLimiter, err := public.FlowLimiterHandler.GetLimiter(public.FlowServicePrefix+serviceDetail.Info.ServiceName, float64(serviceDetail.AccessControl.ServiceFlowLimit))
			if err != nil {
				ctx.conn.Write([]byte(err.Error()))
				ctx.Abort()
				return
			}
			if !serviceLimiter.Allow() {
				ctx.conn.Write([]byte(fmt.Sprintf("service flow limit %v", serviceDetail.AccessControl.ServiceFlowLimit)))
				ctx.Abort()
				return
			}
		}
		splits := strings.Split(ctx.conn.RemoteAddr().String(), ":")
		ClientIP := ""
		if len(splits) == 2 {
			ClientIP = splits[0]
		}
		if serviceDetail.AccessControl.ClientIPFlowLimit > 0 {
			clientLimiter, err := public.FlowLimiterHandler.GetLimiter(public.FlowServicePrefix+serviceDetail.Info.ServiceName+"_" + ClientIP, float64(serviceDetail.AccessControl.ClientIPFlowLimit))
			if err != nil {
				ctx.conn.Write([]byte(err.Error()))
				ctx.Abort()
				return
			}
			if !clientLimiter.Allow() {
				ctx.conn.Write([]byte(fmt.Sprintf("client ip flow limit %v", serviceDetail.AccessControl.ClientIPFlowLimit)))
				ctx.Abort()
				return
			}
		}
		ctx.Next()
	}
}
