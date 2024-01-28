package api

import (
	"encoding/base64"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/e421083458/golang_common/lib"
	"github.com/gin-gonic/gin"
	"github.com/hugokung/micro_gateway/internal/dao"
	"github.com/hugokung/micro_gateway/internal/dto"
	"github.com/hugokung/micro_gateway/pkg/public"
	"github.com/hugokung/micro_gateway/pkg/response"
	"github.com/pkg/errors"
)

type OAuthController struct {}

func OAuthRegister(group *gin.RouterGroup) {
	controller := &OAuthController{}
	group.POST("/tokens", controller.Tokens)
}

// Tokens godoc
// @Summary 获取TOKEN
// @Description 获取TOKEN
// @Tags OAUTH
// @ID /oauth/tokens
// @Accept  json
// @Produce  json
// @Param body body dto.TokensInput true "body"
// @Success 200 {object} middleware.Response{data=dto.TokensOutput} "success"
// @Router /oauth/tokens [post]
func (oauth *OAuthController) Tokens(c *gin.Context) {
	params := &dto.TokensInput{}
	if err := params.BindValidParam(c); err != nil {
		response.ResponseError(c, 20000, err)
		return
	}
	headers := strings.Split(c.GetHeader("Authorization"), " ")
	if len(headers) != 2 {
		response.ResponseError(c, 20001, errors.New("用户名或密码格式错误"))
		return
	}
	appSecret, err := base64.StdEncoding.DecodeString(headers[1])
	if err != nil {
		response.ResponseError(c, 20002, err)
		return
	}
	//  取出 app_id secret
	//  生成 app_list
	//  匹配 app_id
	//  基于 jwt生成token
	//  生成 output

	parts := strings.Split(string(appSecret), ":")
	if len(parts) != 2 {
		response.ResponseError(c, 20003, errors.New("用户名或密码格式错误"))
		return
	}

	appList := dao.AppManagerHandler.GetAppList()
	for _, appInfo := range appList {
		if appInfo.AppID == parts[0] && appInfo.Secret == parts[1]{
			claims := jwt.StandardClaims{
				Issuer: appInfo.AppID,
				ExpiresAt: time.Now().Add(public.JwtExpires * time.Second).In(lib.TimeLocation).Unix(),
			}
			token, err := public.JwtEncode(claims)
			if err != nil {
				response.ResponseError(c, 20004, err)
				return
			}
			output := &dto.TokensOutput{
				ExpiresIn: public.JwtExpires,
				TokenType: "Bearer",
				AccessToken: token,
				Scope: "read_write",
			}
			response.ResponseSuccess(c, output)
			return
		}
	}
	response.ResponseError(c, 20005, errors.New("未匹配到正确的app信息"))
}