package controller

import (
	"time"

	"github.com/e421083458/golang_common/lib"
	"github.com/gin-gonic/gin"
	"github.com/hugokung/micro_gateway/dao"
	"github.com/hugokung/micro_gateway/dto"
	"github.com/hugokung/micro_gateway/middleware"
	"github.com/hugokung/micro_gateway/public"
	"github.com/pkg/errors"
)

type DashBoardController struct {}

func DashBoardRegister(group *gin.RouterGroup) {
	controller := &DashBoardController{}
	group.GET("/panel_group_data", controller.PanelGroupData)
	group.GET("/service_stat", controller.ServiceStat)
	group.GET("/flow_stat", controller.FlowStat)
}

// PanelGroupData godoc
// @Summary 指标统计
// @Description 指标统计
// @Tags 首页大盘
// @ID /dashboard/panel_group_data
// @Accept  json
// @Produce  json
// @Success 200 {object} middleware.Response{data=dto.PanelGroupDataOutput} "success"
// @Router /dashboard/panel_group_data [get]
func (d *DashBoardController) PanelGroupData(c *gin.Context) {
	serviceInfo := &dao.ServiceInfo{}
	_, serviceNum, err := serviceInfo.PageList(c, lib.GORMDefaultPool, &dto.ServiceInfoInput{PageSize: 1, PageNo: 1})
	if err != nil {
		middleware.ResponseError(c, 2000, err)
		return
	}

	appInfo := &dao.App{}
	_, appNum, err := appInfo.PageList(c, lib.GORMDefaultPool, &dto.AppInfoListInput{PageNo: 1, PageSize: 1})
	if err != nil {
		middleware.ResponseError(c, 20001, err)
		return
	}

	counter, err := public.FlowCounterHandler.GetCounter(public.FlowTotal)
	if err != nil {
		middleware.ResponseError(c, 20002, err)
		return
	}

	out := &dto.PanelGroupDataOutput{
		ServiceNum: serviceNum,
		AppNum: appNum,
		CurrentQPS: counter.QPS,
		TodayRequestNum: counter.TotalCount,
	}
	middleware.ResponseSuccess(c, out)
}

// ServiceStat godoc
// @Summary 服务统计
// @Description 服务统计
// @Tags 首页大盘
// @ID /dashboard/service_stat
// @Accept  json
// @Produce  json
// @Success 200 {object} middleware.Response{data=dto.DashServiceStatOutput} "success"
// @Router /dashboard/service_stat [get]
func (d *DashBoardController) ServiceStat(c *gin.Context) {
	serviceInfo := &dao.ServiceInfo{}
	itemList, err := serviceInfo.GroupByLoadType(c, lib.GORMDefaultPool)
	if err != nil {
		middleware.ResponseError(c, 20000, err)
		return
	}
	out := &dto.DashServiceStatOutput{
		Legend: []string{},
		Data: itemList,
	}
	for idx, item := range itemList {
		name, ok := public.LoadTypeMap[item.LoadType]
		if !ok {
			middleware.ResponseError(c, 20001, errors.New("load type not found"))
			return
		}
		out.Legend = append(out.Legend, name)
		out.Data[idx].Name = name
	}
	middleware.ResponseSuccess(c, out)
}

// FlowStat godoc
// @Summary 服务统计
// @Description 服务统计
// @Tags 首页大盘
// @ID /dashboard/flow_stat
// @Accept  json
// @Produce  json
// @Success 200 {object} middleware.Response{data=dto.ServiceStatOutput} "success"
// @Router /dashboard/flow_stat [get]
func (service *DashBoardController) FlowStat(c *gin.Context) {

	counter, err := public.FlowCounterHandler.GetCounter(public.FlowTotal)

	if err != nil {
		middleware.ResponseError(c, 20000, err)
		return
	}
	todayList := []int64{}
	currentTime := time.Now()
	for i := 0; i <= currentTime.Hour(); i++ {
		dateTime := time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), i, 0, 0, 0, lib.TimeLocation)
		hourData, _ := counter.GetHourData(dateTime)
		todayList = append(todayList, hourData)
	}

	yesterdayList := []int64{}
	yesterdayTime:= currentTime.Add(-1 * time.Duration(time.Hour*24))
	for i := 0; i <= 23; i++ {
		dateTime := time.Date(yesterdayTime.Year(), yesterdayTime.Month(), yesterdayTime.Day(), i, 0, 0, 0, lib.TimeLocation)
		hourData, _ := counter.GetHourData(dateTime)
		yesterdayList = append(yesterdayList, hourData)
	}
	middleware.ResponseSuccess(c, &dto.ServiceStatOutput{
		Today:     todayList,
		Yesterday: yesterdayList,
	})
}
