package tcp_proxy_middleware

import (
	"fmt"
	"strings"

	"github.com/hugokung/micro_gateway/dao"
	"github.com/hugokung/micro_gateway/public"
)

func TCPWhileListMiddleware() func(c *TcpSliceRouterContext) {
	return func(ctx *TcpSliceRouterContext) {
		serviceInterface := ctx.Get("service")
		if serviceInterface == nil {
			ctx.conn.Write([]byte("service not found"))
			ctx.Abort()
			return
		}
		serviceDetail := serviceInterface.(*dao.ServiceDetail)
		var whileList []string
		if serviceDetail.AccessControl.WhiteList != "" {
			whileList = strings.Split(serviceDetail.AccessControl.WhiteList, ",")
		}
		splits := strings.Split(ctx.conn.RemoteAddr().String(), ":")
		ClientIP := ""
		if len(splits) == 2 {
			ClientIP = splits[0]
		}
		if serviceDetail.AccessControl.OpenAuth == 1 && len(whileList) > 0 {
			if !public.InStringSlice(whileList, ClientIP) {
				ctx.conn.Write([]byte(fmt.Sprintf("%s not in white ip list", ClientIP)))
				ctx.Abort()
				return
			}
		}
		ctx.Next()
	}
}
