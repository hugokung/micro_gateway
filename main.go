package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/e421083458/golang_common/lib"
	"github.com/hugokung/micro_gateway/dao"
	"github.com/hugokung/micro_gateway/http_proxy_router"
	"github.com/hugokung/micro_gateway/router"
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
		dao.ServiceManagerHandler.LoadOnce()
		go func ()  {
			http_proxy_router.HttpServerRun()
		}()
		go func ()  {
			http_proxy_router.HttpsServerRun()
		}()
		quit := make(chan os.Signal)
		signal.Notify(quit, syscall.SIGKILL, syscall.SIGQUIT, syscall.SIGINT, syscall.SIGTERM)
		<-quit

		http_proxy_router.HttpServerStop()
		http_proxy_router.HttpsServerStop()
	}

}
