package api

import (
	"fmt"
	"strings"
	"time"

	"github.com/e421083458/golang_common/lib"
	"github.com/gin-gonic/gin"
	"github.com/hugokung/micro_gateway/internal/dao"
	"github.com/hugokung/micro_gateway/internal/dto"
	"github.com/hugokung/micro_gateway/pkg/public"
	"github.com/hugokung/micro_gateway/pkg/response"
	"github.com/pkg/errors"
)

type ServiceController struct {}

func ServiceRegister(group *gin.RouterGroup) {
	controller := &ServiceController{}
	group.GET("/service_info", controller.ServiceInfoList)
	group.GET("/service_delete", controller.ServiceDelete)
	group.GET("/service_detail", controller.ServiceDetail)
	group.GET("/service_stat", controller.ServiceStat)
	group.POST("/service_add_http", controller.ServiceAddHTTP)
	group.POST("/service_update_http", controller.ServiceUpdateHTTP)
	group.POST("/service_add_tcp", controller.ServiceAddTCP)
	group.POST("/service_update_tcp", controller.ServiceUpdateTCP)
	group.POST("/service_add_grpc", controller.ServiceAddGRPC)
	group.POST("/service_update_grpc", controller.ServiceUpdateGRPC)
}


// ServiceInfoList godoc
// @Summary 服务信息列表
// @Description 服务信息列表
// @Tags 服务管理
// @ID /service/service_info
// @Accept  json
// @Produce  json
// @Param info query string false "关键词"
// @Param page_no query int true "页数"
// @Param page_size query int true "每页个数"
// @Success 200 {object} response.Response{data=dto.ServiceInfoOutput} "success"
// @Router /service/service_info [get]
func (service *ServiceController) ServiceInfoList(c *gin.Context) {
	params := &dto.ServiceInfoInput{}

	if err := params.BindValidParam(c); err != nil {
		response.ResponseError(c, 20000, err)
		return
	}

	tx, err := lib.GetGormPool("default")
	if err != nil {
		response.ResponseError(c, 20001, err)
		return
	}

	out := &dto.ServiceInfoOutput{}
	serviceInfo := &dao.ServiceInfo{}
	infoList, total, err1 := serviceInfo.PageList(c, tx, params)
	if err1 != nil {
		response.ResponseError(c, 20002, err)
		return
	}
	out.Total = total
	itemList := make([]*dto.ServiceInfoItem, 0)
	for _, info := range infoList {

		serviceAddr := "unkown"
		clusterIP := lib.GetStringConf("base.cluster.cluster_ip")
		clusterPort := lib.GetStringConf("base.cluster.cluster_port")
		clusterSSLPort := lib.GetStringConf("base.cluster.cluster_ssl_port")

		serviceDetail, err2 := info.ServiceDetail(c, tx, &info)
		if err2 != nil {
			response.ResponseError(c, 20003, err2)
			return
		}
		
		if info.LoadType == public.LoadTypeHTTP && 
		serviceDetail.HTTPRule.RuleType == public.RuleTypePrefixURL && 
		serviceDetail.HTTPRule.NeedHttps == 1 {
			serviceAddr = fmt.Sprintf("%s:%s%s", clusterIP, clusterSSLPort, serviceDetail.HTTPRule.Rule)
		}

		if info.LoadType == public.LoadTypeHTTP && 
		serviceDetail.HTTPRule.RuleType == public.RuleTypePrefixURL && 
		serviceDetail.HTTPRule.NeedHttps == 0 {
			serviceAddr = fmt.Sprintf("%s:%s%s", clusterIP, clusterPort, serviceDetail.HTTPRule.Rule)
		}

		if serviceDetail.Info.LoadType == public.LoadTypeHTTP &&
			serviceDetail.HTTPRule.RuleType == public.RuleTypeDomin {
			serviceAddr = serviceDetail.HTTPRule.Rule
		}
		
		if serviceDetail.Info.LoadType == public.LoadTypeTCP {
			serviceAddr = fmt.Sprintf("%s:%d", clusterIP, serviceDetail.TCPRule.Port)
		}
		
		if serviceDetail.Info.LoadType == public.LoadTypeGRPC {
			serviceAddr = fmt.Sprintf("%s:%d", clusterIP, serviceDetail.GRPCRule.Port)
		}

		counter, err := public.FlowCounterHandler.GetCounter(public.FlowCountServicePrefix + info.ServiceName)
		if err != nil {
			response.ResponseError(c, 20004, err)
			return
		}
		
		item := &dto.ServiceInfoItem{
			ID: info.ID,
			LoadType: info.LoadType,
			ServiceName: info.ServiceName,
			ServiceDesc: info.ServiceDesc,
			ServiceAddr: serviceAddr,
			Qps: counter.QPS,
			Qpd: counter.TotalCount,
			TotalNode: len(serviceDetail.LoadBalance.GetIPListByModel()),
		}
		itemList = append(itemList, item)
	}
	out.InfoList = itemList
	response.ResponseSuccess(c, out)
	return
}

// ServiceDelete godoc
// @Summary 服务删除
// @Description 服务删除
// @Tags 服务管理
// @ID /service/service_delete
// @Accept  json
// @Produce  json
// @Param id query string true "服务ID"
// @Success 200 {object} response.Response{data=string} "success"
// @Router /service/service_delete [get]
func (s *ServiceController) ServiceDelete(c *gin.Context) {
	params := &dto.ServiceDeleteInput{}
	if err := params.BindValidParam(c); err != nil {
		response.ResponseError(c, 2000, err)
		return
	}
	tx, err := lib.GetGormPool("default")
	if err != nil {
		response.ResponseError(c, 20001, err)
		return
	}
	serviceInfo := &dao.ServiceInfo{ID: params.ID}
	serviceInfo, err = serviceInfo.Find(c, tx, serviceInfo)
	if err != nil {
		response.ResponseError(c, 20002, err)
		return
	}
	serviceInfo.IsDelete = 1
	err = serviceInfo.Save(c, tx)
	if err != nil {
		response.ResponseError(c, 20003, err)
		return
	}
	response.ResponseSuccess(c, "")
	return
}

// ServiceAddHTTP godoc
// @Summary 添加HTTP服务
// @Description 添加HTTP服务
// @Tags 服务管理
// @ID /service/service_add_http
// @Accept  json
// @Produce  json
// @Param body body dto.ServiceAddHTTPInput true "body"
// @Success 200 {object} response.Response{data=string} "success"
// @Router /service/service_add_http [post]
func (s *ServiceController) ServiceAddHTTP(c *gin.Context) {
	params := &dto.ServiceAddHTTPInput{}
	if err := params.BindValidParam(c); err != nil {
		response.ResponseError(c, 20000, err)
		return
	}

	if len(strings.Split(params.IpList, ",")) != len(strings.Split(params.WeightList, ",")) {
		response.ResponseError(c, 20004, errors.New("IP列表与权重列表数量不一致"))
		return
	}

	tx, err := lib.GetGormPool("default")
	if err != nil {
		response.ResponseError(c, 20001, err)
		return
	}
	tx = tx.Begin()
	serviceInfo := &dao.ServiceInfo{
		ServiceName: params.ServiceName,
	}
	tmp, err := serviceInfo.Find(c, tx, serviceInfo)
	if err == nil {
		fmt.Println(tmp)
		tx.Rollback()
		response.ResponseError(c, 20002, errors.New("服务已存在??"))
		return
	}

	httpUrl := &dao.HttpRule{RuleType: params.RuleType, Rule: params.Rule}
	if _, err := httpUrl.Find(c, tx, httpUrl); err == nil {
		tx.Rollback()
		response.ResponseError(c, 20003, errors.New("服务接入前缀或域名已存在"))
		return
	}

	

	serviceModel := &dao.ServiceInfo{
		ServiceName: params.ServiceName,
		ServiceDesc: params.ServiceDesc,
	}
	if err := serviceModel.Save(c, tx); err != nil {
		tx.Rollback()
		response.ResponseError(c, 20005, err)
		return
	}
	//serviceModel.ID
	httpRule := &dao.HttpRule{
		ServiceID:      serviceModel.ID,
		RuleType:       params.RuleType,
		Rule:           params.Rule,
		NeedHttps:      params.NeedHttps,
		NeedStripUri:   params.NeedStripUri,
		NeedWebsocket:  params.NeedWebsocket,
		UrlRewrite:     params.UrlRewrite,
		HeaderTransfor: params.HeaderTransfor,
	}
	if err := httpRule.Save(c, tx); err != nil {
		tx.Rollback()
		response.ResponseError(c, 20006, err)
		return
	}

	accessControl := &dao.AccessControl{
		ServiceID:         serviceModel.ID,
		OpenAuth:          params.OpenAuth,
		BlackList:         params.BlackList,
		WhiteList:         params.WhiteList,
		ClientIPFlowLimit: params.ClientipFlowLimit,
		ServiceFlowLimit:  params.ServiceFlowLimit,
	}
	if err := accessControl.Save(c, tx); err != nil {
		tx.Rollback()
		response.ResponseError(c, 20007, err)
		return
	}

	loadbalance := &dao.LoadBalance{
		ServiceID:              serviceModel.ID,
		RoundType:              params.RoundType,
		IpList:                 params.IpList,
		WeightList:             params.WeightList,
		UpstreamConnectTimeout: params.UpstreamConnectTimeout,
		UpstreamHeaderTimeout:  params.UpstreamHeaderTimeout,
		UpstreamIdleTimeout:    params.UpstreamIdleTimeout,
		UpstreamMaxIdle:        params.UpstreamMaxIdle,
	}
	if err := loadbalance.Save(c, tx); err != nil {
		tx.Rollback()
		response.ResponseError(c, 20008, err)
		return
	}
	tx.Commit()
	response.ResponseSuccess(c, "")
}

// ServiceUpdateHTTP godoc
// @Summary 修改HTTP服务
// @Description 修改HTTP服务
// @Tags 服务管理
// @ID /service/service_update_http
// @Accept  json
// @Produce  json
// @Param body body dto.ServiceUpdateHTTPInput true "body"
// @Success 200 {object} response.Response{data=string} "success"
// @Router /service/service_update_http [post]
func (service *ServiceController) ServiceUpdateHTTP(c *gin.Context) {
	params := &dto.ServiceUpdateHTTPInput{}
	if err := params.BindValidParam(c); err != nil {
		response.ResponseError(c, 2000, err)
		return
	}

	if len(strings.Split(params.IpList, ",")) != len(strings.Split(params.WeightList, ",")) {
		response.ResponseError(c, 20001, errors.New("IP列表与权重列表数量不一致"))
		return
	}

	tx, err := lib.GetGormPool("default")
	if err != nil {
		response.ResponseError(c, 20002, err)
		return
	}
	tx = tx.Begin()
	serviceInfo := &dao.ServiceInfo{ServiceName: params.ServiceName}
	serviceInfo, err = serviceInfo.Find(c, tx, serviceInfo)
	if err != nil {
		tx.Rollback()
		response.ResponseError(c, 20003, errors.New("服务不存在"))
		return
	}
	serviceDetail, err := serviceInfo.ServiceDetail(c, tx, serviceInfo)
	if err != nil {
		tx.Rollback()
		response.ResponseError(c, 20004, errors.New("服务不存在"))
		return
	}

	info := serviceDetail.Info
	info.ServiceDesc = params.ServiceDesc
	if err := info.Save(c, tx); err != nil {
		tx.Rollback()
		response.ResponseError(c, 20005, err)
		return
	}

	httpRule := serviceDetail.HTTPRule
	httpRule.NeedHttps = params.NeedHttps
	httpRule.NeedStripUri = params.NeedStripUri
	httpRule.NeedWebsocket = params.NeedWebsocket
	httpRule.UrlRewrite = params.UrlRewrite
	httpRule.HeaderTransfor = params.HeaderTransfor
	if err := httpRule.Save(c, tx); err != nil {
		tx.Rollback()
		response.ResponseError(c, 20006, err)
		return
	}

	accessControl := serviceDetail.AccessControl
	accessControl.OpenAuth = params.OpenAuth
	accessControl.BlackList = params.BlackList
	accessControl.WhiteList = params.WhiteList
	accessControl.ClientIPFlowLimit = params.ClientipFlowLimit
	accessControl.ServiceFlowLimit = params.ServiceFlowLimit
	if err := accessControl.Save(c, tx); err != nil {
		tx.Rollback()
		response.ResponseError(c, 20007, err)
		return
	}

	loadbalance := serviceDetail.LoadBalance
	loadbalance.RoundType = params.RoundType
	loadbalance.IpList = params.IpList
	loadbalance.WeightList = params.WeightList
	loadbalance.UpstreamConnectTimeout = params.UpstreamConnectTimeout
	loadbalance.UpstreamHeaderTimeout = params.UpstreamHeaderTimeout
	loadbalance.UpstreamIdleTimeout = params.UpstreamIdleTimeout
	loadbalance.UpstreamMaxIdle = params.UpstreamMaxIdle
	if err := loadbalance.Save(c, tx); err != nil {
		tx.Rollback()
		response.ResponseError(c, 20008, err)
		return
	}
	tx.Commit()
	response.ResponseSuccess(c, "")
}

// ServiceDetail godoc
// @Summary 服务详情
// @Description 服务详情
// @Tags 服务管理
// @ID /service/service_detail
// @Accept  json
// @Produce  json
// @Param id query string true "服务ID"
// @Success 200 {object} response.Response{data=dao.ServiceDetail} "success"
// @Router /service/service_detail [get]
func (service *ServiceController) ServiceDetail(c *gin.Context) {
	params := &dto.ServiceDeleteInput{}
	if err := params.BindValidParam(c); err != nil {
		response.ResponseError(c, 2000, err)
		return
	}

	tx, err := lib.GetGormPool("default")
	if err != nil {
		response.ResponseError(c, 2001, err)
		return
	}

	//读取基本信息
	serviceInfo := &dao.ServiceInfo{ID: params.ID}
	serviceInfo, err = serviceInfo.Find(c, tx, serviceInfo)
	if err != nil {
		response.ResponseError(c, 2002, err)
		return
	}
	serviceDetail, err := serviceInfo.ServiceDetail(c, tx, serviceInfo)
	if err != nil {
		response.ResponseError(c, 2003, err)
		return
	}
	response.ResponseSuccess(c, serviceDetail)
}

// ServiceAddTCP godoc
// @Summary 添加TCP服务
// @Description 添加TCP服务
// @Tags 服务管理
// @ID /service/service_add_tcp
// @Accept  json
// @Produce  json
// @Param body body dto.ServiceAddTCPInput true "body"
// @Success 200 {object} response.Response{data=string} "success"
// @Router /service/service_add_tcp [post]
func (s *ServiceController) ServiceAddTCP(c *gin.Context) {
	params := &dto.ServiceAddTCPInput{}
	if err := params.BindValidParam(c); err != nil {
		response.ResponseError(c, 20000, err)
		return
	}

	if len(strings.Split(params.IpList, ",")) != len(strings.Split(params.WeightList, ",")) {
			response.ResponseError(c, 20004, errors.New("IP列表与权重列表数量不一致"))
			return
	}

	tx, err := lib.GetGormPool("default")
	if err != nil {
		response.ResponseError(c, 20001, err)
		return
	}

	serviceInfo := &dao.ServiceInfo{
		ServiceName: params.ServiceName,
		IsDelete: 0,
	}
	if _, err = serviceInfo.Find(c, tx, serviceInfo); err == nil {
		response.ResponseError(c, 20002, errors.New("服务名已存在"))
		return
	}

	TcpUrl := &dao.TcpRule{Port: params.Port}
	if _, err := TcpUrl.Find(c, tx, TcpUrl); err == nil {
		response.ResponseError(c, 20003, errors.New("端口被占用"))
		return
	}

	tx = tx.Begin()
	serviceModel := &dao.ServiceInfo{
		LoadType:    public.LoadTypeTCP,
		ServiceName: params.ServiceName,
		ServiceDesc: params.ServiceDesc,
	}
	if err := serviceModel.Save(c, tx); err != nil {
		tx.Rollback()
		response.ResponseError(c, 20005, err)
		return
	}
	//serviceModel.ID
	tcpRule := &dao.TcpRule{
		ServiceID:      serviceModel.ID,
		Port: params.Port,	
	}
	if err := tcpRule.Save(c, tx); err != nil {
		tx.Rollback()
		response.ResponseError(c, 20006, err)
		return
	}

	accessControl := &dao.AccessControl{
		ServiceID:         serviceModel.ID,
		OpenAuth:          params.OpenAuth,
		BlackList:         params.BlackList,
		WhiteList:         params.WhiteList,
		WhiteHostName: 		params.WhiteHostName,
		ClientIPFlowLimit: params.ClientIPFlowLimit,
		ServiceFlowLimit:  params.ServiceFlowLimit,
	}
	if err := accessControl.Save(c, tx); err != nil {
		tx.Rollback()
		response.ResponseError(c, 20007, err)
		return
	}

	loadbalance := &dao.LoadBalance{
		ServiceID:              serviceModel.ID,
		RoundType:              params.RoundType,
		IpList:                 params.IpList,
		WeightList:             params.WeightList,
		ForbidList: 			params.ForbidList,
	}
	if err := loadbalance.Save(c, tx); err != nil {
		tx.Rollback()
		response.ResponseError(c, 20008, err)
		return
	}
	tx.Commit()
	response.ResponseSuccess(c, "")
}

// ServiceUpdateTCP godoc
// @Summary 修改TCP服务
// @Description 修改TCP服务
// @Tags 服务管理
// @ID /service/service_update_tcp
// @Accept  json
// @Produce  json
// @Param body body dto.ServiceUpdateTCPInput true "body"
// @Success 200 {object} response.Response{data=string} "success"
// @Router /service/service_update_tcp [post]
func (service *ServiceController) ServiceUpdateTCP(c *gin.Context) {
	params := &dto.ServiceUpdateTCPInput{}
	if err := params.BindValidParam(c); err != nil {
		response.ResponseError(c, 2000, err)
		return
	}

	if len(strings.Split(params.IpList, ",")) != len(strings.Split(params.WeightList, ",")) {
		response.ResponseError(c, 20001, errors.New("IP列表与权重列表数量不一致"))
		return
	}

	tx, err := lib.GetGormPool("default")
	if err != nil {
		response.ResponseError(c, 20002, err)
		return
	}

	serviceInfo := &dao.ServiceInfo{
		ServiceName: params.ServiceName,
		IsDelete: 0,
	}
	serviceInfo, err = serviceInfo.Find(c, tx, serviceInfo)
	if err != nil {
		response.ResponseError(c, 20003, errors.New("服务不存在"))
		return
	}

	serviceDetail, err := serviceInfo.ServiceDetail(c, tx, serviceInfo)
	if err != nil {
		response.ResponseError(c, 20004, errors.New("服务不存在"))
		return
	}

	tx = tx.Begin()
	info := serviceDetail.Info
	info.ServiceDesc = params.ServiceDesc
	if err := info.Save(c, tx); err != nil {
		tx.Rollback()
		response.ResponseError(c, 20005, err)
		return
	}

	tcpRule := &dao.TcpRule{}
	if serviceDetail.TCPRule != nil {
		tcpRule = serviceDetail.TCPRule
	}
	tcpRule.ServiceID = info.ID
	tcpRule.Port = params.Port
	if err := tcpRule.Save(c, tx); err != nil {
		tx.Rollback()
		response.ResponseError(c, 20006, err)
		return
	}

	accessControl := &dao.AccessControl{}
	if serviceDetail.AccessControl != nil {
		accessControl = serviceDetail.AccessControl
	}
	accessControl.OpenAuth = params.OpenAuth
	accessControl.BlackList = params.BlackList
	accessControl.WhiteList = params.WhiteList
	accessControl.ClientIPFlowLimit = params.ClientIPFlowLimit
	accessControl.ServiceFlowLimit = params.ServiceFlowLimit
	accessControl.WhiteHostName = params.WhiteHostName
	if err := accessControl.Save(c, tx); err != nil {
		tx.Rollback()
		response.ResponseError(c, 20007, err)
		return
	}

	loadbalance := &dao.LoadBalance{}
	if serviceDetail.LoadBalance != nil {
		loadbalance = serviceDetail.LoadBalance
	}
	loadbalance.ServiceID = info.ID
	loadbalance.RoundType = params.RoundType
	loadbalance.IpList = params.IpList
	loadbalance.WeightList = params.WeightList
	loadbalance.ForbidList = params.ForbidList
	if err := loadbalance.Save(c, tx); err != nil {
		tx.Rollback()
		response.ResponseError(c, 20008, err)
		return
	}
	tx.Commit()
	response.ResponseSuccess(c, "")
}

// ServiceAddGRPC godoc
// @Summary 添加GRPC服务
// @Description 添加GRPC服务
// @Tags 服务管理
// @ID /service/service_add_grpc
// @Accept  json
// @Produce  json
// @Param body body dto.ServiceAddGRPCInput true "body"
// @Success 200 {object} response.Response{data=string} "success"
// @Router /service/service_add_grpc [post]
func (s *ServiceController) ServiceAddGRPC(c *gin.Context) {
	params := &dto.ServiceAddGRPCInput{}
	if err := params.BindValidParam(c); err != nil {
		response.ResponseError(c, 20000, err)
		return
	}

	if len(strings.Split(params.IpList, ",")) != len(strings.Split(params.WeightList, ",")) {
			response.ResponseError(c, 20004, errors.New("IP列表与权重列表数量不一致"))
			return
	}

	tx, err := lib.GetGormPool("default")
	if err != nil {
		response.ResponseError(c, 20001, err)
		return
	}

	serviceInfo := &dao.ServiceInfo{
		ServiceName: params.ServiceName,
		IsDelete: 0,
	}
	if _, err = serviceInfo.Find(c, tx, serviceInfo); err == nil {
		response.ResponseError(c, 20002, errors.New("服务名已存在"))
		return
	}

	GrpcUrl := &dao.GrpcRule{Port: params.Port}
	if _, err := GrpcUrl.Find(c, tx, GrpcUrl); err == nil {
		response.ResponseError(c, 20003, errors.New("端口被占用"))
		return
	}

	tx = tx.Begin()
	serviceModel := &dao.ServiceInfo{
		LoadType:    public.LoadTypeGRPC,
		ServiceName: params.ServiceName,
		ServiceDesc: params.ServiceDesc,
	}
	if err := serviceModel.Save(c, tx); err != nil {
		tx.Rollback()
		response.ResponseError(c, 20005, err)
		return
	}
	//serviceModel.ID
	grpcRule := &dao.GrpcRule{
		ServiceID:      serviceModel.ID,
		Port: 			params.Port,	
		HeaderTransfor: params.HeaderTransfor,
	}
	if err := grpcRule.Save(c, tx); err != nil {
		tx.Rollback()
		response.ResponseError(c, 20006, err)
		return
	}

	accessControl := &dao.AccessControl{
		ServiceID:         serviceModel.ID,
		OpenAuth:          params.OpenAuth,
		BlackList:         params.BlackList,
		WhiteList:         params.WhiteList,
		WhiteHostName: 		params.WhiteHostName,
		ClientIPFlowLimit: params.ClientIPFlowLimit,
		ServiceFlowLimit:  params.ServiceFlowLimit,
	}
	if err := accessControl.Save(c, tx); err != nil {
		tx.Rollback()
		response.ResponseError(c, 20007, err)
		return
	}

	loadbalance := &dao.LoadBalance{
		ServiceID:              serviceModel.ID,
		RoundType:              params.RoundType,
		IpList:                 params.IpList,
		WeightList:             params.WeightList,
		ForbidList: 			params.ForbidList,
	}
	if err := loadbalance.Save(c, tx); err != nil {
		tx.Rollback()
		response.ResponseError(c, 20008, err)
		return
	}
	tx.Commit()
	response.ResponseSuccess(c, "")
}

// ServiceUpdateGRPC godoc
// @Summary 修改GRPC服务
// @Description 修改GRPC服务
// @Tags 服务管理
// @ID /service/service_update_grpc
// @Accept  json
// @Produce  json
// @Param body body dto.ServiceUpdateGRPCInput true "body"
// @Success 200 {object} response.Response{data=string} "success"
// @Router /service/service_update_grpc [post]
func (service *ServiceController) ServiceUpdateGRPC(c *gin.Context) {
	params := &dto.ServiceUpdateGRPCInput{}
	if err := params.BindValidParam(c); err != nil {
		response.ResponseError(c, 2000, err)
		return
	}

	if len(strings.Split(params.IpList, ",")) != len(strings.Split(params.WeightList, ",")) {
		response.ResponseError(c, 20001, errors.New("IP列表与权重列表数量不一致"))
		return
	}

	tx, err := lib.GetGormPool("default")
	if err != nil {
		response.ResponseError(c, 20002, err)
		return
	}

	serviceInfo := &dao.ServiceInfo{
		ServiceName: params.ServiceName,
		IsDelete: 0,
	}
	serviceInfo, err = serviceInfo.Find(c, tx, serviceInfo)
	if err != nil {
		response.ResponseError(c, 20003, errors.New("服务不存在"))
		return
	}

	serviceDetail, err := serviceInfo.ServiceDetail(c, tx, serviceInfo)
	if err != nil {
		response.ResponseError(c, 20004, errors.New("服务不存在"))
		return
	}

	tx = tx.Begin()
	info := serviceDetail.Info
	info.ServiceDesc = params.ServiceDesc
	if err := info.Save(c, tx); err != nil {
		tx.Rollback()
		response.ResponseError(c, 20005, err)
		return
	}

	grpcRule := &dao.GrpcRule{}
	if serviceDetail.TCPRule != nil {
		grpcRule = serviceDetail.GRPCRule
	}
	grpcRule.ServiceID = info.ID
	grpcRule.Port = params.Port
	grpcRule.HeaderTransfor = params.HeaderTransfor
	if err := grpcRule.Save(c, tx); err != nil {
		tx.Rollback()
		response.ResponseError(c, 20006, err)
		return
	}

	accessControl := &dao.AccessControl{}
	if serviceDetail.AccessControl != nil {
		accessControl = serviceDetail.AccessControl
	}
	accessControl.OpenAuth = params.OpenAuth
	accessControl.BlackList = params.BlackList
	accessControl.WhiteList = params.WhiteList
	accessControl.ClientIPFlowLimit = params.ClientIPFlowLimit
	accessControl.ServiceFlowLimit = params.ServiceFlowLimit
	accessControl.WhiteHostName = params.WhiteHostName
	if err := accessControl.Save(c, tx); err != nil {
		tx.Rollback()
		response.ResponseError(c, 20007, err)
		return
	}

	loadbalance := &dao.LoadBalance{}
	if serviceDetail.LoadBalance != nil {
		loadbalance = serviceDetail.LoadBalance
	}
	loadbalance.ServiceID = info.ID
	loadbalance.RoundType = params.RoundType
	loadbalance.IpList = params.IpList
	loadbalance.WeightList = params.WeightList
	loadbalance.ForbidList = params.ForbidList
	if err := loadbalance.Save(c, tx); err != nil {
		tx.Rollback()
		response.ResponseError(c, 20008, err)
		return
	}
	tx.Commit()
	response.ResponseSuccess(c, "")
}

// ServiceStat godoc
// @Summary 服务统计
// @Description 服务统计
// @Tags 服务管理
// @ID /service/service_stat
// @Accept  json
// @Produce  json
// @Param id query string true "服务ID"
// @Success 200 {object} response.Response{data=dto.ServiceStatOutput} "success"
// @Router /service/service_stat [get]
func (service *ServiceController) ServiceStat(c *gin.Context) {
	params := &dto.ServiceDeleteInput{}
	if err := params.BindValidParam(c); err != nil {
		response.ResponseError(c, 2000, err)
		return
	}

	tx, err := lib.GetGormPool("default")
	if err != nil {
		response.ResponseError(c, 2001, err)
		return
	}

	//读取基本信息
	serviceInfo := &dao.ServiceInfo{ID: params.ID, IsDelete: 0}
	serviceInfo, err = serviceInfo.Find(c, tx, serviceInfo)
	if err != nil {
		response.ResponseError(c, 2002, err)
		return
	}

	// serviceDetail, err := serviceInfo.ServiceDetail(c, tx, serviceInfo)
	counter, err := public.FlowCounterHandler.GetCounter(public.FlowCountServicePrefix + serviceInfo.ServiceName)

	if err != nil {
		response.ResponseError(c, 2003, err)
		return
	}
	todayList := []int64{}
	currentTime:= time.Now()
	for i := 0; i <= currentTime.Hour(); i++ {
		dateTime := time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), i, 0, 0, 0, lib.TimeLocation)
		hourData, err := counter.GetHourData(dateTime)
		if err != nil {
			fmt.Printf("GetHourData err : %v", err)
		}
		todayList = append(todayList, hourData)
	}
	fmt.Println(todayList)
	yesterdayList := []int64{}
	yesterdayTime:= currentTime.Add(-1 * time.Duration(time.Hour*24))
	for i := 0; i <= 23; i++ {
		dateTime := time.Date(yesterdayTime.Year(), yesterdayTime.Month(), yesterdayTime.Day(), i, 0, 0, 0, lib.TimeLocation)
		hourData, _ := counter.GetHourData(dateTime)
		yesterdayList = append(yesterdayList, hourData)
	}

	response.ResponseSuccess(c, &dto.ServiceStatOutput{
		Today:     todayList,
		Yesterday: yesterdayList,
	})
}