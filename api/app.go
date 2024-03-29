package api

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hugokung/micro_gateway/internal/dao"
	"github.com/hugokung/micro_gateway/internal/dto"
	"github.com/hugokung/micro_gateway/pkg/golang_common/lib"
	"github.com/hugokung/micro_gateway/pkg/public"
	"github.com/hugokung/micro_gateway/pkg/response"
	"github.com/pkg/errors"
)

type AppController struct {}

func AppRegister(group *gin.RouterGroup) {
	controller := &AppController{}
	group.POST("/app_info_add", controller.AppInfoAdd)
	group.GET("/app_delete", controller.AppDelete)
	group.POST("/app_update", controller.AppUpdate)
	group.GET("/app_detail", controller.AppDetail)
	group.GET("/app_list", controller.AppList)
	group.GET("/app_stat", controller.AppStatistics)
}

// AppInfoAdd godoc
// @Summary 租户添加
// @Description 租户添加
// @Tags 租户管理
// @ID /app/app_info_add
// @Accept  json
// @Produce  json
// @Param body body dto.AppAddInput true "body"
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /app/app_info_add [post]
func (a *AppController) AppInfoAdd(c *gin.Context) {
	params := &dto.AppAddInput{}
	if err := params.GetValidParams(c); err != nil {
		response.ResponseError(c, 20000, err)
		return
	}

	search := &dao.App{
		AppID: params.AppID,
		IsDelete: 0,
	}
	if _, err := search.Find(c, lib.GORMDefaultPool, search); err == nil {
		response.ResponseError(c, 20001, errors.New("app_id已存在，请重新输入"))
		return
	}

	if params.Secret == "" {
		params.Secret = public.MD5(params.AppID)
	}
	appModel := &dao.App{
		AppID: params.AppID,
		Name: params.Name,
		WhiteIPS: params.WhiteIPS,
		Secret: params.Secret,
		Qpd: params.Qpd,
		Qps: params.Qps,
	}
	if err := appModel.Save(c, lib.GORMDefaultPool); err != nil {
		response.ResponseError(c, 20002, err)
		return
	}
	response.ResponseSuccess(c, "")
	return
}

// AppDelete godoc
// @Summary 租户删除
// @Description 租户删除
// @Tags 租户管理
// @ID /app/app_delete
// @Accept  json
// @Produce  json
// @Param id query string true "租户ID"
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /app/app_delete [get]
func (a *AppController) AppDelete(c *gin.Context) {
	params := &dto.AppDeleteInput{}
	if err := params.GetValidParams(c); err != nil {
		response.ResponseError(c, 2000, err)
		return
	}
	search := &dao.App{
		ID: params.ID,
		IsDelete: 0,
	}
	res, err := search.Find(c, lib.GORMDefaultPool, search)
	if err != nil {
		response.ResponseError(c, 20001, err)
		return
	}
	if res == nil {
		response.ResponseError(c, 20002, errors.New("无此租户"))
		return
	}
	res.IsDelete = 1
	err = res.Save(c, lib.GORMDefaultPool)
	if err != nil {
		response.ResponseError(c, 20003, err)
		return
	}
	response.ResponseSuccess(c, "")
}

// AppUpdate godoc
// @Summary 租户更新
// @Description 租户更新
// @Tags 租户管理
// @ID /app/app_update
// @Accept  json
// @Produce  json
// @Param body body dto.AppUpdateInput true "body"
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /app/app_update [post]
func (a *AppController) AppUpdate(c *gin.Context) {
	params := &dto.AppUpdateInput{}
	if err := params.GetValidParams(c); err != nil {
		response.ResponseError(c, 20000, err)
		return
	}

	search := &dao.App{
		AppID: params.AppID,
		IsDelete: 0,
	}
	info, err := search.Find(c, lib.GORMDefaultPool, search)
	if err != nil {
		response.ResponseError(c, 20001, err)
		return
	}

	if params.Secret == "" {
		params.Secret = public.MD5(params.AppID)
	}

	info.Name = params.Name
	info.Secret = params.Secret
	info.WhiteIPS = params.WhiteIPS
	info.Qps = params.Qps
	info.Qpd = params.Qpd

	if err := info.Save(c, lib.GORMDefaultPool); err != nil {
		response.ResponseError(c, 20002, err)
		return
	}

	response.ResponseSuccess(c, "")
}

// AppDetail godoc
// @Summary 租户详情
// @Description 租户详情
// @Tags 租户管理
// @ID /app/app_detail
// @Accept  json
// @Produce  json
// @Param id query string true "租户ID"
// @Success 200 {object} middleware.Response{data=dao.App} "success"
// @Router /app/app_detail [get]
func (a *AppController) AppDetail(c *gin.Context) {
	params := &dto.AppDetailInput{}
	if err := params.GetValidParams(c); err != nil {
		response.ResponseError(c, 2000, err)
		return
	}
	search := &dao.App{
		ID: params.ID,
		IsDelete: 0,
	}
	info, err := search.Find(c, lib.GORMDefaultPool, search)
	if err != nil {
		response.ResponseError(c, 20001, err)
		return
	}
	response.ResponseSuccess(c, info)
}

// AppList godoc
// @Summary 租户列表
// @Description 租户列表
// @Tags 租户管理
// @ID /app/app_list
// @Accept  json
// @Produce  json
// @Param info query string false "关键词"
// @Param page_size query int true "每页多少条"
// @Param page_no query int true "页码"
// @Success 200 {object} middleware.Response{data=dto.AppListOutput} "success"
// @Router /app/app_list [get]
func (a *AppController) AppList(c *gin.Context) {
	params := &dto.AppInfoListInput{}
	if err := params.GetValidParams(c); err != nil {
		response.ResponseError(c, 20000, err)
		return
	}
	fmt.Println(params)
	query := &dao.App{}
	appList, total, err := query.PageList(c, lib.GORMDefaultPool, params)
	if err != nil {
		response.ResponseError(c, 20001, err)
		return
	}
	out := &dto.AppListOutput{
		Total: total,
	}
	out.List = make([]dto.AppListItemOutput, 0)
	for _, app := range appList {
		appCounter, err := public.FlowCounterHandler.GetCounter(public.FlowCountAppPrefix + app.AppID)
		if err != nil {
			response.ResponseError(c, 20002, err)
			return
		}
		item := dto.AppListItemOutput{
			ID: app.ID,
			AppID: app.AppID,
			Secret: app.Secret,
			Name: app.Name,
			WhiteIPS: app.WhiteIPS,
			Qpd: app.Qpd,
			Qps: app.Qps,
			RealQpd: appCounter.TotalCount,
			RealQps: appCounter.QPS,
			UpdatedAt: app.UpdatedAt,
			CreatedAt: app.CreatedAt,
			IsDelete: app.IsDelete,
		}
		out.List = append(out.List, item)
	}
	response.ResponseSuccess(c, out)
}

// AppStatistics godoc
// @Summary 租户统计
// @Description 租户统计
// @Tags 租户管理
// @ID /app/app_stat
// @Accept  json
// @Produce  json
// @Param id query string true "租户ID"
// @Success 200 {object} middleware.Response{data=dto.StatisticsOutput} "success"
// @Router /app/app_stat [get]
func (admin *AppController) AppStatistics(c *gin.Context) {
	params := &dto.AppDetailInput{}
	if err := params.GetValidParams(c); err != nil {
		response.ResponseError(c, 2001, err)
		return
	}

	tx, err := lib.GetGormPool("default")
	if err != nil {
		response.ResponseError(c, 2001, err)
		return
	}
	appInfo := &dao.App{
		ID: params.ID,
		IsDelete: 0,
	}
	appInfo, err = appInfo.Find(c, tx, appInfo)
	if err != nil {
		response.ResponseError(c, 20002, err)
		return
	}
	appCounter, err := public.FlowCounterHandler.GetCounter(public.FlowCountAppPrefix + appInfo.AppID)
	if err != nil {
		response.ResponseError(c, 20003, err)
		return
	}

	currentTime := time.Now()
	todayStat := []int64{}
	for i := 0; i <= currentTime.In(lib.TimeLocation).Hour(); i++ {
		dateTime := time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), i, 0, 0, 0, lib.TimeLocation)
		hourData, _ := appCounter.GetHourData(dateTime)
		todayStat = append(todayStat, hourData)
	}

	//昨日流量全天小时级访问统计
	yesterdayStat := []int64{}
	yesterdayTime := time.Now().Add(-1 * time.Duration(time.Hour * 24))
	for i := 0; i <= 23; i++ {
		dateTime := time.Date(yesterdayTime.Year(), yesterdayTime.Month(), yesterdayTime.Day(), i, 0, 0, 0, lib.TimeLocation)
		hourData, _ := appCounter.GetHourData(dateTime)
		yesterdayStat = append(yesterdayStat, hourData)
	}
	stat := dto.StatisticsOutput{
		Today:     todayStat,
		Yesterday: yesterdayStat,
	}
	response.ResponseSuccess(c, stat)
	return
}