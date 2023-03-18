package http_proxy_router

import (
	"github.com/gin-gonic/gin"
	"github.com/hugokung/micro_gateway/http_proxy_middleware"
)

func InitRouter(middlewares ...gin.HandlerFunc) *gin.Engine {

	router := gin.Default()
	router.Use(middlewares...)
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	
	router.Use(
	http_proxy_middleware.HTTPAccessModeMiddleware(), 
	http_proxy_middleware.HTTPFlowCountMiddleware(),
	http_proxy_middleware.HTTPFlowLimitMiddleware(),
	http_proxy_middleware.HTTPWhileListMiddleware(),
	http_proxy_middleware.HTTPBlacListMiddleware(),
	http_proxy_middleware.HTTPHeaderTransferMiddleware(), 
	http_proxy_middleware.HTTPStripUriMiddleware(),
	http_proxy_middleware.HTTPUrlRewriteMiddleware(),
	http_proxy_middleware.HTTPReverseProxyMiddleware())
	return router
}
