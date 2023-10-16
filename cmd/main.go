package main

import (
	"fmt"
	"gateway-project/dao"
	"gateway-project/router"
	"github.com/e421083458/golang_common/lib"
	"os"
	"os/signal"
	"syscall"
)

//var (
//	endpoint = flag.String("endpoint", "", "input endpoint dashboard or server")
//)

// endpoint dashboard 后台管理 server 代理服务器
func main() {

	//启动后台管理
	fmt.Printf("welcome")
	dao.ServiceManagerHandler.LoadOnce()
	dao.AppManagerHandler.LoadOnce()
	lib.InitModule("./conf/dev/", []string{"base", "mysql", "redis"})
	defer lib.Destroy()
	router.HttpServerRun()

	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGKILL, syscall.SIGQUIT, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	router.HttpServerStop()

}
