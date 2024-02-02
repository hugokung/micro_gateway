package http_proxy_middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/hugokung/micro_gateway/internal/dao"
	"github.com/hugokung/micro_gateway/pkg/response"
	"github.com/pkg/errors"
)

func HTTPHeaderTransferMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		serviceInterface, ok := ctx.Get("service")
		if !ok {
			response.ResponseError(ctx, 2001, errors.New("service not found"))
			ctx.Abort()
			return
		}
		serviceDetail := serviceInterface.(*dao.ServiceDetail)
		for _, item := range strings.Split(serviceDetail.HTTPRule.HeaderTransfor, ",") {
			items := strings.Split(item, " ")
			if len(items) < 3 || len(items) > 3 {
				continue
			}
			if items[0] == "add" || items[0] == "edit" {
				ctx.Request.Header.Set(items[1], items[2])
			}
			if items[0] == "del" {
				ctx.Request.Header.Del(items[1])
			}
		}
		ctx.Next()
	}
}