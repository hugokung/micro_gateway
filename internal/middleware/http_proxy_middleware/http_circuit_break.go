package http_proxy_middleware

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/afex/hystrix-go/hystrix"
	"github.com/gin-gonic/gin"
	"github.com/hugokung/micro_gateway/internal/dao"
	"github.com/hugokung/micro_gateway/pkg/response"
)

func HTTPCircuitBreakMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		serviceInterface, ok := ctx.Get("service")
		if !ok {
			response.ResponseError(ctx, 2001, errors.New("service not found"))
			ctx.Abort()
			return
		}
		serviceDetail := serviceInterface.(*dao.ServiceDetail)

		if serviceDetail.CircuitConfig.NeedCircuit == 1 {
			hystrix.Do(serviceDetail.Info.ServiceName, func () error {
				ctx.Next()
				code := ctx.Writer.Status()
				if code != http.StatusOK {
					return fmt.Errorf("status code %d", code)
				}
				return nil
			}, func (err error) error {
				if err != nil {
					//TODO: report error event
					response.ResponseError(ctx, 50000, errors.New(serviceDetail.CircuitConfig.FallBackMsg))
					ctx.Abort()
				}
				return nil
			})
			return
		}
		ctx.Next()
	}
}