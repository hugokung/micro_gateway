package http_proxy_middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/hugokung/micro_gateway/dao"
	"github.com/hugokung/micro_gateway/middleware"
)

func HTTPAccessModeMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		service, err := dao.ServiceManagerHandler.HttpAccessMode(ctx)
		if err != nil {
			middleware.ResponseError(ctx, 1001, err)
			ctx.Abort()
			return
		}
		ctx.Set("service", service)
		ctx.Next()
	}
}