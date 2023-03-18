package dto

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hugokung/micro_gateway/public"
)

type AdminLoginInput struct {
	UserName	string	`json:"username" form:"username" comment:"用户名" example:"admin" validate:"required"`
	Passwd		string	`json:"password" form:"password" comment:"密码" example:"123456" validate:"required"`
}

func (param *AdminLoginInput) BindValidParam(c *gin.Context) error {
	return public.DefaultGetValidParams(c, param)
}

type AdminLoginOutput struct {
	Token string `json:"token" form:"token" comment:"token" example:"token" validate:""` //token
}

type AdminSessionInfo struct {
	ID 				int			`json:"id"`
	UserName 		string 		`json:"user_name"`
	LoginTime 		time.Time 	`json:"login_time"`
}