package dao

import (
	"net/http/httptest"
	"sync"
	"time"

	"github.com/e421083458/golang_common/lib"
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
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, 0, err
	}
	total := int64(0)
	err = query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	return appList, total, nil
}

var AppManagerHandler *AppManager

func init() {
	AppManagerHandler = NewAppManager()
}

type AppManager struct {
	AppMap   map[string]*App
	AppSlice []*App
	Locker   sync.RWMutex
	init     sync.Once
	err      error
}

func NewAppManager() *AppManager {
	return &AppManager{
		AppMap:   map[string]*App{},
		AppSlice: []*App{},
		Locker:   sync.RWMutex{},
		init:     sync.Once{},
	}
}

func (s *AppManager) GetAppList() []*App {
	return s.AppSlice
}

func (s *AppManager) LoadOnce() error {
	s.init.Do(func() {
		appInfo := &App{}
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		tx, err := lib.GetGormPool("default")
		if err != nil {
			s.err = err
			return
		}
		params := &dto.AppInfoListInput{
			PageNo: 1,
			PageSize: 99999,
		}
		list, _, err := appInfo.PageList(c, tx, params)
		if err != nil {
			s.err = err
			return
		}
		s.Locker.Lock()
		defer s.Locker.Unlock()
		for _, listItem := range list {
			tmpItem := listItem
			s.AppMap[listItem.AppID] = &tmpItem
			s.AppSlice = append(s.AppSlice, &tmpItem)
		}
	})
	return s.err
}