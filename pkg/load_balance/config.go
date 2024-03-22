package load_balance

import (
	"fmt"
	"log"

	"github.com/hugokung/micro_gateway/pkg/etcd"
	"github.com/hugokung/micro_gateway/pkg/zookeeper"
)

// LoadBalanceConf 配置主题
type LoadBalanceConf interface {
	Attach(o Observer)
	GetConf() []string
	WatchConf()
	UpdateConf(conf []string)
	CloseWatch()
}

type LoadBalanceZkConf struct {
	observers    []Observer
	path         string
	zkHosts      []string
	confIpWeight map[string]string
	activeList   []string
	format       string
	name         string
	closeChan    chan bool
}

func (s *LoadBalanceZkConf) Attach(o Observer) {
	s.observers = append(s.observers, o)
}

func (s *LoadBalanceZkConf) NotifyAllObservers() {
	for _, obs := range s.observers {
		obs.Update()
	}
}

func (s *LoadBalanceZkConf) GetConf() []string {
	confList := []string{}
	for _, ip := range s.activeList {
		weight, ok := s.confIpWeight[ip]
		if !ok {
			weight = "50" //默认weight
		}
		confList = append(confList, fmt.Sprintf(s.format, ip)+","+weight)
	}
	return confList
}

// WatchConf 更新配置时，通知监听者也更新
func (s *LoadBalanceZkConf) WatchConf() {
	zkManager := zookeeper.NewZkManager(s.zkHosts)
	zkManager.GetConnect()
	fmt.Println("watchConf")
	chanList, chanErr := zkManager.WatchServerListByPath(s.path)
	go func() {
		defer zkManager.Close()
		for {
			select {
			case changeErr := <-chanErr:
				fmt.Println("changeErr", changeErr)
			case changedList := <-chanList:
				fmt.Println("watch node changed")
				s.UpdateConf(changedList)
			}
		}
	}()
}

// UpdateConf 更新配置时，通知监听者也更新
func (s *LoadBalanceZkConf) UpdateConf(conf []string) {
	s.activeList = conf
	for _, obs := range s.observers {
		obs.Update()
	}
}

func NewLoadBalanceEtcdConf(format, path string, etcdHosts []string, conf map[string]string) (*LoadBalanceEtcdConf, error) {
	etcdManager, err := etcd.NewEtcdManager(etcdHosts, 5)
	if err != nil {
		log.Fatalf("etcd.NewEtcdManager error: %v", err)
	}
	defer etcdManager.Close()
	elist, err := etcdManager.GetServerListByPrefix(path)
	if err != nil {
		return nil, err
	}
	mConf := &LoadBalanceEtcdConf{format: format, activeList: elist, confIpWeight: conf, etcdHosts: etcdHosts, path: path, closeChan: make(chan bool, 1)}
	mConf.WatchConf()
	return mConf, nil
}

func (s *LoadBalanceEtcdConf) CloseWatch() {
	s.closeChan <- true
	close(s.closeChan)
}

type LoadBalanceEtcdConf struct {
	observers    []Observer
	path         string
	etcdHosts    []string
	confIpWeight map[string]string
	activeList   []string
	format       string
	name         string
	closeChan    chan bool
}

func (s *LoadBalanceEtcdConf) Attach(o Observer) {
	s.observers = append(s.observers, o)
}

func (s *LoadBalanceEtcdConf) NotifyAllObservers() {
	for _, obs := range s.observers {
		obs.Update()
	}
}

func (s *LoadBalanceEtcdConf) GetConf() []string {
	confList := []string{}
	for _, ip := range s.activeList {
		weight, ok := s.confIpWeight[ip]
		if !ok {
			weight = "50" //默认weight
		}
		confList = append(confList, fmt.Sprintf(s.format, ip)+","+weight)
	}
	return confList
}

// WatchConf 更新配置时，通知监听者也更新
func (s *LoadBalanceEtcdConf) WatchConf() {
	etcdManager, err := etcd.NewEtcdManager(s.etcdHosts, 5)
	if err != nil {
		log.Fatalf("etcd.NewEtcdManager error: %v", err)
	}
	fmt.Println("watchConf")
	chanList, chanErr := etcdManager.WatchServerListByPrefix(s.path)
	go func() {
		defer etcdManager.Close()
		for {
			select {
			case changeErr := <-chanErr:
				fmt.Println("changeErr", changeErr)
			case changedList := <-chanList:
				fmt.Println("watch node changed")
				s.UpdateConf(changedList)
			}
		}
	}()
}

// UpdateConf 更新配置时，通知监听者也更新
func (s *LoadBalanceEtcdConf) UpdateConf(conf []string) {
	s.activeList = conf
	for _, obs := range s.observers {
		obs.Update()
	}
}

func NewLoadBalanceZkConf(format, path string, zkHosts []string, conf map[string]string) (*LoadBalanceZkConf, error) {
	zkManager := zookeeper.NewZkManager(zkHosts)
	zkManager.GetConnect()
	defer zkManager.Close()
	zlist, err := zkManager.GetServerListByPath(path)
	if err != nil {
		return nil, err
	}
	mConf := &LoadBalanceZkConf{format: format, activeList: zlist, confIpWeight: conf, zkHosts: zkHosts, path: path, closeChan: make(chan bool, 1)}
	mConf.WatchConf()
	return mConf, nil
}

func (s *LoadBalanceZkConf) CloseWatch() {
	s.closeChan <- true
	close(s.closeChan)
}

type Observer interface {
	Update()
}

type LoadBalanceObserver struct {
	ModuleConf *LoadBalanceZkConf
}

func (l *LoadBalanceObserver) Update() {
	fmt.Println("Update get conf:", l.ModuleConf.GetConf())
}

func NewLoadBalanceObserver(conf *LoadBalanceZkConf) *LoadBalanceObserver {
	return &LoadBalanceObserver{
		ModuleConf: conf,
	}
}
