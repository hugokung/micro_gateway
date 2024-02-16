package service

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)

var (
	_ server = (*httpServer)(nil)
)

// httpServer wraper for gin.engine and http.Server
type httpServer struct {
	*baseServer

	e      *gin.Engine
	server *http.Server
}

func (s *httpServer) start() error {
	return s.server.ListenAndServe()
}

func (s *httpServer) stop() error {
	return s.server.Shutdown(context.Background())
}