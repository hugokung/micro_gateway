package service


import (
	"github.com/gin-gonic/gin"
)

type baseHttpService struct {
	baseService

	server *httpServer
}

func (s *baseHttpService) registerRoute(srv Service, h func(e *gin.Engine)) {
	if h != nil {
		h(s.server.e)
	}
	s.server.addService(srv)
}

func (s *baseHttpService) OnStart() error {
	// do nothing default
	return nil
}

func (s *baseHttpService) OnStop() error {
	// do nothing default
	return nil
}