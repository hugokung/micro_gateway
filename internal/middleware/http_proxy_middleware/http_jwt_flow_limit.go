package http_proxy_middleware

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/hugokung/micro_gateway/internal/dao"
	"github.com/hugokung/micro_gateway/pkg/public"
	"github.com/hugokung/micro_gateway/pkg/response"
	"github.com/pkg/errors"
)

func HTTPJwtFlowLimitMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		appInterface, ok := ctx.Get("app")
		if !ok {
			ctx.Next()
			return
		}
		appInfo := appInterface.(*dao.App)
		if appInfo.Qps > 0 {
			clientLimiter, err := public.FlowLimiterHandler.GetLimiter(public.FlowAppPrefix + appInfo.AppID + "_" + ctx.ClientIP(), float64(appInfo.Qps))
			if err != nil {
				response.ResponseError(ctx, 20001, err)
				ctx.Abort()
				return
			}
			if !clientLimiter.Allow() {
				response.ResponseError(ctx, 20002, errors.New(fmt.Sprintf("%v flow limit %v", ctx.ClientIP(), appInfo.Qps)))
				ctx.Abort()
				return
			}
		}
		ctx.Next()
	}
}