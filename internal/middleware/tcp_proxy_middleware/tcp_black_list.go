package tcp_proxy_middleware

import (
	"fmt"
	"strings"

	"github.com/hugokung/micro_gateway/internal/dao"
	"github.com/hugokung/micro_gateway/pkg/public"
)

func TCPBlacListMiddleware() func(c *TcpSliceRouterContext) {
	return func(ctx *TcpSliceRouterContext) {
		serviceInterface := ctx.Get("service")
		if serviceInterface == nil {
			ctx.conn.Write([]byte("service not found"))
			ctx.Abort()
			return
		}
		serviceDetail := serviceInterface.(*dao.ServiceDetail)
		var whileList []string
		var blackList []string

		if serviceDetail.AccessControl.BlackList != "" {
			blackList = strings.Split(serviceDetail.AccessControl.BlackList, ",")
		}
		if serviceDetail.AccessControl.WhiteList != "" {
			whileList = strings.Split(serviceDetail.AccessControl.WhiteList, ",")
		}
		splits := strings.Split(ctx.conn.RemoteAddr().String(), ":")
		ClientIP := ""
		if len(splits) == 2 {
			ClientIP = splits[0]
		}
		if serviceDetail.AccessControl.OpenAuth == 1 && len(whileList) == 0 && len(blackList) > 0 {
			if public.InStringSlice(blackList, ClientIP) {
				ctx.conn.Write([]byte(fmt.Sprintf("%s in black ip list", ClientIP)))
				ctx.Abort()
				return
			}
		}
		ctx.Next()
	}
}
