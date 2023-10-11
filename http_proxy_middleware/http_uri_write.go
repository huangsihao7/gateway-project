package http_proxy_middleware

import (
	"gateway-project/dao"
	"gateway-project/middleware"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"regexp"
	"strings"
)

// HTTPAccessModeMiddleware 接入模式中间件  把该地址的服务详情 写到请求头当中
func HTTPUrlReWriteMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		serviceInterface, ok := c.Get("service")
		if !ok {
			middleware.ResponseError(c, 2001, errors.New("http proxy middleware service not found"))
			c.Abort()
			return
		}
		serviceDetail := serviceInterface.(*dao.ServiceDetail)

		for _, item := range strings.Split(serviceDetail.HttpRule.UrlRewrite, ",") {
			items := strings.Split(item, " ")
			if len(items) != 2 {
				continue
			}
			regexp, err := regexp.Compile(items[0])
			if err != nil {
				//fmt.Println("regexp.Compile err",err)
				continue
			}
			//fmt.Println("before rewrite", c.Request.URL.Path)
			replacePath := regexp.ReplaceAll([]byte(c.Request.URL.Path), []byte(items[1]))
			c.Request.URL.Path = string(replacePath)
			//fmt.Println("after rewrite", c.Request.URL.Path)
		}
		c.Next()
	}
}
