package http_proxy_middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/hugokung/micro_gateway/dao"
	"github.com/hugokung/micro_gateway/middleware"
	"github.com/hugokung/micro_gateway/reverse_proxy"
	"github.com/pkg/errors"
)

func HTTPReverseProxyMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		serviceInterface, ok := ctx.Get("service")
		if !ok {
			middleware.ResponseError(ctx, 2001, errors.New("service not found"))
			ctx.Abort()
			return
		}
		serviceDetail := serviceInterface.(*dao.ServiceDetail)
		lb, err := dao.LoadBalancerHandler.GetLoadBalancer(serviceDetail)
		if err != nil {
			middleware.ResponseError(ctx, 2002, err)
			ctx.Abort()
			return
		}
		trans, err := dao.TransportorHandler.GetTransportor(serviceDetail)
		if err != nil {
			middleware.ResponseError(ctx, 2003, err)
			ctx.Abort()
			return
		}
		proxy := reverse_proxy.NewLoadBalanceReverseProxy(ctx, lb, trans)
		proxy.ServeHTTP(ctx.Writer, ctx.Request)
	}
}