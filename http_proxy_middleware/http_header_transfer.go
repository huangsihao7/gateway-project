package http_proxy_middleware

import (
	"gateway-project/dao"
	"gateway-project/middleware"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"strings"
)

// HTTPHeaderTransferMiddleware  在代理的过程中修改请求头
func HTTPHeaderTransferMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		serviceInterface, ok := c.Get("service")
		if !ok {
			middleware.ResponseError(c, 2001, errors.New("http proxy middleware service not found"))
			c.Abort()
			return
		}
		serviceDetail := serviceInterface.(*dao.ServiceDetail)

		for _, item := range strings.Split(serviceDetail.HttpRule.HeaderTransfor, ",") {
			items := strings.Split(item, " ")

			if len(items) != 3 {
				continue
			}
			if items[0] == "add" || items[0] == "edit" {
				c.Request.Header.Set(items[1], items[2])
			}
			if items[0] == "del" {
				c.Request.Header.Del(items[1])
			}
		}

		c.Next()
	}
}
