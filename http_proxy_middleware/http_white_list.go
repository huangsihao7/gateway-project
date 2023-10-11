package http_proxy_middleware

import (
	"fmt"
	"gateway-project/dao"
	"gateway-project/middleware"
	"gateway-project/public"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"strings"
)

//检查输入的IP是否存在白名单里 存在放行 不存在不放行

func HTTPWhiteListMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		serviceInterface, ok := c.Get("service")
		if !ok {
			middleware.ResponseError(c, 2001, errors.New("http proxy middleware service not found"))
			c.Abort()
			return
		}
		serviceDetail := serviceInterface.(*dao.ServiceDetail)
		iplist := []string{}
		if serviceDetail.AccessControl.WhiteList != "" {
			iplist = strings.Split(serviceDetail.AccessControl.WhiteList, ",")
		}
		if serviceDetail.AccessControl.OpenAuth == 1 && len(iplist) > 0 {
			if !public.InStringSlice(iplist, c.ClientIP()) {
				middleware.ResponseError(c, 3001, errors.New(fmt.Sprintf("%s not in white ip list", c.ClientIP())))
				c.Abort()
				return
			}
		}
		c.Next()
	}
}
