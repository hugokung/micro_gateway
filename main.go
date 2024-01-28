package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/e421083458/golang_common/lib"
	"github.com/hugokung/micro_gateway/internal/dao"
	"github.com/hugokung/micro_gateway/internal/server"
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
		server.HttpServerRun()

		quit := make(chan os.Signal)
		signal.Notify(quit, syscall.SIGKILL, syscall.SIGQUIT, syscall.SIGINT, syscall.SIGTERM)
		<-quit

		server.HttpServerStop()
	} else {
		lib.InitModule(*config, []string{"base", "mysql", "redis"})
		defer lib.Destroy()

		//加载下游服务的信息到内存
		dao.ServiceManagerHandler.LoadAndWatch()
		//加载租户信息到内存
		dao.AppManagerHandler.LoadAndWatch()
		go func() {
			server.HttpProxyServerRun()
		}()
		go func() {
			server.HttpsProxyServerRun()
		}()
		go func() {
			server.TcpManagerHandler.TcpServerRun()
		}()
		go func() {
			server.GrpcManagerHandler.GrpcServerRun()
		}()
		quit := make(chan os.Signal)
		signal.Notify(quit, syscall.SIGKILL, syscall.SIGQUIT, syscall.SIGINT, syscall.SIGTERM)
		<-quit
		server.TcpManagerHandler.TcpServerStop()
		server.GrpcManagerHandler.GrpcServerStop()
		server.HttpProxyServerStop()
		server.HttpsProxyServerStop()
	}

}
