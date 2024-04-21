package router

import (
	"github.com/gin-gonic/gin"
	"github.com/hugokung/micro_gateway/api"
	"github.com/hugokung/micro_gateway/internal/middleware"
	"github.com/hugokung/micro_gateway/internal/middleware/http_proxy_middleware"
)

func InitProxyRouter(middlewares ...gin.HandlerFunc) *gin.Engine {

	gin.SetMode("release")
	router := gin.Default()
	router.Use(middlewares...)
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	oauth := router.Group("/oauth")
	oauth.Use(middleware.TranslationMiddleware())
	{
		api.OAuthRegister(oauth)
	}
	router.Use(
		http_proxy_middleware.HTTPAccessModeMiddleware(),
		http_proxy_middleware.HTTPFlowCountMiddleware(),
		http_proxy_middleware.HTTPFlowLimitMiddleware(),
		//http_proxy_middleware.HTTPCircuitBreakMiddleware(),
		http_proxy_middleware.HTTPJwtAuthTokenMiddleware(),
		http_proxy_middleware.HTTPJwtFlowCountMiddleware(),
		http_proxy_middleware.HTTPJwtFlowLimitMiddleware(),
		http_proxy_middleware.HTTPWhileListMiddleware(),
		http_proxy_middleware.HTTPBlacListMiddleware(),
		http_proxy_middleware.HTTPHeaderTransferMiddleware(),
		http_proxy_middleware.HTTPStripUriMiddleware(),
		http_proxy_middleware.HTTPUrlRewriteMiddleware(),
		http_proxy_middleware.HTTPReverseProxyMiddleware())
	return router
}
