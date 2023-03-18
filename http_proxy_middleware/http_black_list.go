package http_proxy_middleware

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/hugokung/micro_gateway/dao"
	"github.com/hugokung/micro_gateway/middleware"
	"github.com/hugokung/micro_gateway/public"
	"github.com/pkg/errors"
)

func HTTPBlacListMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		serviceInterface, ok := ctx.Get("service")
		if !ok {
			middleware.ResponseError(ctx, 2001, errors.New("service not found"))
			ctx.Abort()
			return
		}
		serviceDetail := serviceInterface.(*dao.ServiceDetail)
		whileList := []string{}
		blackList := []string{}

		if serviceDetail.AccessControl.BlackList != "" {
			blackList = strings.Split(serviceDetail.AccessControl.BlackList, ",")
		}
		if serviceDetail.AccessControl.WhiteList != "" {
			whileList = strings.Split(serviceDetail.AccessControl.WhiteList, ",")
		}
		if serviceDetail.AccessControl.OpenAuth == 1 && len(whileList) == 0 && len(blackList) > 0 {
			if public.InStringSlice(blackList, ctx.ClientIP()) {
				middleware.ResponseError(ctx, 30001, errors.New(fmt.Sprintf("%s in black ip list", ctx.ClientIP())))
				ctx.Abort()
				return
			}
		}
		ctx.Next()
	}
}