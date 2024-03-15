package dao

import (
	"log"
	"net/http/httptest"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hugokung/micro_gateway/internal/dto"
	"github.com/hugokung/micro_gateway/pkg/golang_common/lib"
	"github.com/hugokung/micro_gateway/pkg/public"
	"github.com/pkg/errors"
)

type ServiceDetail struct {
	Info          *ServiceInfo   `json:"info" description:"基本信息"`
	HTTPRule      *HttpRule      `json:"http_rule" description:"http代理规则"`
	TCPRule       *TcpRule       `json:"tcp_rule" description:"tcp代理规则"`
	GRPCRule      *GrpcRule      `json:"grpc_rule" description:"grpc代理规则"`
	LoadBalance   *LoadBalance   `json:"load_balance" description:"负载均衡"`
	AccessControl *AccessControl `json:"access_control" description:"权限校验"`
	CircuitConfig *CircuitConfig `json:"circuit_config" description:"熔断配置"`
	Environment   *Environment   `json:"environment" description:"服务发现"`
}

var ServiceManagerHandler *ServiceManager

func init() {
	ServiceManagerHandler = NewServiceManager()
}

// ServiceEvent 通知事件
type ServiceEvent struct {
	DeleteService []*ServiceDetail
	AddService    []*ServiceDetail
	UpdateService []*ServiceDetail
}

// ServiceObserver 观察者接口
type ServiceObserver interface {
	Update(*ServiceEvent)
}

// ServiceSubject 被观察者接口
type ServiceSubject interface {
	Register(ServiceObserver)
	Deregister(ServiceObserver)
	Notify(*ServiceEvent)
}

func (s *ServiceManager) Register(ob ServiceObserver) {
	s.Observers[ob] = true
}

func (s *ServiceManager) Deregister(ob ServiceObserver) {
	delete(s.Observers, ob)
}

func (s *ServiceManager) Notify(e *ServiceEvent) {
	for ob, _ := range s.Observers {
		ob.Update(e)
	}
}

type ServiceManager struct {
	ServiceMap   map[string]*ServiceDetail
	ServiceSlice []*ServiceDetail
	err          error
	UpdateAt     time.Time
	Observers    map[ServiceObserver]bool
}

func NewServiceManager() *ServiceManager {
	return &ServiceManager{
		ServiceMap:   map[string]*ServiceDetail{},
		ServiceSlice: []*ServiceDetail{},
		Observers:    map[ServiceObserver]bool{},
	}
}

func (s *ServiceManager) HttpAccessMode(c *gin.Context) (*ServiceDetail, error) {
	//1、前缀匹配 /abc ==> serviceSlice.rule
	//2、域名匹配 www.test.com ==> serviceSlice.rule
	//host := c.Request.Host[0:strings.Index(c.Request.Host, ":")]
	for _, serviceItem := range s.ServiceSlice {
		if serviceItem.Info.LoadType != public.LoadTypeHTTP {
			continue
		}
		if serviceItem.HTTPRule.RuleType == public.RuleTypeDomin {
			if serviceItem.HTTPRule.Rule == c.Request.Host[0:strings.Index(c.Request.Host, ":")] {
				return serviceItem, nil
			}
		}
		if serviceItem.HTTPRule.RuleType == public.RuleTypePrefixURL {
			if strings.HasPrefix(c.Request.URL.Path, serviceItem.HTTPRule.Rule) {
				return serviceItem, nil
			}
		}
	}
	return nil, errors.New("not matched service")
}

func (s *ServiceManager) LoadService() *ServiceManager {
	//log.Printf(" [INFO] ServiceManager.LoadService begin\n")
	ns := NewServiceManager()
	defer func() {
		if ns.err != nil {
			log.Printf(" [ERROR] ServiceManager.LoadService error:%v\n", ns.err)
		}
	}()
	serviceInfo := &ServiceInfo{}
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("GET", "/", nil)
	tx, err := lib.GetGormPool("default")
	if err != nil {
		ns.err = err
		return ns
	}
	params := &dto.ServiceInfoInput{PageNo: 1, PageSize: 99999}
	list, _, err := serviceInfo.PageList(c, tx, params)
	if err != nil {
		ns.err = err
		return ns
	}
	for _, listItem := range list {
		tmpItem := listItem
		serviceDetail, err := tmpItem.ServiceDetail(c, tx, &tmpItem)
		if err != nil {
			ns.err = err
			return ns
		}
		ns.ServiceMap[listItem.ServiceName] = serviceDetail
		ns.ServiceSlice = append(ns.ServiceSlice, serviceDetail)
		if listItem.UpdatedAt.Unix() > ns.UpdateAt.Unix() {
			ns.UpdateAt = listItem.UpdatedAt
		}
	}
	//log.Printf(" [INFO] ServiceManager.LoadService end\n")
	return ns
}

// LoadAndWatch 动态更新API配置
func (s *ServiceManager) LoadAndWatch() error {
	ns := s.LoadService()
	if ns.err != nil {
		return ns.err
	}
	s.ServiceSlice = ns.ServiceSlice
	s.ServiceMap = ns.ServiceMap
	s.UpdateAt = ns.UpdateAt
	e := &ServiceEvent{AddService: ns.ServiceSlice}
	s.Notify(e)
	go func() {
		// db定时检查update_time是否变更过
		for {
			time.Sleep(10 * time.Second)
			ns := s.LoadService()
			if ns.err != nil {
				log.Printf("ns.err:%v ns.UpdateAt:%v\n", ns.err, ns.UpdateAt)
				continue
			}
			if ns.UpdateAt != s.UpdateAt || len(ns.ServiceSlice) != len(s.ServiceSlice) {
				log.Printf("ServiceManager s.UpdateAt:%v ns.UpdateAt:%v\n", s.UpdateAt.Format(lib.TimeFormat), ns.UpdateAt.Format(lib.TimeFormat))
				e := &ServiceEvent{}

				//老服务存在，新服务不存在，则为删除
				for _, service := range s.ServiceSlice {
					matched := false
					for _, newService := range ns.ServiceSlice {
						if service.Info.ServiceName == newService.Info.ServiceName {
							matched = true
						}
					}
					if !matched {
						e.DeleteService = append(e.DeleteService, service)
					}
				}
				//新服务有，老服务不存在，则为新增
				for _, newService := range ns.ServiceSlice {
					matched := false
					for _, service := range s.ServiceSlice {
						if service.Info.ServiceName == newService.Info.ServiceName {
							matched = true
						}
					}
					if !matched {
						e.AddService = append(e.AddService, newService)
					}
				}
				//服务名相同，更新时间不同，则为更新
				for _, newService := range ns.ServiceSlice {
					matched := false
					for _, service := range s.ServiceSlice {
						if service.Info.ServiceName == newService.Info.ServiceName && service.Info.UpdatedAt != newService.Info.UpdatedAt {
							matched = true
						}
					}
					if matched {
						e.UpdateService = append(e.UpdateService, newService)
					}
				}
				s.ServiceSlice = ns.ServiceSlice
				s.ServiceMap = ns.ServiceMap
				s.UpdateAt = ns.UpdateAt

				log.Printf("ServiceManager e:%v delScv.len=%d addScv.len=%d uploadScv.len=%d\n", e, len(e.DeleteService), len(e.AddService), len(e.UpdateService))
				s.Notify(e)
			}
		}
	}()
	return s.err
}

func (s *ServiceManager) GetTcpServiceList() []*ServiceDetail {
	var list []*ServiceDetail
	for _, serviceItem := range s.ServiceSlice {
		tmpItem := serviceItem
		if tmpItem.Info.LoadType == public.LoadTypeTCP {
			list = append(list, tmpItem)
		}
	}
	return list
}

func (s *ServiceManager) GetGrpcServiceList() []*ServiceDetail {
	var list []*ServiceDetail
	for _, serviceItem := range s.ServiceSlice {
		tmpItem := serviceItem
		if tmpItem.Info.LoadType == public.LoadTypeGRPC {
			list = append(list, tmpItem)
		}
	}
	return list
}
