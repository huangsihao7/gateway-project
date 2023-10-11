package http_proxy_middleware

import (
	"gateway-project/dao"
	"gateway-project/middleware"
	"gateway-project/public"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

func HTTPFlowCountMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		serviceInterface, ok := c.Get("service")
		if !ok {
			middleware.ResponseError(c, 2001, errors.New("http proxy middleware service not found"))
			c.Abort()
			return
		}
		serviceDetail := serviceInterface.(*dao.ServiceDetail)

		//统计项 1.全站 2.服务 3.租户

		//1.全站
		totalCount, err := public.FlowCounterHandler.GetCounter(public.FlowTotal)
		if err != nil {
			middleware.ResponseError(c, 4001, err)
			c.Abort()
			return
		}
		totalCount.Increase()
		//2.服务
		serviceCount, err := public.FlowCounterHandler.GetCounter(public.FlowServicePrefix + serviceDetail.Info.ServiceName)
		if err != nil {
			middleware.ResponseError(c, 4002, err)
			c.Abort()
			return
		}
		serviceCount.Increase()

		c.Next()
	}
}
