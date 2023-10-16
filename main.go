package main

import (
	"flag"
	"gateway-project/dao"
	"gateway-project/grpc_proxy_router"
	"gateway-project/http_proxy_router"
	"gateway-project/tcp_proxy_router"
	"github.com/e421083458/golang_common/lib"
	"os"
	"os/signal"
	"syscall"
)

var (
	endpoint = flag.String("endpoint", "", "input endpoint dashboard or server")
)

// endpoint dashboard 后台管理 server 代理服务器
func main() {

	lib.InitModule("./conf/dev/", []string{"base", "mysql", "redis"})
	defer lib.Destroy()
	dao.ServiceManagerHandler.LoadOnce()
	dao.AppManagerHandler.LoadOnce()

	go func() {
		http_proxy_router.HttpServerRun()
	}()

	go func() {
		http_proxy_router.HttpsServerRun()
	}()

	go func() {
		grpc_proxy_router.GrpcServerRun()
	}()

	go func() {
		tcp_proxy_router.TcpServerRun()
	}()

	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	http_proxy_router.HttpServerStop()
	http_proxy_router.HttpsServerStop()
	grpc_proxy_router.GrpcServerStop()
	tcp_proxy_router.TcpServerStop()
}

//flag.Parse()
//if *endpoint == "" {
//	flag.Usage()
//	os.Exit(1)
//}
//if *endpoint == "dashboard" {
//	//启动后台管理
//	fmt.Printf("welcome")
//	lib.InitModule("./conf/dev/", []string{"base", "mysql", "redis"})
//	defer lib.Destroy()
//	router.HttpServerRun()
//
//	quit := make(chan os.Signal)
//	signal.Notify(quit, syscall.SIGKILL, syscall.SIGQUIT, syscall.SIGINT, syscall.SIGTERM)
//	<-quit
//
//	router.HttpServerStop()
//} else {
//
//	lib.InitModule("./conf/dev/", []string{"base", "mysql", "redis"})
//	defer lib.Destroy()
//	dao.ServiceManagerHandler.LoadOnce()
//
//	go func() {
//		http_proxy_router.HttpServerRun()
//	}()
//
//	go func() {
//		http_proxy_router.HttpsServerRun()
//	}()
//
//	quit := make(chan os.Signal)
//	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
//	<-quit
//
//	http_proxy_router.HttpServerStop()
//	http_proxy_router.HttpsServerStop()
//}
