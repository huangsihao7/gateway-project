package http_proxy_middleware

import (
	"gateway-project/dao"
	"gateway-project/middleware"
	"gateway-project/public"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"strings"
)

// HTTPAccessModeMiddleware 接入模式中间件  把该地址的服务详情 写到请求头当中
func HTTPStripUriMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		serviceInterface, ok := c.Get("service")
		if !ok {
			middleware.ResponseError(c, 2001, errors.New("http proxy middleware service not found"))
			c.Abort()
			return
		}
		serviceDetail := serviceInterface.(*dao.ServiceDetail)

		//判断是不是前缀url
		if serviceDetail.HttpRule.RuleType == public.HTTPRuleTypePrefixURL && serviceDetail.HttpRule.NeedStripUri == 1 {
			//fmt.Printf(c.Request.URL.Path)
			//fmt.Printf(serviceDetail.HttpRule.Rule)
			c.Request.URL.Path = strings.Replace(c.Request.URL.Path, serviceDetail.HttpRule.Rule, "", 1)
		}
		c.Next()
	}
}
