package dao

import (
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Environment struct {
	ID       int64  `json:"id" gorm:"primary_key"`
	EnvName  int64  `json:"env_name" gorm:"column:env_name" description:"服务发现的环境名称"`
	IpList   string `json:"ip_list" gorm:"column:ip_list" description:"ip 地址"`
	IsDelete int    `json:"is_delete" gorm:"column:is_delete" description:"是否被删除"`
}

func (e *Environment) TableName() string {
	return "gateway_service_environment"
}

func (e *Environment) Find(c *gin.Context, tx *gorm.DB, search *Environment) (*Environment, error) {
	model := &Environment{}
	err := tx.WithContext(c).Where(search).Find(model).Error
	return model, err
}

func (e *Environment) FindAll(c *gin.Context, tx *gorm.DB) ([]*Environment, error) {
	list := []*Environment{}
	err := tx.WithContext(c).Where("is_delete = ?", 0).Find(list).Error
	if err != nil {
		return nil, err
	}
	return list, nil
}

func (e *Environment) Save(c *gin.Context, tx *gorm.DB) error {
	if err := tx.WithContext(c).Save(e).Error; err != nil {
		return err
	}
	return nil
}

func (e *Environment) GetIPListByModel() []string {
	return strings.Split(e.IpList, ",")
}
