package dao

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hugokung/micro_gateway/dto"
	"gorm.io/gorm"
)

type App struct {
	ID        int64     `json:"id" gorm:"primary_key"`
	AppID     string    `json:"app_id" gorm:"column:app_id" description:"租户id	"`
	Name      string    `json:"name" gorm:"column:name" description:"租户名称	"`
	Secret    string    `json:"secret" gorm:"column:secret" description:"密钥"`
	WhiteIPS  string    `json:"white_ips" gorm:"column:white_ips" description:"ip白名单，支持前缀匹配"`
	Qpd       int64     `json:"qpd" gorm:"column:qpd" description:"日请求量限制"`
	Qps       int64     `json:"qps" gorm:"column:qps" description:"每秒请求量限制"`
	CreatedAt time.Time `json:"create_at" gorm:"column:create_at" description:"添加时间	"`
	UpdatedAt time.Time `json:"update_at" gorm:"column:update_at" description:"更新时间"`
	IsDelete  int8      `json:"is_delete" gorm:"column:is_delete" description:"是否已删除；0：否；1：是"`
}

func (t *App) TableName() string {
	return "gateway_app"
}

func (a *App) Find(c *gin.Context, tx *gorm.DB, search *App) (*App, error) {
	app := &App{}
	query := tx.WithContext(c).Where(search)
	err := query.Where("is_delete", 0).First(app).Error
	if err != nil {
		return nil, err
	}
	return app, err
}

func (a *App) Save(c *gin.Context, tx *gorm.DB) error {
	return tx.WithContext(c).Save(a).Error
}

func (a *App) PageList(c *gin.Context, tx *gorm.DB, search *dto.AppInfoListInput) ([]App, int64, error) {
	keyword := ""
	query := tx.WithContext(c)
	if search.Info != "" {
		keyword = "%" + search.Info + "%"
		query = query.Where("(name like ? or app_id like ?)", keyword, keyword)
	}
	offset := (search.PageNo - 1) * search.PageSize
	query = query.Table(a.TableName()).Where("is_delete = ?", 0).Offset(offset).Limit(search.PageSize)
	appList := []App{}
	err := query.Find(&appList).Error
	if err != nil && err != gorm.ErrRecordNotFound{
		return nil, 0, err
	}
	total := int64(0)
	err = query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	return appList, total, nil
}