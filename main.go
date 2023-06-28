package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/e421083458/golang_common/lib"
	"github.com/hugokung/micro_gateway/dao"
	"github.com/hugokung/micro_gateway/grpc_proxy_router"
	"github.com/hugokung/micro_gateway/http_proxy_router"
	"github.com/hugokung/micro_gateway/router"
	"github.com/hugokung/micro_gateway/tcp_proxy_router"
)

var (
	endpoint = flag.String("endpoint", "", "input endpoint dashboard or server")
	config   = flag.String("conf", "", "input config file like ./conf/dev/")
)

func main() {

	flag.Parse()
	if *endpoint == "" {
		flag.Usage()
		os.Exit(1)
	}
	if *config == "" {
		flag.Usage()
		os.Exit(1)
	}

	if *endpoint == "dashboard" {
		lib.InitModule(*config, []string{"base", "mysql", "redis"})
		defer lib.Destroy()
		router.HttpServerRun()

		quit := make(chan os.Signal)
		signal.Notify(quit, syscall.SIGKILL, syscall.SIGQUIT, syscall.SIGINT, syscall.SIGTERM)
		<-quit

		router.HttpServerStop()
	} else {
		lib.InitModule(*config, []string{"base", "mysql", "redis"})
		defer lib.Destroy()

		//加载下游服务的信息到内存
		dao.ServiceManagerHandler.LoadAndWatch()
		//加载租户信息到内存
		dao.AppManagerHandler.LoadAndWatch()
		go func() {
			http_proxy_router.HttpServerRun()
		}()
		go func() {
			http_proxy_router.HttpsServerRun()
		}()
		go func() {
			tcp_proxy_router.TcpManagerHandler.TcpServerRun()
		}()
		go func() {
			grpc_proxy_router.GrpcManagerHandler.GrpcServerRun()
		}()
		quit := make(chan os.Signal)
		signal.Notify(quit, syscall.SIGKILL, syscall.SIGQUIT, syscall.SIGINT, syscall.SIGTERM)
		<-quit
		tcp_proxy_router.TcpManagerHandler.TcpServerStop()
		grpc_proxy_router.GrpcManagerHandler.GrpcServerStop()
		http_proxy_router.HttpServerStop()
		http_proxy_router.HttpsServerStop()
	}

}
