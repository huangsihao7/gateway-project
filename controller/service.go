package controller

import (
	"gateway-project/dao"
	"gateway-project/dto"
	"gateway-project/middleware"
	"github.com/e421083458/golang_common/lib"
	"github.com/gin-gonic/gin"
)

type ServiceController struct {
}

func ServiceRegister(group *gin.RouterGroup) {
	admin := &ServiceController{}
	group.GET("/service_list", admin.ServiceList)

}

// ServiceList godoc
// @Summary 服务列表
// @Description 服务列表
// @Tags 服务管理
// @ID /service/service_list
// @Accept  json
// @Produce  json
// @Param info query string false "关键词"
// @Param pageNum query int true "页数"
// @Param pageSize query int true "每页数量"
// @Success 200 {object} middleware.Response{data=dto.ServiceItemOutput} "success"
// @Router /service/service_list [get]
func (service *ServiceController) ServiceList(c *gin.Context) {

	params := &dto.ServiceListInput{}
	if err := params.BindValidParam(c); err != nil {
		middleware.ResponseError(c, 2000, err)
	}
	serviceInfo := dao.ServiceInfo{}
	tx, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(c, 2001, err)
		return
	}
	list, total, err := serviceInfo.PageList(c, tx, params)
	if err != nil {
		middleware.ResponseError(c, 2002, err)
		return
	}
	var outputLists []dto.ServiceItemOutput
	for _, v := range list {
		outList := dto.ServiceItemOutput{
			ID:          int64(v.Id),
			ServiceName: v.ServiceName,
			ServiceDesc: v.ServiceDesc,
		}
		outputLists = append(outputLists, outList)
	}
	out := dto.ServiceOutput{
		Total: total,
		List:  outputLists,
	}
	middleware.ResponseSuccess(c, out)
}
