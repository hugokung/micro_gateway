package dao

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CircuitConfig struct {
	ID						int64 	`json:"id" gorm:"primary_key"`
	ServiceID				int64 	`json:"service_id" gorm:"column:service_id" description:"服务id"`
	ServiceName				string	`json:"service_name" gorm:"column:service_name"`
	Timeout                	int 	`json:"timeout" gorm:"column:timeout"`
	MaxConcurrentRequests  	int 	`json:"max_concurrent_requests" gorm:"column:max_concurrent_requests"`
	RequestVolumeThreshold 	int 	`json:"request_volume_threshold" gorm:"column:request_volume_threshold"`
	SleepWindow            	int 	`json:"sleep_window" gorm:"column:sleep_window"`
	ErrorPercentThreshold  	int 	`json:"error_percent_threshold" gorm:"column:error_percent_threshold"`
	FallBackMsg				string	`json:"fall_back_msg" gorm:"column:fall_back_msg"`
	NeedCircuit				int		`json:"need_circuit" gorm:"column:need_circuit"` //0: close, 1: open
}

func (c *CircuitConfig) TableName() string {
	return "gateway_service_circuit_config"
}

func (ci *CircuitConfig) Find(c *gin.Context, tx *gorm.DB, search *CircuitConfig) (*CircuitConfig, error) {
	model := &CircuitConfig{}
	err := tx.WithContext(c).Where(search).Find(model).Error
	if err != nil {
		return nil, err
	}
	return model, nil
}

func (ci *CircuitConfig) Save(c *gin.Context, tx *gorm.DB) error {
	if err := tx.WithContext(c).Save(ci).Error; err != nil {
		return err
	}
	return nil
}

func (ci *CircuitConfig) FindAll(c *gin.Context, tx *gorm.DB) ([]CircuitConfig, error) {
	var res []CircuitConfig
	query := tx.WithContext(c)
	query  = query.Table(ci.TableName()).Select("*").Where("need_ciruit=?", 1)
	err := query.Find(&res).Error
	if err != nil {
		return nil, err
	}
	return res, nil

}