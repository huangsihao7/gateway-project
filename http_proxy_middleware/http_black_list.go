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

// 在白名单为空的情况下 黑名单才会生效

func HTTPBlackListMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		serviceInterface, ok := c.Get("service")
		if !ok {
			middleware.ResponseError(c, 2001, errors.New("http proxy middleware service not found"))
			c.Abort()
			return
		}
		serviceDetail := serviceInterface.(*dao.ServiceDetail)
		whiteList := []string{}
		if serviceDetail.AccessControl.WhiteList != "" {
			whiteList = strings.Split(serviceDetail.AccessControl.WhiteList, ",")
		}
		blackList := []string{}
		if serviceDetail.AccessControl.BlackList != "" {
			blackList = strings.Split(serviceDetail.AccessControl.BlackList, ",")
		}
		if serviceDetail.AccessControl.OpenAuth == 1 && len(whiteList) == 0 && len(blackList) > 0 {
			if public.InStringSlice(blackList, c.ClientIP()) {
				middleware.ResponseError(c, 3001, errors.New(fmt.Sprintf("%s in black ip list", c.ClientIP())))
				c.Abort()
				return
			}
		}
		c.Next()
	}
}
