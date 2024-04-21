package dao

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hugokung/micro_gateway/internal/dto"
	"gorm.io/gorm"
)

type ServiceInfo struct {
	ID               int64     `json:"id" gorm:"primary_key" description:"自增主键"`
	LoadType         int       `json:"load_type" gorm:"column:load_type" description:"负载类型"`
	ServiceName      string    `json:"service_name" gorm:"column:service_name" description:"服务名称"`
	ServiceDesc      string    `json:"service_desc" gorm:"column:service_desc" description:"服务描述"`
	ServiceDiscovery int       `json:"service_discovery" gorm:"column:service_discovery" description:"服务发现类型"`
	UpdatedAt        time.Time `json:"update_at" gorm:"column:update_at" description:"更新时间"`
	CreatedAt        time.Time `json:"create_at" gorm:"column:create_at" description:"创建时间"`
	IsDelete         int       `json:"is_delete" gorm:"column:is_delete" description:"是否删除"`
}

func (t *ServiceInfo) TableName() string {
	return "gateway_service_info"
}

func (s *ServiceInfo) PageList(c *gin.Context, tx *gorm.DB, param *dto.ServiceInfoInput) ([]ServiceInfo, int64, error) {
	list := []ServiceInfo{}
	offset := (param.PageNo - 1) * param.PageSize
	query := tx.WithContext(c)
	query = query.Limit(param.PageSize).Offset(offset)
	query = query.Table(s.TableName()).Where("is_delete = ?", 0)
	if param.Info != "" {
		query = query.Where("(service_name like ? or service_desc like ?)", "%"+param.Info+"%", "%"+param.Info+"%")
	}
	err := query.Find(&list).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, 0, err
	}
	var total int64 = 0
	err = query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (a *ServiceInfo) Find(c *gin.Context, tx *gorm.DB, search *ServiceInfo) (*ServiceInfo, error) {
	out := &ServiceInfo{}
	err := tx.WithContext(c).Where(search).First(out).Error
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (s *ServiceInfo) GroupByLoadType(c *gin.Context, tx *gorm.DB) ([]dto.DashServiceStatItemOutput, error) {
	items := []dto.DashServiceStatItemOutput{}
	query := tx.WithContext(c)
	query = query.Table(s.TableName()).Where("is_delete = ?", 0)
	if err := query.Select("load_type, count(*) as value").Group("load_type").Scan(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (a *ServiceInfo) Save(c *gin.Context, tx *gorm.DB) error {
	return tx.WithContext(c).Save(a).Error
}

func (t *ServiceInfo) ServiceDetail(c *gin.Context, tx *gorm.DB, search *ServiceInfo) (*ServiceDetail, error) {
	if search.ServiceName == "" {
		info, err := t.Find(c, tx, search)
		if err != nil {
			return nil, err
		}
		search = info
	}
	httpRule := &HttpRule{ServiceID: search.ID}
	httpRule, err := httpRule.Find(c, tx, httpRule)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	tcpRule := &TcpRule{ServiceID: search.ID}
	tcpRule, err = tcpRule.Find(c, tx, tcpRule)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	grpcRule := &GrpcRule{ServiceID: search.ID}
	grpcRule, err = grpcRule.Find(c, tx, grpcRule)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	accessControl := &AccessControl{ServiceID: search.ID}
	accessControl, err = accessControl.Find(c, tx, accessControl)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	loadBalance := &LoadBalance{ServiceID: search.ID}
	loadBalance, err = loadBalance.Find(c, tx, loadBalance)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	environment := &Environment{ServiceID: search.ID}
	environment, err = environment.Find(c, tx, environment)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	circuitConfig := &CircuitConfig{ServiceID: search.ID}
	circuitConfig, err = circuitConfig.Find(c, tx, circuitConfig)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	detail := &ServiceDetail{
		Info:          search,
		HTTPRule:      httpRule,
		TCPRule:       tcpRule,
		GRPCRule:      grpcRule,
		LoadBalance:   loadBalance,
		AccessControl: accessControl,
		Environment:   environment,
		CircuitConfig: circuitConfig,
	}
	return detail, nil
}
