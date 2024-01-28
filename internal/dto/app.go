package dto

import (
	"time"
	"github.com/hugokung/micro_gateway/pkg/public"
	"github.com/gin-gonic/gin"
)

type AppAddInput struct {
	AppID    string `json:"app_id" form:"app_id" comment:"租户id" validate:"required"`
	Name     string `json:"name" form:"name" comment:"租户名称" validate:"required"`
	Secret   string `json:"secret" form:"secret" comment:"密钥" validate:""`
	WhiteIPS string `json:"white_ips" form:"white_ips" comment:"ip白名单，支持前缀匹配"`
	Qpd      int64  `json:"qpd" form:"qpd" comment:"日请求量限制" validate:""`
	Qps      int64  `json:"qps" form:"qps" comment:"每秒请求量限制" validate:""`
}

func (a *AppAddInput) GetValidParams(c *gin.Context) error {
	return public.DefaultGetValidParams(c, a)
}

type AppDeleteInput struct {
	ID	int64	`json:"id" form:"id" comment:"租户id" validate:"required"`
}

func (a *AppDeleteInput) GetValidParams(c *gin.Context) error {
	return public.DefaultGetValidParams(c, a)
}

type AppUpdateInput struct {
	ID		 int64	`json:"id" form:"id" comment:"租户id"`
	AppID    string `json:"app_id" form:"app_id" comment:"租户id" validate:"required"`
	Name     string `json:"name" form:"name" comment:"租户名称" validate:"required"`
	Secret   string `json:"secret" form:"secret" comment:"密钥" validate:""`
	WhiteIPS string `json:"white_ips" form:"white_ips" comment:"ip白名单，支持前缀匹配"`
	Qpd      int64  `json:"qpd" form:"qpd" comment:"日请求量限制" validate:""`
	Qps      int64  `json:"qps" form:"qps" comment:"每秒请求量限制" validate:""`
}

func (a *AppUpdateInput) GetValidParams(c *gin.Context) error {
	return public.DefaultGetValidParams(c, a)
}

type AppDetailInput struct {
	ID	int64	`json:"id" form:"id" comment:"租户id" validate:"required"`
}

func (a *AppDetailInput) GetValidParams(c *gin.Context) error {
	return public.DefaultGetValidParams(c, a)
}

type AppInfoListInput struct {
	Info		string		`json:"info" form:"info" comment:"关键字" example:"http"`
	PageNo		int			`json:"page_no" form:"page_no" comment:"页数" example:"1" validate:"required,min=1,max=999"`
	PageSize	int			`json:"page_size" form:"page_size" comment:"页面大小" example:"3" validate:"required,min=1,max=999"`
}

func (a *AppInfoListInput) GetValidParams(c *gin.Context) error {
	return public.DefaultGetValidParams(c, a)
}

type AppListOutput struct {
	List  []AppListItemOutput `json:"list" form:"list" comment:"租户列表"`
	Total int64               `json:"total" form:"total" comment:"租户总数"`
}

type AppListItemOutput struct {
	ID        int64     `json:"id" gorm:"primary_key"`
	AppID     string    `json:"app_id" gorm:"column:app_id" description:"租户id	"`
	Name      string    `json:"name" gorm:"column:name" description:"租户名称	"`
	Secret    string    `json:"secret" gorm:"column:secret" description:"密钥"`
	WhiteIPS  string    `json:"white_ips" gorm:"column:white_ips" description:"ip白名单，支持前缀匹配		"`
	Qpd       int64     `json:"qpd" gorm:"column:qpd" description:"日请求量限制"`
	Qps       int64     `json:"qps" gorm:"column:qps" description:"每秒请求量限制"`
	RealQpd   int64       `json:"real_qpd" description:"日请求量限制"`
	RealQps   int64       `json:"real_qps" description:"每秒请求量限制"`
	UpdatedAt time.Time `json:"create_at" gorm:"column:create_at" description:"添加时间	"`
	CreatedAt time.Time `json:"update_at" gorm:"column:update_at" description:"更新时间"`
	IsDelete  int8      `json:"is_delete" gorm:"column:is_delete" description:"是否已删除；0：否；1：是"`
}

type StatisticsOutput struct {
	Today     []int64 `json:"today" form:"today" comment:"今日统计" validate:"required"`
	Yesterday []int64 `json:"yesterday" form:"yesterday" comment:"昨日统计" validate:"required"`
}