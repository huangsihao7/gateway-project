package http_proxy_middleware

import (
	"gateway-project/dao"
	"gateway-project/middleware"
	"gateway-project/reverse_proxy"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

func HTTPReverseProxyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		serviceInterface, ok := c.Get("service")
		if !ok {
			middleware.ResponseError(c, 2001, errors.New("http proxy middleware service not found"))
			c.Abort()
			return
		}
		//类型转换
		serviceDetail := serviceInterface.(*dao.ServiceDetail)
		lb, err := dao.LoadBalancerHandler.GetLoadBalancer(serviceDetail)
		if err != nil {
			middleware.ResponseError(c, 2002, err)
			c.Abort()
			return
		}
		trans, err := dao.TransportorHandler.GetTrans(serviceDetail)
		if err != nil {
			middleware.ResponseError(c, 2003, err)
			c.Abort()
			return
		}
		//创建 reverse proxy
		proxy := reverse_proxy.NewLoadBalanceReverseProxy(c, lb, trans)
		proxy.ServeHTTP(c.Writer, c.Request)
		c.Next()
	}
}
