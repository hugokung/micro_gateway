package middleware

import (
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/hugokung/micro_gateway/pkg/golang_common/lib"
	"github.com/hugokung/micro_gateway/pkg/response"
)

func IPAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		isMatched := false
		for _, host := range lib.GetStringSliceConf("base.http.allow_ip") {
			if c.ClientIP() == host {
				isMatched = true
			}
		}
		if len(lib.GetStringSliceConf("base.http.allow_ip")) == 0 {
			isMatched = true
		}
		if !isMatched{
			response.ResponseError(c, response.InternalErrorCode, errors.New(fmt.Sprintf("%v, not in iplist", c.ClientIP())))
			c.Abort()
			return
		}
		c.Next()
	}
}
