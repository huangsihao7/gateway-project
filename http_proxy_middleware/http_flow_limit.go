package http_proxy_middleware

import (
	"fmt"
	"gateway-project/dao"
	"gateway-project/middleware"
	"gateway-project/public"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

func HTTPFlowLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		serviceInterface, ok := c.Get("service")
		if !ok {
			middleware.ResponseError(c, 2001, errors.New("http proxy middleware service not found"))
			c.Abort()
			return
		}
		serviceDetail := serviceInterface.(*dao.ServiceDetail)

		if serviceDetail.AccessControl.ServiceFlowLimit > 0 {
			serviceLimiter, err := public.FlowLimiterHandler.GetLimiter(public.FlowServicePrefix+serviceDetail.Info.ServiceName, float64(serviceDetail.AccessControl.ServiceFlowLimit))
			if err != nil {
				middleware.ResponseError(c, 5001, err)
				c.Abort()
				return
			}
			if !serviceLimiter.Allow() {
				middleware.ResponseError(c, 5002, errors.New(fmt.Sprintf("service flow limit %v", serviceDetail.AccessControl.ServiceFlowLimit)))
				c.Abort()
				return
			}
		}

		if serviceDetail.AccessControl.ClientIPFlowLimit > 0 {
			clientLimiter, err := public.FlowLimiterHandler.GetLimiter(
				public.FlowServicePrefix+serviceDetail.Info.ServiceName+"_"+c.ClientIP(),
				float64(serviceDetail.AccessControl.ClientIPFlowLimit))
			if err != nil {
				middleware.ResponseError(c, 5003, err)
				c.Abort()
				return
			}
			if !clientLimiter.Allow() {
				middleware.ResponseError(c, 5002, errors.New(fmt.Sprintf("%v flow limit %v", c.ClientIP(), serviceDetail.AccessControl.ClientIPFlowLimit)))
				c.Abort()
				return
			}
		}
		c.Next()
	}
}
