package http_proxy_middleware

import (
	"fmt"
	"gateway-project/dao"
	"gateway-project/middleware"
	"gateway-project/public"
	"github.com/gin-gonic/gin"
)

// HTTPAccessModeMiddleware 接入模式中间件  把该地址的服务详情 写到请求头当中
func HTTPAccessModeMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		service, err := dao.ServiceManagerHandler.HTTPAccessMode(c)
		if err != nil {
			middleware.ResponseError(c, 1001, err)
			c.Abort()
			return
		}
		fmt.Println("matched service", public.Obj2Json(service))
		c.Set("service", service)
		c.Next()
	}
}
