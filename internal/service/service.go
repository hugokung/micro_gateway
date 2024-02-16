package service

import (
	"log"

	"github.com/Masterminds/semver/v3"
	"github.com/hugokung/micro_gateway/pkg/golang_common/lib"
)

type Service interface {
	Name() string
	Version() *semver.Version
	OnInit() error
	OnStart() error
	OnStop() error
}

type baseService struct{}

func (baseService) Name() string {
	return ""
}

func (baseService) Version() *semver.Version {
	return semver.MustParse("v0.0.1")
}

func (baseService) String() string {
	return ""
}

func newService() (ss []Service) {
	if lib.GetStringConf("base.pprof.addr") == "" {
		return
	}
	ss = append(ss, newPprofileService())
	return
}

func MustInitService() []Service {
	ss := newService()
	for _, s := range ss {
		if err := s.OnInit(); err != nil {
			log.Fatalf("initial %s service error: %s", s.Name(), err)
		}
	}
	return ss
}
