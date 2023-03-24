package dto

import (
	"github.com/gin-gonic/gin"
	"github.com/hugokung/micro_gateway/public"
)

type TokensInput struct {
	GrantType string `json:"grant_type" form:"grant_type" comment:"授权类型" example:"client_credentials" validate:"required"` //授权类型
	Scope     string `json:"scope" form:"scope" comment:"权限范围" example:"read_write" validate:"required"`                   //权限范围
}

func (param *TokensInput) BindValidParam(c *gin.Context) error {
	return public.DefaultGetValidParams(c, param)
}

type TokensOutput struct {
	AccessToken string `json:"access_token" form:"access_token"`
	ExpiresIn   int    `json:"expires_in" form:"expires_in"`
	TokenType   string `json:"token_type" form:"token_type"`
	Scope       string `json:"scope" form:"scope"`
}
