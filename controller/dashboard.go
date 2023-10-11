package controller

import (
	"gateway-project/dao"
	"gateway-project/dto"
	"gateway-project/middleware"
	"gateway-project/public"
	"github.com/e421083458/golang_common/lib"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

type DashboardController struct {
}

func DashboardRegister(group *gin.RouterGroup) {
	dashboard := &DashboardController{}
	group.GET("/panelGroupData", dashboard.PanelGroupData)
	group.GET("/flowStat", dashboard.FlowStat)
	group.GET("/serviceStat", dashboard.ServiceStat)
}

// PanelGroupData godoc
// @Summary 指标统计
// @Description 指标统计
// @Tags 首页大盘
// @ID /main/panelGroupData
// @Accept  json
// @Produce  json
// @Success 200 {object} middleware.Response{data=dto.PanelGroupOutput} "success"
// @Router /main/panelGroupData [get]
func (dashboardController *DashboardController) PanelGroupData(c *gin.Context) {
	// 数据库查询
	tx, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(c, 2000, err) // 去取数据库连接失败
		return
	}

	//统计服务数
	serviceInfo := &dao.ServiceInfo{}

	_, serviceTotal, err := serviceInfo.PageList(c, tx, &dto.ServiceListInput{PageSize: 1, PageNum: 1})
	if err != nil {
		middleware.ResponseError(c, 2001, err) // 去取数据库连接失败
		return
	}
	//统计租户数
	appInfo := &dao.App{}
	_, appTotal, err := appInfo.APPList(c, tx, &dto.APPListInput{PageSize: 1, PageNo: 1})
	if err != nil {
		middleware.ResponseError(c, 2002, err) // 去取数据库连接失败
		return
	}
	//统计总体 的当日请求量 统计QPS query per second
	counter, err := public.FlowCounterHandler.GetCounter(public.FlowTotal)
	if err != nil {
		middleware.ResponseError(c, 2003, err)
		return
	}

	out := &dto.PanelGroupOutput{
		ServiceNum:      serviceTotal,
		AppNum:          appTotal,
		CurrentNum:      counter.TotalCount,
		TodayRequestNum: counter.QPS,
	}
	middleware.ResponseSuccess(c, out)
}

// FlowStat godoc
// @Summary 访问统计
// @Description 访问统计
// @Tags 首页大盘
// @ID /main/flowStat
// @Accept  json
// @Produce  json
// @Success 200 {object} middleware.Response{data=dto.ServiceStatOutput} "success"
// @Router /main/flowStat [get]
func (dashboardController *DashboardController) FlowStat(c *gin.Context) {
	var todayList []int64
	var yesterdayList []int64

	for i := 0; i < 23; i++ {
		todayList = append(todayList, 0)
	}

	for i := 0; i < 23; i++ {
		yesterdayList = append(yesterdayList, 0)
	}

	middleware.ResponseSuccess(c, &dto.ServiceStatOutput{
		Today:     todayList,
		Yesterday: yesterdayList,
	})

}

// ServiceStat godoc
// @Summary 服务统计
// @Description 服务统计
// @Tags 首页大盘
// @ID /main/serviceStat
// @Accept  json
// @Produce  json
// @Success 200 {object} middleware.Response{data=dto.ServiceMainStatOutput} "success"
// @Router /main/serviceStat [get]
func (dashboardController *DashboardController) ServiceStat(c *gin.Context) {

	// 数据库查询
	tx, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(c, 2000, err) // 去取数据库连接失败
		return
	}
	serviceInfo := &dao.ServiceInfo{}
	list, err := serviceInfo.GroupByLoadType(c, tx)
	if err != nil {
		middleware.ResponseError(c, 2001, err) // 去取数据库连接失败
		return
	}

	var legend []string

	for i, v := range list {
		name, ok := public.LoadTypeMap[v.LoadType]
		if !ok {
			middleware.ResponseError(c, 2002, errors.New("LoadType 不存在类型")) // 不存在类型
		}
		list[i].Name = name
		legend = append(legend, name)
	}
	out := &dto.ServiceMainStatOutput{
		Legend: legend,
		Data:   list,
	}

	middleware.ResponseSuccess(c, out)
}
