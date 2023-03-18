package controller

import (
	"encoding/json"
	"time"

	"github.com/e421083458/golang_common/lib"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/hugokung/micro_gateway/dao"
	"github.com/hugokung/micro_gateway/dto"
	"github.com/hugokung/micro_gateway/middleware"
	"github.com/hugokung/micro_gateway/public"
)

type AdminLoginController struct {}

func AdminLoginRegister(group *gin.RouterGroup) {
	controller := &AdminLoginController{}
	group.POST("/login", controller.AdminLogin)
	group.GET("logout", controller.AdminLoginOut)
}


// AdminLogin godoc
// @Summary 管理员登陆
// @Description 管理员登陆
// @Tags 管理员接口
// @ID /admin_login/login
// @Accept  json
// @Produce  json
// @Param body body dto.AdminLoginInput true "body"
// @Success 200 {object} middleware.Response{data=dto.AdminLoginOutput} "success"
// @Router /admin_login/login [post]
func (adminlogin *AdminLoginController) AdminLogin(c *gin.Context) {
	params := &dto.AdminLoginInput{}
	if err := params.BindValidParam(c); err != nil {
		middleware.ResponseError(c, 1001, err)
		return
	}

	admin := &dao.Admin{}
	tx, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(c, 2000, err)
		return
	}
	admin, err = admin.LoginCheck(c, tx, params)
	if err != nil {
		middleware.ResponseError(c, 2001, err)
		return
	}

	sessInfo := &dto.AdminSessionInfo{
		ID: 		admin.Id,
		UserName: 	admin.UserName,
		LoginTime:  time.Now(),
	}
	sessBts, err1 := json.Marshal(sessInfo)
	if err1 != nil {
		middleware.ResponseError(c, 2002, err1)
		return
	}
	sess := sessions.Default(c)
	//为什么key是常量???
	sess.Set(public.AdminSessionInfoKey, string(sessBts))
	sess.Save()

	out := &dto.AdminLoginOutput{Token: admin.UserName}
	middleware.ResponseSuccess(c, out)
}

// AdminLoginOut godoc
// @Summary 管理员退出
// @Description 管理员退出
// @Tags 管理员接口
// @ID /admin_login/logout
// @Accept  json
// @Produce  json
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /admin_login/logout [get]
func (adminlogin *AdminLoginController) AdminLoginOut(c *gin.Context) {

	sess := sessions.Default(c)
	sess.Delete(public.AdminSessionInfoKey)
	sess.Save()
	middleware.ResponseSuccess(c, "")
}