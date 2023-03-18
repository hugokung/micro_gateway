package dao

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AccessControl struct {
	ID                int64  `json:"id" gorm:"primary_key"`
	ServiceID         int64  `json:"service_id" gorm:"column:service_id" description:"服务id"`
	OpenAuth          int    `json:"open_auth" gorm:"column:open_auth" description:"是否开启权限 1=开启"`
	BlackList         string `json:"black_list" gorm:"column:black_list" description:"黑名单ip	"`
	WhiteList         string `json:"white_list" gorm:"column:white_list" description:"白名单ip	"`
	WhiteHostName     string `json:"white_host_name" gorm:"column:white_host_name" description:"白名单主机	"`
	ClientIPFlowLimit int  `json:"clientip_flow_limit" gorm:"column:clientip_flow_limit" description:"客户端ip限流	"`
	ServiceFlowLimit  int  `json:"service_flow_limit" gorm:"column:service_flow_limit" description:"服务端限流	"`
}

func (t *AccessControl) TableName() string {
	return "gateway_service_access_control"
}

func (t *AccessControl) Find(c *gin.Context, tx *gorm.DB, search *AccessControl) (*AccessControl, error) {
	model := &AccessControl{}
	err := tx.WithContext(c).Where(search).Find(model).Error
	if err != nil {
		return nil, err
	}
	return model, nil
}

func (t *AccessControl) Save(c *gin.Context, tx *gorm.DB) error {
	if err := tx.WithContext(c).Save(t).Error; err != nil {
		return err
	}
	return nil
}

func (t *AccessControl) ListBYServiceID(c *gin.Context, tx *gorm.DB, serviceID int64) ([]AccessControl, int64, error) {
	var list []AccessControl
	var count int64
	query := tx.WithContext(c)
	query = query.Table(t.TableName()).Select("*")
	query = query.Where("service_id=?", serviceID)
	err := query.Order("id desc").Find(&list).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, 0, err
	}
	errCount := query.Count(&count).Error
	if errCount != nil {
		return nil, 0, err
	}
	return list, count, nil
}
