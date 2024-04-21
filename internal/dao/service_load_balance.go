package dao

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/hugokung/micro_gateway/pkg/load_balance"
	"github.com/hugokung/micro_gateway/pkg/public"
	"gorm.io/gorm"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

type LoadBalance struct {
	ID            int64  `json:"id" gorm:"primary_key"`
	ServiceID     int64  `json:"service_id" gorm:"column:service_id" description:"服务id	"`
	CheckMethod   int    `json:"check_method" gorm:"column:check_method" description:"检查方法 tcpchk=检测端口是否握手成功	"`
	CheckTimeout  int    `json:"check_timeout" gorm:"column:check_timeout" description:"check超时时间	"`
	CheckInterval int    `json:"check_interval" gorm:"column:check_interval" description:"检查间隔, 单位s		"`
	RoundType     int    `json:"round_type" gorm:"column:round_type" description:"轮询方式 round/weight_round/random/ip_hash"`
	IpList        string `json:"ip_list" gorm:"column:ip_list" description:"ip列表"`
	WeightList    string `json:"weight_list" gorm:"column:weight_list" description:"权重列表"`
	ForbidList    string `json:"forbid_list" gorm:"column:forbid_list" description:"禁用ip列表"`

	UpstreamConnectTimeout int `json:"upstream_connect_timeout" gorm:"column:upstream_connect_timeout" description:"下游建立连接超时, 单位s"`
	UpstreamHeaderTimeout  int `json:"upstream_header_timeout" gorm:"column:upstream_header_timeout" description:"下游获取header超时, 单位s	"`
	UpstreamIdleTimeout    int `json:"upstream_idle_timeout" gorm:"column:upstream_idle_timeout" description:"下游链接最大空闲时间, 单位s	"`
	UpstreamMaxIdle        int `json:"upstream_max_idle" gorm:"column:upstream_max_idle" description:"下游最大空闲链接数"`
}

func (t *LoadBalance) TableName() string {
	return "gateway_service_load_balance"
}

func (t *LoadBalance) Find(c *gin.Context, tx *gorm.DB, search *LoadBalance) (*LoadBalance, error) {
	model := &LoadBalance{}
	err := tx.WithContext(c).Where(search).Find(model).Error
	return model, err
}

func (t *LoadBalance) Save(c *gin.Context, tx *gorm.DB) error {
	if err := tx.WithContext(c).Save(t).Error; err != nil {
		return err
	}
	return nil
}

func (t *LoadBalance) GetIPListByModel() []string {
	return strings.Split(t.IpList, ",")
}

func (t *LoadBalance) GetWeightListByModel() []string {
	return strings.Split(t.WeightList, ",")
}

type LoadBalancer struct {
	LoadBalanceMap   map[string]*LoadBalancerItem
	LoadBalanceSlice []*LoadBalancerItem
	Locker           sync.RWMutex
}
type LoadBalancerItem struct {
	LoadBalance load_balance.LoadBalance
	ServiceName string
	UpdatedAt   time.Time
}

var LoadBalancerHandler *LoadBalancer

func NewLoadBalancer() *LoadBalancer {
	return &LoadBalancer{
		LoadBalanceMap:   make(map[string]*LoadBalancerItem),
		LoadBalanceSlice: []*LoadBalancerItem{},
		Locker:           sync.RWMutex{},
	}
}

func init() {
	LoadBalancerHandler = NewLoadBalancer()
	ServiceManagerHandler.Register(LoadBalancerHandler)
	TransportorHandler = NewTransportor()
	ServiceManagerHandler.Register(TransportorHandler)
}

func (lbr *LoadBalancer) Update(e *ServiceEvent) {
	log.Printf("LoadBalancer.Update\n")
	for _, service := range e.AddService {
		lbr.GetLoadBalancer(service)
	}
	for _, service := range e.UpdateService {
		lbr.GetLoadBalancer(service)
	}
	var newLBSlice []*LoadBalancerItem
	for _, lbrItem := range lbr.LoadBalanceSlice {
		matched := false
		for _, service := range e.DeleteService {
			if lbrItem.ServiceName == service.Info.ServiceName {
				lbrItem.LoadBalance.Close()
				matched = true
			}
		}
		if matched {
			delete(lbr.LoadBalanceMap, lbrItem.ServiceName)
		} else {
			newLBSlice = append(newLBSlice, lbrItem)
		}
	}
	lbr.LoadBalanceSlice = newLBSlice
}

func GetLoadBalancerConf(service *ServiceDetail) (load_balance.LoadBalanceConf, error) {
	schema := "http://"
	if service.HTTPRule.NeedHttps == 1 {
		schema = "https://"
	}
	if service.Info.LoadType == public.LoadTypeTCP || service.Info.LoadType == public.LoadTypeGRPC {
		schema = ""
	}

	switch service.Info.ServiceDiscovery {
	case public.StaticConfig:
		ipList := service.LoadBalance.GetIPListByModel()
		weightList := service.LoadBalance.GetWeightListByModel()
		ipConf := map[string]string{}
		for idx, item := range ipList {
			ipConf[item] = weightList[idx]
		}
		mConf, err := load_balance.NewLoadBalanceCheckConf(service.Info.ServiceName,
			fmt.Sprintf("%s%s", schema, "%s"), ipConf)
		if err != nil {
			return nil, err
		}
		return mConf, nil
	case public.ZookeeperConfig:
		ipConf := map[string]string{}
		mConf, err := load_balance.NewLoadBalanceZkConf(fmt.Sprintf("%s%s", schema, "%s"),
			"/"+service.Info.ServiceName,
			service.Environment.GetIPListByModel(), ipConf)
		if err != nil {
			return nil, err
		}
		return mConf, nil
	case public.EtcdConfig:
		ipConf := map[string]string{}
		mConf, err := load_balance.NewLoadBalanceEtcdConf(fmt.Sprintf("%s%s", schema, "%s"), service.Info.ServiceName,
			service.Environment.GetIPListByModel(), ipConf)
		if err != nil {
			return nil, err
		}
		return mConf, nil
	}
	return nil, errors.New("Discovery type not exist")
}

func (lbr *LoadBalancer) GetLoadBalancer(service *ServiceDetail) (load_balance.LoadBalance, error) {
	for _, lbrItem := range lbr.LoadBalanceSlice {
		if lbrItem.ServiceName == service.Info.ServiceName && lbrItem.UpdatedAt == service.Info.UpdatedAt {
			return lbrItem.LoadBalance, nil
		}
	}
	mConf, err := GetLoadBalancerConf(service)
	if err != nil {
		return nil, err
	}
	lb := load_balance.LoadBanlanceFactorWithConf(load_balance.LbType(service.LoadBalance.RoundType), mConf)

	//save to map and slice
	matched := false
	for _, lbrItem := range lbr.LoadBalanceSlice {
		if lbrItem.ServiceName == service.Info.ServiceName {
			matched = true
			lbrItem.LoadBalance.Close()
			lbrItem.LoadBalance = lb
			lbrItem.UpdatedAt = service.Info.UpdatedAt
		}
	}
	if !matched {
		lbItem := &LoadBalancerItem{
			LoadBalance: lb,
			ServiceName: service.Info.ServiceName,
			UpdatedAt:   service.Info.UpdatedAt,
		}
		lbr.LoadBalanceSlice = append(lbr.LoadBalanceSlice, lbItem)
		lbr.Locker.Lock()
		defer lbr.Locker.Unlock()
		lbr.LoadBalanceMap[service.Info.ServiceName] = lbItem
	}
	return lb, nil
}

type Transportor struct {
	TransportMap   map[string]*TransportItem
	TransportSlice []*TransportItem
	Locker         sync.Locker
}
type TransportItem struct {
	Trans       *RetryTransport
	ServiceName string
	UpdateAt    time.Time
}

type RetryTransport struct {
	Transport http.RoundTripper
	Retries   int
}

func (r *RetryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	var resp *http.Response
	var err error

	for i := 0; i <= r.Retries; i++ {
		resp, err = r.Transport.RoundTrip(req)

		if err == nil && resp.StatusCode < 500 {
			break
		}
		// TODO: 如何设计更合理的重试间隔
		time.Sleep(1 * time.Second)
	}
	return resp, err
}

func newRetryTransport(baseTransport http.RoundTripper, retries int) *RetryTransport {
	return &RetryTransport{
		Transport: baseTransport,
		Retries:   retries,
	}
}

var TransportorHandler *Transportor

func NewTransportor() *Transportor {
	return &Transportor{
		TransportMap:   map[string]*TransportItem{},
		TransportSlice: []*TransportItem{},
		Locker:         &sync.RWMutex{},
	}
}

func (t *Transportor) Update(e *ServiceEvent) {
	log.Printf("Transportor.Update\n")
	for _, service := range e.AddService {
		t.GetTransportor(service)
	}
	for _, service := range e.UpdateService {
		t.GetTransportor(service)
	}
	var newSlice []*TransportItem
	for _, tItem := range t.TransportSlice {
		matched := false
		for _, service := range e.DeleteService {
			if tItem.ServiceName == service.Info.ServiceName {
				matched = true
			}
		}
		if matched {
			delete(t.TransportMap, tItem.ServiceName)
		} else {
			newSlice = append(newSlice, tItem)
		}
	}
	t.TransportSlice = newSlice
}

func (t *Transportor) GetTransportor(service *ServiceDetail) (*RetryTransport, error) {
	for _, transItem := range t.TransportSlice {
		if transItem.ServiceName == service.Info.ServiceName {
			return transItem.Trans, nil
		}
	}

	if service.LoadBalance.UpstreamConnectTimeout == 0 {
		service.LoadBalance.UpstreamConnectTimeout = 30
	}
	if service.LoadBalance.UpstreamMaxIdle == 0 {
		service.LoadBalance.UpstreamMaxIdle = 4000
	}
	if service.LoadBalance.UpstreamIdleTimeout == 0 {
		service.LoadBalance.UpstreamIdleTimeout = 90
	}
	if service.LoadBalance.UpstreamHeaderTimeout == 0 {
		service.LoadBalance.UpstreamHeaderTimeout = 30
	}
	perhost := 0
	if len(service.LoadBalance.GetIPListByModel()) > 0 {
		perhost = service.LoadBalance.UpstreamMaxIdle / len(service.LoadBalance.GetIPListByModel())
	}
	trans := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   time.Duration(service.LoadBalance.UpstreamConnectTimeout) * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConnsPerHost:   perhost,
		MaxIdleConns:          service.LoadBalance.UpstreamMaxIdle,
		WriteBufferSize:       1 << 18, //200m
		ReadBufferSize:        1 << 18, //200m
		IdleConnTimeout:       time.Duration(service.LoadBalance.UpstreamIdleTimeout) * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: time.Duration(service.LoadBalance.UpstreamHeaderTimeout) * time.Second,
	}

	//save to map and slice
	matched := false
	retryTrans := newRetryTransport(trans, service.HTTPRule.Retries)
	for _, transItem := range t.TransportSlice {
		if transItem.ServiceName == service.Info.ServiceName {
			matched = true
			transItem.Trans = retryTrans
			transItem.UpdateAt = service.Info.UpdatedAt
		}
	}
	if !matched {
		transItem := &TransportItem{
			Trans:       retryTrans,
			ServiceName: service.Info.ServiceName,
			UpdateAt:    service.Info.UpdatedAt,
		}
		t.TransportSlice = append(t.TransportSlice, transItem)
		t.Locker.Lock()
		defer t.Locker.Unlock()
		t.TransportMap[service.Info.ServiceName] = transItem
	}
	return retryTrans, nil
}
