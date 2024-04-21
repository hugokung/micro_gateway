package http_proxy_middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/hugokung/micro_gateway/internal/dao"
	"github.com/hugokung/micro_gateway/internal/load_balancer"
	"github.com/hugokung/micro_gateway/pkg/load_balance"
	"github.com/hugokung/micro_gateway/pkg/response"
	"github.com/pkg/errors"
)

func HTTPReverseProxyMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		serviceInterface, ok := ctx.Get("service")
		if !ok {
			response.ResponseError(ctx, 2001, errors.New("service not found"))
			ctx.Abort()
			return
		}
		serviceDetail := serviceInterface.(*dao.ServiceDetail)
		var lb load_balance.LoadBalance
		var err error
		lb, err = dao.LoadBalancerHandler.GetLoadBalancer(serviceDetail)
		if err != nil {
			response.ResponseError(ctx, 2002, err)
			ctx.Abort()
			return
		}
		trans, err := dao.TransportorHandler.GetTransportor(serviceDetail)
		if err != nil {
			response.ResponseError(ctx, 2003, err)
			ctx.Abort()
			return
		}
		proxy := load_balancer.NewLoadBalanceReverseProxy(ctx, lb, trans)
		proxy.ServeHTTP(ctx.Writer, ctx.Request)
	}
}
