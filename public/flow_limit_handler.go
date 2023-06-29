package public

import (
	"math"
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
	Qps float64
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
	isUpdate := false
	var newLimiter *rate.Limiter
	for i, item := range limiter.FlowLimiterSlice {
		if item.ServiceName == serviceName {
			if math.Abs(item.Qps - qps) <= 1e-4 {
				return item.Limiter, nil
			} else {
				isUpdate = true
				limiter.FlowLimiterSlice[i].Qps = qps
				newLimiter = rate.NewLimiter(rate.Limit(qps), int(qps) * 3)
				limiter.FlowLimiterSlice[i].Limiter = newLimiter
				break
			}
		}
	}

	if !isUpdate {
		newLimiter = rate.NewLimiter(rate.Limit(qps), int(qps) * 3)
		item := &FlowLimiterItem{
			ServiceName: serviceName,
			Limiter: newLimiter,
			Qps: qps,
		}
		limiter.FlowLimiterSlice = append(limiter.FlowLimiterSlice, item)
		limiter.Locker.Lock()
		defer limiter.Locker.Unlock()
		limiter.FlowLimiertMap[serviceName] = item

	} else {
		limiter.Locker.Lock()
		defer limiter.Locker.Unlock()
		limiter.FlowLimiertMap[serviceName].Qps = qps
		limiter.FlowLimiertMap[serviceName].Limiter = newLimiter
	}
	return newLimiter, nil
}