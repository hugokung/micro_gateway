package service

import (
	"net/http"

	"github.com/Masterminds/semver/v3"
	"github.com/hugokung/micro_gateway/pkg/golang_common/lib"
)



type pprofService struct {
	*baseHttpService
}

func (s *pprofService) Name() string {
	return "PprofService"
}

func (s *pprofService) Version() *semver.Version {
	return semver.MustParse("v0.1.0")
}

func (s *pprofService) OnInit() error {
	s.registerRoute(s, nil)
	return nil
}

func newPprofileService() Service {
	addr := lib.GetStringConf("base.pprof.addr")
	server := httpServers.from(addr, func() *httpServer {
		return &httpServer{
			baseServer: newBaseServe(),
			server: &http.Server{
				Addr:    addr,
				Handler: http.DefaultServeMux,
			},
		}
	})
	return &pprofService{
		baseHttpService: &baseHttpService{
			server: server,
		},
	}
}

