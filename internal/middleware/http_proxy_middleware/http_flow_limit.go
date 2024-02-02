package http_proxy_middleware

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/hugokung/micro_gateway/internal/dao"
	"github.com/hugokung/micro_gateway/pkg/public"
	"github.com/hugokung/micro_gateway/pkg/response"
	"github.com/pkg/errors"
)

func HTTPFlowLimitMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		serviceInterface, ok := ctx.Get("service")
		if !ok {
			response.ResponseError(ctx, 2001, errors.New("service not found"))
			ctx.Abort()
			return
		}
		serviceDetail := serviceInterface.(*dao.ServiceDetail)
		if serviceDetail.AccessControl.ServiceFlowLimit > 0 {
			serviceLimiter, err := public.FlowLimiterHandler.GetLimiter(public.FlowServicePrefix+serviceDetail.Info.ServiceName, float64(serviceDetail.AccessControl.ServiceFlowLimit))
			if err != nil {
				response.ResponseError(ctx, 5001, err)
				ctx.Abort()
				return
			}
			if !serviceLimiter.Allow() {
				response.ResponseError(ctx, 5002, errors.New(fmt.Sprintf("service flow limit %v", serviceDetail.AccessControl.ServiceFlowLimit)))
				ctx.Abort()
				return
			}
		}

		if serviceDetail.AccessControl.ClientIPFlowLimit > 0 {
			clientLimiter, err := public.FlowLimiterHandler.GetLimiter(public.FlowServicePrefix+serviceDetail.Info.ServiceName+"_"+ctx.ClientIP(), float64(serviceDetail.AccessControl.ClientIPFlowLimit))
			if err != nil {
				response.ResponseError(ctx, 5001, err)
				ctx.Abort()
				return
			}
			if !clientLimiter.Allow() {
				response.ResponseError(ctx, 5003, errors.New(fmt.Sprintf("client ip flow limit %v", serviceDetail.AccessControl.ClientIPFlowLimit)))
				ctx.Abort()
				return
			}
		}
		ctx.Next()
	}
}
