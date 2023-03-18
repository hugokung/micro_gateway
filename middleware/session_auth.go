package middleware

import (
	"errors"

	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/hugokung/micro_gateway/public"
)

func SessionAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		if adminInfo,ok:=session.Get(public.AdminSessionInfoKey).(string);!ok||adminInfo==""{
			ResponseError(c, InternalErrorCode, errors.New("user not login"))
			c.Abort()
			return
		}
		c.Next()
	}
}
