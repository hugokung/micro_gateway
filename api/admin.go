package api

import (
	"encoding/json"
	"fmt"

	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/hugokung/micro_gateway/internal/dao"
	"github.com/hugokung/micro_gateway/internal/dto"
	"github.com/hugokung/micro_gateway/pkg/golang_common/lib"
	"github.com/hugokung/micro_gateway/pkg/public"
	"github.com/hugokung/micro_gateway/pkg/response"
)

type AdminController struct {}

func AdminRegister(group *gin.RouterGroup) {
	controller := &AdminController{}
	group.GET("/admin_info", controller.AdminInfo)
	group.POST("/change_pwd", controller.ChangePwd)
}


// AdminInfo godoc
// @Summary 管理员信息
// @Description 管理员信息
// @Tags 管理员接口
// @ID /admin/admin_info
// @Accept  json
// @Produce  json
// @Success 200 {object} middleware.Response{data=dto.AdminInfoOutput} "success"
// @Router /admin/admin_info [get]
func (adminlogin *AdminController) AdminInfo(c *gin.Context) {
	sess := sessions.Default(c)
	sessInfo := sess.Get(public.AdminSessionInfoKey)
	adminSessionInfo := &dto.AdminSessionInfo{}
	if err := json.Unmarshal([]byte(fmt.Sprint(sessInfo)), adminSessionInfo); err != nil {
		response.ResponseError(c, 20003, err)
		return
	}
	out := &dto.AdminInfoOutput{
		ID: adminSessionInfo.ID,
		Name: adminSessionInfo.UserName,
		LoginTime: adminSessionInfo.LoginTime,
		Avatar: "https://wpimg.wallstcn.com/f778738c-e4f8-4870-b634-56703b4acafe.gif",
		Introduction: "i am a super administrator",
		Roles: []string{"admin"},
	}
	response.ResponseSuccess(c, out)
}


// ChangePwd godoc
// @Summary 密码修改
// @Description 密码修改
// @Tags 管理员接口
// @ID /admin/change_pwd
// @Accept  json
// @Produce  json
// @Param body body dto.ChangePwdInput true "body"
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /admin/change_pwd [post]
func (admin *AdminController) ChangePwd(c *gin.Context) {
	params := &dto.ChangePwdInput{}
	if err := params.BindValidParam(c); err != nil {
		response.ResponseError(c, 20001, err)
		return
	}

	sess := sessions.Default(c)
	sessInfo := sess.Get(public.AdminSessionInfoKey)
	adminSessionInfo := &dto.AdminSessionInfo{}
	if err := json.Unmarshal([]byte(fmt.Sprint(sessInfo)), adminSessionInfo); err != nil {
		response.ResponseError(c, 20002, err)
		return
	}

	tx, err := lib.GetGormPool("default")
	if err != nil {
		response.ResponseError(c, 20003, err)
		return
	}
	adminUser := &dao.Admin{}
	adminUser, err = adminUser.Find(c, tx, &dao.Admin{
		UserName: adminSessionInfo.UserName,
	})
	if err != nil {
		response.ResponseError(c, 20004, err)
		return
	}
	saltPassword := public.GenSaltPassword(adminUser.Salt, params.Password)
	adminUser.Password = saltPassword
	err = adminUser.Save(c, tx)
	if err != nil {
		response.ResponseError(c, 20005, err)
		return
	}

	response.ResponseSuccess(c, "")
}