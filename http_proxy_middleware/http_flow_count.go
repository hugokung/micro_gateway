package http_proxy_middleware

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hugokung/micro_gateway/dao"
	"github.com/hugokung/micro_gateway/middleware"
	"github.com/hugokung/micro_gateway/public"
	"github.com/pkg/errors"
)

func HTTPFlowCountMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		serviceInterface, ok := ctx.Get("service")
		if !ok {
			middleware.ResponseError(ctx, 2001, errors.New("service not found"))
			ctx.Abort()
			return
		}
		serviceDetail := serviceInterface.(*dao.ServiceDetail)
		totalCounter, err := public.FlowCounterHandler.GetCounter(public.FlowTotal)
		if err != nil {
			middleware.ResponseError(ctx, 4001, err)
			ctx.Abort()
			return
		}
		totalCounter.Increase()

		dayCount, _ := totalCounter.GetDayData(time.Now())
		fmt.Printf("totalCounter qps: %v, dayCount: %v", totalCounter.QPS, dayCount)

		serviceCounter, err := public.FlowCounterHandler.GetCounter(public.FlowCountServicePrefix + serviceDetail.Info.ServiceName)
		if err != nil {
			middleware.ResponseError(ctx, 4001, err)
			ctx.Abort()
			return
		}
		serviceCounter.Increase()

		dayServiceCount, _ := serviceCounter.GetDayData(time.Now())
		fmt.Printf("totalCounter qps: %v, dayCount: %v", serviceCounter.QPS, dayServiceCount)

		ctx.Next()
	}
}