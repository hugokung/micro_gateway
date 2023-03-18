package http_proxy_middleware

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/hugokung/micro_gateway/dao"
	"github.com/hugokung/micro_gateway/middleware"
	"github.com/pkg/errors"
)

func HTTPUrlRewriteMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		serviceInterface, ok := ctx.Get("service")
		if !ok {
			middleware.ResponseError(ctx, 2001, errors.New("service not found"))
			ctx.Abort()
			return
		}
		serviceDetail := serviceInterface.(*dao.ServiceDetail)

		for _, item := range strings.Split(serviceDetail.HTTPRule.UrlRewrite, ",") {
			items := strings.Split(item, " ")
			if len(items) == 2 {
				reg, err := regexp.Compile(items[0])
				if err != nil {
					fmt.Println("regexp Complie err", err)
					continue
				}
				replacePath := reg.ReplaceAll([]byte(ctx.Request.URL.Path), []byte(items[1]))
				ctx.Request.URL.Path = string(replacePath)
			}
		}
		ctx.Next()
	}
}