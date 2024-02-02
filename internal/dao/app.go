package dao

import (
	"log"
	"net/http/httptest"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hugokung/micro_gateway/internal/dto"
	"github.com/hugokung/micro_gateway/pkg/golang_common/lib"
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
	var appList []App
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

// AppEvent 通知事件
type AppEvent struct {
	DeleteApp []*App
	AddApp    []*App
	UpdateApp []*App
}

// AppObserver 观察者接口
type AppObserver interface {
	Update(*AppEvent)
}

// AppSubject 被观察者接口
type AppSubject interface {
	Register(ServiceObserver)
	Deregister(ServiceObserver)
	Notify(*AppEvent)
}

func (s *AppManager) Register(ob AppObserver) {
	s.Observers[ob] = true
}

func (s *AppManager) Deregister(ob AppObserver) {
	delete(s.Observers, ob)
}

func (s *AppManager) Notify(e *AppEvent) {
	for ob, _ := range s.Observers {
		ob.Update(e)
	}
}

type AppManager struct {
	AppMap    map[string]*App
	AppSlice  []*App
	err       error
	UpdateAt  time.Time
	Observers map[AppObserver]bool
}

func NewAppManager() *AppManager {
	return &AppManager{
		AppMap:   map[string]*App{},
		AppSlice: []*App{},
	}
}

func (s *AppManager) GetAppList() []*App {
	return s.AppSlice
}

func (s *AppManager) LoadApp() *AppManager {
	//log.Printf(" [INFO] AppManager.LoadApp begin\n")
	ns := NewAppManager()
	defer func() {
		if ns.err != nil {
			log.Printf(" [ERROR] AppManager.LoadApp error:%v\n", ns.err)
		}
	}()
	appInfo := &App{}
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("GET", "/", nil)
	tx, err := lib.GetGormPool("default")
	if err != nil {
		ns.err = err
		return ns
	}
	params := &dto.AppInfoListInput{PageNo: 1, PageSize: 99999}
	list, _, err := appInfo.PageList(c, tx, params)
	if err != nil {
		ns.err = err
		return ns
	}
	for _, listItem := range list {
		tmpItem := listItem
		ns.AppMap[listItem.AppID] = &tmpItem
		ns.AppSlice = append(ns.AppSlice, &tmpItem)
		if listItem.UpdatedAt.Unix() > ns.UpdateAt.Unix() {
			ns.UpdateAt = listItem.UpdatedAt
		}
	}
	//log.Printf(" [INFO] AppManager.LoadApp end\n")
	return ns
}

// LoadAndWatch 动态更新API配置
func (s *AppManager) LoadAndWatch() error {
	ns := s.LoadApp()
	if ns.err != nil {
		return ns.err
	}
	s.AppSlice = ns.AppSlice
	s.AppMap = ns.AppMap
	s.UpdateAt = ns.UpdateAt
	go func() {
		for true {
			time.Sleep(10 * time.Second)
			ns := s.LoadApp()
			if ns.err != nil {
				continue
			}
			if ns.UpdateAt != s.UpdateAt || len(ns.AppSlice) != len(s.AppSlice) {
				log.Printf("s.UpdateAt:%v ns.UpdateAt:%v\n", s.UpdateAt.Format(lib.TimeFormat), ns.UpdateAt.Format(lib.TimeFormat))
				e := &AppEvent{}

				//老服务存在，新服务不存在，则为删除
				for _, app := range s.AppSlice {
					matched := false
					for _, newApp := range ns.AppSlice {
						if app.AppID == newApp.AppID {
							matched = true
						}
					}
					if !matched {
						e.DeleteApp = append(e.DeleteApp, app)
					}
				}
				//新服务有，老服务不存在，则为新增
				for _, newApp := range ns.AppSlice {
					matched := false
					for _, app := range s.AppSlice {
						if app.AppID == newApp.AppID {
							matched = true
						}
					}
					if !matched {
						e.AddApp = append(e.AddApp, newApp)
					}
				}
				//服务名相同，更新时间不同，则为更新
				for _, newApp := range ns.AppSlice {
					matched := false
					for _, app := range s.AppSlice {
						if app.AppID == newApp.AppID && app.UpdatedAt != newApp.UpdatedAt {
							matched = true
						}
					}
					if matched {
						e.UpdateApp = append(e.UpdateApp, newApp)
					}
				}
				s.AppSlice = ns.AppSlice
				s.AppMap = ns.AppMap
				s.UpdateAt = ns.UpdateAt

				log.Printf("e:%v\n", e)
				s.Notify(e)
			}
		}
	}()
	return s.err
}
