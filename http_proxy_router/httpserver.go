package http_proxy_router

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/e421083458/golang_common/lib"
	"github.com/gin-gonic/gin"
	"github.com/hugokung/micro_gateway/cert_file"
	"github.com/hugokung/micro_gateway/middleware"
)

var (
	HttpSrvHandler *http.Server
	HttpsSrvHandler *http.Server
)

func HttpServerRun() {
	gin.SetMode(lib.ConfBase.DebugMode)
	r := InitRouter(middleware.RecoveryMiddleware(),middleware.RequestLog())
	HttpSrvHandler = &http.Server{
		Addr:           lib.GetStringConf("proxy.http.addr"),
		Handler:        r,
		ReadTimeout:    time.Duration(lib.GetIntConf("proxy.http.read_timeout")) * time.Second,
		WriteTimeout:   time.Duration(lib.GetIntConf("proxy.http.write_timeout")) * time.Second,
		MaxHeaderBytes: 1 << uint(lib.GetIntConf("proxy.http.max_header_bytes")),
	}
	log.Printf(" [INFO] Http_Proxy_Run:%s\n",lib.GetStringConf("proxy.http.addr"))
	if err := HttpSrvHandler.ListenAndServe(); err != nil {
		log.Fatalf(" [ERROR] Http_Proxy_Run:%s err:%v\n", lib.GetStringConf("proxy.http.addr"), err)
	}
}

func HttpsServerRun() {
	gin.SetMode(lib.ConfBase.DebugMode)
	r := InitRouter(middleware.RecoveryMiddleware(),middleware.RequestLog())
	HttpsSrvHandler = &http.Server{
		Addr:           lib.GetStringConf("proxy.https.addr"),
		Handler:        r,
		ReadTimeout:    time.Duration(lib.GetIntConf("proxy.https.read_timeout")) * time.Second,
		WriteTimeout:   time.Duration(lib.GetIntConf("proxy.https.write_timeout")) * time.Second,
		MaxHeaderBytes: 1 << uint(lib.GetIntConf("proxy.https.max_header_bytes")),
	}
	log.Printf(" [INFO] Https_Proxy_Run:%s\n",lib.GetStringConf("proxy.https.addr"))
	if err := HttpsSrvHandler.ListenAndServeTLS(cert_file.Path("server.crt"), cert_file.Path("server.key")); err != nil {
		log.Fatalf(" [ERROR] Https_Proxy_Run:%s err:%v\n", lib.GetStringConf("proxy.https.addr"), err)
	}
}

func HttpServerStop() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := HttpSrvHandler.Shutdown(ctx); err != nil {
		log.Fatalf(" [ERROR] Http_Proxy_Stop err:%v\n", err)
	}
	log.Printf(" [INFO] Http_Proxy stopped\n")
}

func HttpsServerStop() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := HttpsSrvHandler.Shutdown(ctx); err != nil {
		log.Fatalf(" [ERROR] Https_Proxy_Stop err:%v\n", err)
	}
	log.Printf(" [INFO] Https_Proxy stopped\n")
}