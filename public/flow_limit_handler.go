package public

import (
	"sync"

	"golang.org/x/time/rate"
)

var FlowLimiterHandler *FlowLimiter

type FlowLimiter struct {
	 FlowLimiertMap map[string]*FlowLimiterItem
	 FlowLimiterSlice []*FlowLimiterItem
	 Locker sync.RWMutex
}

type FlowLimiterItem struct {
	ServiceName string
	Limiter *rate.Limiter
}

func NewFlowLimiter() *FlowLimiter {
	return &FlowLimiter{
		FlowLimiertMap: map[string]*FlowLimiterItem{},
		FlowLimiterSlice: []*FlowLimiterItem{},
		Locker: sync.RWMutex{},
	}
}

func init() {
	FlowLimiterHandler = NewFlowLimiter()
}

func (limiter *FlowLimiter) GetLimiter(serviceName string, qps float64) (*rate.Limiter, error) {
	for _, item := range limiter.FlowLimiterSlice {
		if item.ServiceName == serviceName {
			return item.Limiter, nil
		}
	}

	newLimiter := rate.NewLimiter(rate.Limit(qps), int(qps) * 3)
	
	item := &FlowLimiterItem{
		ServiceName: serviceName,
		Limiter: newLimiter,
	}

	limiter.FlowLimiterSlice = append(limiter.FlowLimiterSlice, item)
	limiter.Locker.Lock()
	defer limiter.Locker.Unlock()
	limiter.FlowLimiertMap[serviceName] = item
	return newLimiter, nil
}