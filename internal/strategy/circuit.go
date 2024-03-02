package strategy

import (
	"log"
	"net/http/httptest"

	"github.com/afex/hystrix-go/hystrix"
	"github.com/gin-gonic/gin"
	"github.com/hugokung/micro_gateway/internal/dao"
	"github.com/hugokung/micro_gateway/pkg/golang_common/lib"
)

func InitCircuitConfig() {
	circuit := dao.CircuitConfig{}
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("GET", "/", nil)
	tx, err := lib.GetGormPool("default")
	if err != nil {
		log.Fatalf("InitCircuitConfig err: %v", err)
	}
	list, err := circuit.FindAll(c, tx)
	if err == nil {
		for i := range list {
			hystrix.ConfigureCommand(list[i].ServiceName, hystrix.CommandConfig{
				Timeout: list[i].Timeout,
				MaxConcurrentRequests: list[i].MaxConcurrentRequests,
				RequestVolumeThreshold: list[i].RequestVolumeThreshold,
				SleepWindow: list[i].SleepWindow,
				ErrorPercentThreshold: list[i].ErrorPercentThreshold,
			})
		}
	}
}

func UpdateCircuitConfig(serviceName string, circuitConfig *dao.CircuitConfig) {
	hystrix.ConfigureCommand(serviceName, hystrix.CommandConfig{
		Timeout: circuitConfig.Timeout,
		MaxConcurrentRequests: circuitConfig.MaxConcurrentRequests,
		RequestVolumeThreshold: circuitConfig.RequestVolumeThreshold,
		SleepWindow: circuitConfig.SleepWindow,
		ErrorPercentThreshold: circuitConfig.ErrorPercentThreshold,
	})
}