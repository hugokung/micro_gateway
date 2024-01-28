package http_proxy_middleware

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/hugokung/micro_gateway/internal/dao"
	"github.com/hugokung/micro_gateway/pkg/public"
	"github.com/hugokung/micro_gateway/pkg/response"
	"github.com/pkg/errors"
)

func HTTPJwtFlowCountMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		appInterface, ok := ctx.Get("app")
		if !ok {
			ctx.Next()
			return
		}
		appInfo := appInterface.(*dao.App)
		appCounter, err := public.FlowCounterHandler.GetCounter(public.FlowCountAppPrefix + appInfo.AppID)
		if err != nil {
			response.ResponseError(ctx, 20001, err)
			ctx.Abort()
			return
		}
		appCounter.Increase()
		if appInfo.Qpd > 0 && appCounter.TotalCount > appInfo.Qpd {
			response.ResponseError(ctx, 20002, errors.New(fmt.Sprintf("租户日请求量限流 limit:%v current:%v",appInfo.Qpd,appCounter.TotalCount)))
			ctx.Abort()
			return
		}
		ctx.Next()
	}
}