package http_proxy_middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/hugokung/micro_gateway/internal/dao"
	"github.com/hugokung/micro_gateway/pkg/public"
	"github.com/hugokung/micro_gateway/pkg/response"
	"github.com/pkg/errors"
)

func HTTPStripUriMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		serviceInterface, ok := ctx.Get("service")
		if !ok {
			response.ResponseError(ctx, 2001, errors.New("service not found"))
			ctx.Abort()
			return
		}
		serviceDetail := serviceInterface.(*dao.ServiceDetail)
		if serviceDetail.HTTPRule.RuleType == public.LoadTypeHTTP && serviceDetail.HTTPRule.NeedStripUri == 1 {
			ctx.Request.URL.Path = strings.Replace(ctx.Request.URL.Path, serviceDetail.HTTPRule.Rule, "", 1)
		}
		ctx.Next()
	}
}