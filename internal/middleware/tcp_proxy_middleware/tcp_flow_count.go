package tcp_proxy_middleware

import (
	"fmt"
	"time"

	"github.com/hugokung/micro_gateway/internal/dao"
	"github.com/hugokung/micro_gateway/pkg/public"
)

func TCPFlowCountMiddleware() func(c *TcpSliceRouterContext) {
	return func(ctx *TcpSliceRouterContext) {
		serviceInterface:= ctx.Get("service")
		if serviceInterface == nil {
			ctx.conn.Write([]byte("service not found"))
			ctx.Abort()
			return
		}
		serviceDetail := serviceInterface.(*dao.ServiceDetail)
		totalCounter, err := public.FlowCounterHandler.GetCounter(public.FlowTotal)
		if err != nil {
			ctx.conn.Write([]byte(err.Error()))
			ctx.Abort()
			return
		}
		totalCounter.Increase()

		dayCount, _ := totalCounter.GetDayData(time.Now())
		fmt.Printf("totalCounter qps: %v, dayCount: %v", totalCounter.QPS, dayCount)

		serviceCounter, err := public.FlowCounterHandler.GetCounter(public.FlowCountServicePrefix + serviceDetail.Info.ServiceName)
		if err != nil {
			ctx.conn.Write([]byte(err.Error()))
			ctx.Abort()
			return
		}
		serviceCounter.Increase()

		dayServiceCount, _ := serviceCounter.GetDayData(time.Now())
		fmt.Printf("totalCounter qps: %v, dayCount: %v", serviceCounter.QPS, dayServiceCount)

		ctx.Next()
	}
}