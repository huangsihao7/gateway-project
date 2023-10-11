package controller

import (
	"fmt"
	"gateway-project/dao"
	"gateway-project/dto"
	"gateway-project/middleware"
	"gateway-project/public"
	"github.com/e421083458/golang_common/lib"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"strings"
	"time"
)

type ServiceController struct {
}

func ServiceRegister(group *gin.RouterGroup) {
	serviceRegister := &ServiceController{}
	group.GET("/service_list", serviceRegister.ServiceList)
	group.GET("/service_delete", serviceRegister.ServiceDelete)
	group.GET("/serviceDetail", serviceRegister.serviceDetail)
	group.GET("/service_stat", serviceRegister.ServiceStat)
	group.POST("/serviceAddHttp", serviceRegister.serviceAddHttp)
	group.POST("/serviceUpdateHttp", serviceRegister.serviceUpdateHttp)

	group.POST("/service_add_tcp", serviceRegister.ServiceAddTcp)
	group.POST("/service_update_tcp", serviceRegister.ServiceUpdateTcp)
	group.POST("/service_add_grpc", serviceRegister.ServiceAddGrpc)
	group.POST("/service_update_grpc", serviceRegister.ServiceUpdateGrpc)

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
func (serviceController *ServiceController) ServiceList(c *gin.Context) {

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
		serviceDetail, err := v.ServiceDetail(c, tx, &v)
		if err != nil {
			middleware.ResponseError(c, 2003, err)
			return
		}

		serviceAddr := "unknown"

		//获取参数
		clusterIp := lib.GetStringConf("base.cluster.cluster_ip")
		clusterPort := lib.GetStringConf("base.cluster.cluster_port")
		clusterSSLPort := lib.GetStringConf("base.cluster.cluster_ssl_port")

		//针对HTTP 分为HTTPS HTTP　　domain prefix
		if serviceDetail.Info.LoadType == public.LoadTypeHTTP && serviceDetail.HttpRule.NeedHttps == 1 {
			if serviceDetail.HttpRule.RuleType == public.HTTPRuleTypePrefixURL {
				serviceAddr = clusterIp + ":" + clusterSSLPort + serviceDetail.HttpRule.Rule
			}
			if serviceDetail.HttpRule.RuleType == public.HTTPRuleTypeDomain {
				serviceAddr = serviceDetail.HttpRule.Rule
			}
		} else if serviceDetail.Info.LoadType == public.LoadTypeHTTP {
			if serviceDetail.HttpRule.RuleType == public.HTTPRuleTypePrefixURL {
				serviceAddr = clusterIp + ":" + clusterPort + serviceDetail.HttpRule.Rule
			}
			if serviceDetail.HttpRule.RuleType == public.HTTPRuleTypeDomain {
				serviceAddr = serviceDetail.HttpRule.Rule
			}
		}
		//针对TCP
		if serviceDetail.Info.LoadType == public.LoadTypeTCP {
			serviceAddr = fmt.Sprintf("%s:%d", clusterIp, serviceDetail.TcpRule.Port)
		}
		//针对GRPC
		if serviceDetail.Info.LoadType == public.LoadTypeGRPC {
			serviceAddr = fmt.Sprintf("%s:%d", clusterIp, serviceDetail.GrpcRule.Port)
		}
		// 拿到IPlist
		ipList := serviceDetail.LoadBalance.GetIPListByModel()
		outList := dto.ServiceItemOutput{
			ID:          int64(v.Id),
			ServiceName: v.ServiceName,
			LoadType:    v.LoadType,
			ServiceDesc: v.ServiceDesc,
			ServiceAddr: serviceAddr,
			Qps:         0,
			Qpd:         0,
			TotalNode:   len(ipList),
		}
		outputLists = append(outputLists, outList)
	}
	out := dto.ServiceOutput{
		Total: total,
		List:  outputLists,
	}
	middleware.ResponseSuccess(c, out)
}

// ServiceDelete godoc
// @Summary 服务删除
// @Description 服务删除
// @Tags 服务管理
// @ID /service/service_delete
// @Accept  json
// @Produce  json
// @Param id query string true "服务ID"
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /service/service_delete [get]
func (serviceController *ServiceController) ServiceDelete(c *gin.Context) {
	params := &dto.ServiceDeleteInput{}
	if err := params.BindValidParam(c); err != nil {
		middleware.ResponseError(c, 2000, err)
	}

	//获取数据库连接
	tx, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(c, 2001, err)
		return
	}
	serviceInfo := &dao.ServiceInfo{Id: params.Id}
	serviceInfo, err = serviceInfo.Find(c, tx, serviceInfo)
	if err != nil {
		middleware.ResponseError(c, 2002, err)
		return
	}
	serviceInfo.IsDelete = 1
	if err = serviceInfo.Save(c, tx); err != nil {
		middleware.ResponseError(c, 2003, err)
		return
	}
	//返回结果
	middleware.ResponseSuccess(c, "删除成功")
}

// ServiceAddHTTP godoc
// @Summary 添加HTTP服务
// @Description 添加HTTP服务
// @Tags 服务管理
// @ID /service/service_add_http
// @Accept  json
// @Produce  json
// @Param body body dto.ServiceAddHTTPInput true "body"
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /service/serviceAddHttp [post]
func (serviceController *ServiceController) serviceAddHttp(c *gin.Context) {

	params := &dto.ServiceAddHTTPInput{}

	if err := params.BindValidParam(c); err != nil {
		middleware.ResponseError(c, 2000, err) // 添加HTTP参数不正确
		return
	}

	//获取数据库连接
	tx, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(c, 2001, err)
		return
	}
	tx = tx.Begin()
	//验证服务名称是否存在
	serviceInfo := &dao.ServiceInfo{ServiceName: params.ServiceName}
	if _, err := serviceInfo.Find(c, tx, serviceInfo); err == nil {
		tx.Rollback()
		middleware.ResponseError(c, 2002, errors.New("服务名称已存在")) // 添加HTTP参数不正确
		return
	}

	//验证接入规则是否重复
	httpURL := &dao.HttpRule{RuleType: params.RuleType, Rule: params.Rule}
	if _, err = httpURL.Find(c, tx, httpURL); err == nil {
		tx.Rollback()
		middleware.ResponseError(c, 2003, errors.New("服务接入前缀域名或域名已存在")) // 添加HTTP参数不正确
		return
	}

	//验证IPList和权重列表长度是否相同
	if len(strings.Split(params.IpList, "\n")) != len(strings.Split(params.WeightList, "\n")) {
		tx.Rollback()
		middleware.ResponseError(c, 2004, errors.New("IP列表与权重列表数量不一致")) // 添加HTTP参数不正确
		return
	}

	//存入数据库
	serviceModel := &dao.ServiceInfo{
		ServiceName: params.ServiceName,
		ServiceDesc: params.ServiceDesc,
	}
	if err = serviceModel.Save(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 2005, err) // 添加HTTP参数不正确
		return
	}

	httpRule := &dao.HttpRule{
		ServiceID:      serviceModel.Id,
		RuleType:       params.RuleType,
		Rule:           params.Rule,
		NeedHttps:      params.NeedHttps,
		NeedStripUri:   params.NeedStripUri,
		NeedWebsocket:  params.NeedWebsocket,
		UrlRewrite:     params.UrlRewrite,
		HeaderTransfor: params.HeaderTransfor,
	}
	if err = httpRule.Save(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 2006, err) // 添加HTTP参数不正确
		return
	}

	accessControl := &dao.AccessControl{
		ServiceID:         serviceModel.Id,
		OpenAuth:          params.OpenAuth,
		BlackList:         params.BlackList,
		WhiteList:         params.WhiteList,
		ClientIPFlowLimit: params.ClientipFlowLimit,
		ServiceFlowLimit:  params.ServiceFlowLimit,
	}
	if err = accessControl.Save(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 2007, err) // 添加HTTP参数不正确
		return
	}

	loadBalance := &dao.LoadBalance{
		ServiceID:              serviceModel.Id,
		RoundType:              params.RoundType,
		IpList:                 params.IpList,
		WeightList:             params.WeightList,
		UpstreamConnectTimeout: params.UpstreamConnectTimeout,
		UpstreamHeaderTimeout:  params.UpstreamHeaderTimeout,
		UpstreamIdleTimeout:    params.UpstreamIdleTimeout,
		UpstreamMaxIdle:        params.UpstreamMaxIdle,
	}
	if err = loadBalance.Save(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 2008, err) // 添加HTTP参数不正确
		return
	}

	tx.Commit()
	//返回结果
	middleware.ResponseSuccess(c, "添加http代理成功")
}

// serviceUpdateHttp godoc
// @Summary 更新HTTP服务
// @Description 更新HTTP服务
// @Tags 服务管理
// @ID /service/serviceUpdateHttp
// @Accept  json
// @Produce  json
// @Param body body dto.ServiceUpdateHTTPInput true "body"
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /service/serviceUpdateHttp [post]
func (serviceController *ServiceController) serviceUpdateHttp(c *gin.Context) {
	params := &dto.ServiceUpdateHTTPInput{}

	if err := params.BindValidParam(c); err != nil {
		middleware.ResponseError(c, 2000, err) // 添加HTTP参数不正确
		return
	}

	//ip与权重数量一致
	if len(strings.Split(params.IpList, ",")) != len(strings.Split(params.WeightList, ",")) {
		middleware.ResponseError(c, 2001, errors.New("ip列表与权重设置不匹配"))
		return
	}

	//获取数据库连接
	tx, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(c, 2001, err)
		return
	}
	tx = tx.Begin()

	//拿到基本信息
	serviceInfo := &dao.ServiceInfo{Id: params.ID}
	serviceInfo, err = serviceInfo.Find(c, tx, serviceInfo)
	if err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 2002, errors.New("基本信息不存在")) // 基本信息不存在
		return
	}
	//拿到总的信息
	//查看服务详情是否存在
	serviceDetail, err := serviceInfo.ServiceDetail(c, tx, serviceInfo)
	if err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 2003, errors.New("服务不存在")) // 更新的服务并不存在
		return
	}
	// 校验服务类型
	if serviceDetail.Info.LoadType != public.LoadTypeHTTP {
		middleware.ResponseError(c, 2004, errors.New("类型错误")) // http不存在该服务
		return
	}

	//更新各个表的信息
	info := serviceDetail.Info
	info.ServiceDesc = params.ServiceDesc
	err = info.Save(c, tx)
	if err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 2005, errors.New("更新Info表格失败"))
	}
	//更新http
	httpRule := serviceDetail.HttpRule
	httpRule.NeedHttps = params.NeedHttps
	httpRule.NeedStripUri = params.NeedStripUri
	httpRule.NeedWebsocket = params.NeedWebsocket
	httpRule.UrlRewrite = params.UrlRewrite
	httpRule.HeaderTransfor = params.HeaderTransfor
	err = httpRule.Save(c, tx)
	if err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 2006, errors.New("更新HTTP表格失败")) // 更新HTTP表格失败
		return
	}

	// 更新权限表格
	accessControl := serviceDetail.AccessControl
	accessControl.OpenAuth = params.OpenAuth
	accessControl.BlackList = params.BlackList
	accessControl.WhiteList = params.WhiteList
	accessControl.ClientIPFlowLimit = params.ClientipFlowLimit
	accessControl.ServiceFlowLimit = params.ServiceFlowLimit
	err = accessControl.Save(c, tx)
	if err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 2018, errors.New("更新权限表格失败")) // 更新权限表格失败
		return
	}
	// 更新负载均衡
	loadBalance := serviceDetail.LoadBalance
	loadBalance.RoundType = params.RoundType
	loadBalance.IpList = params.IpList
	loadBalance.WeightList = params.WeightList
	loadBalance.UpstreamHeaderTimeout = params.UpstreamHeaderTimeout
	loadBalance.UpstreamIdleTimeout = params.UpstreamIdleTimeout
	loadBalance.UpstreamMaxIdle = params.UpstreamMaxIdle
	loadBalance.UpstreamConnectTimeout = params.UpstreamConnectTimeout
	err = loadBalance.Save(c, tx)
	if err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 2019, errors.New("更新负载均衡表格失败")) // 更新负载均衡表格失败
		return
	}
	tx.Commit()
	// 返回结果
	middleware.ResponseSuccess(c, "更新成功")
}

// serviceDetail godoc
// @Summary 服务详情
// @Description 服务详情
// @Tags 服务管理
// @ID /service/serviceDetail
// @Accept  json
// @Produce  json
// @Param id query int true "服务ID"
// @Success 200 {object} middleware.Response{data=dao.ServiceDetail} "success"
// @Router /service/serviceDetail [get]
func (serviceController *ServiceController) serviceDetail(c *gin.Context) {
	params := &dto.ServiceDeleteInput{}

	if err := params.BindValidParam(c); err != nil {
		middleware.ResponseError(c, 2000, err) // 添加HTTP参数不正确
		return
	}

	//获取数据库连接
	tx, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(c, 2001, err)
		return
	}

	serviceInfo := &dao.ServiceInfo{Id: params.Id}
	serviceInfo, err = serviceInfo.Find(c, tx, serviceInfo)
	if err != nil {
		middleware.ResponseError(c, 2002, err)
		return
	}
	serviceDetail := &dao.ServiceDetail{}
	serviceDetail, err = serviceInfo.ServiceDetail(c, tx, serviceInfo)
	if err != nil {
		middleware.ResponseError(c, 2003, err)
		return
	}

	//返回成功
	middleware.ResponseSuccess(c, serviceDetail)
}

// ServiceStat godoc
// @Summary 服务统计
// @Description 服务统计
// @Tags 服务管理
// @ID /service/service_stat
// @Accept  json
// @Produce  json
// @Param id query string true "服务ID"
// @Success 200 {object} middleware.Response{data=dto.ServiceStatOutput} "success"
// @Router /service/service_stat [get]
func (service *ServiceController) ServiceStat(c *gin.Context) {
	params := &dto.ServiceDeleteInput{}

	if err := params.BindValidParam(c); err != nil {
		middleware.ResponseError(c, 2000, err) // 添加HTTP参数不正确
		return
	}

	var Today []int64
	var Yesterday []int64

	for i := 0; i < time.Now().Hour(); i++ {
		Today = append(Today, 0)
	}
	for i := 0; i < 24; i++ {
		Yesterday = append(Yesterday, 0)
	}

	middleware.ResponseSuccess(c, &dto.ServiceStatOutput{
		Today:     Today,
		Yesterday: Yesterday,
	})
}

// ServiceAddTcp godoc
// @Summary tcp服务添加
// @Description tcp服务添加
// @Tags 服务管理
// @ID /service/service_add_tcp
// @Accept  json
// @Produce  json
// @Param body body dto.ServiceAddTcpInput true "body"
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /service/service_add_tcp [post]
func (admin *ServiceController) ServiceAddTcp(c *gin.Context) {
	params := &dto.ServiceAddTcpInput{}
	if err := params.GetValidParams(c); err != nil {
		middleware.ResponseError(c, 2001, err)
		return
	}

	//验证 service_name 是否被占用
	infoSearch := &dao.ServiceInfo{
		ServiceName: params.ServiceName,
		IsDelete:    0,
	}
	if _, err := infoSearch.Find(c, lib.GORMDefaultPool, infoSearch); err == nil {
		middleware.ResponseError(c, 2002, errors.New("服务名被占用，请重新输入"))
		return
	}

	//验证端口是否被占用?
	tcpRuleSearch := &dao.TcpRule{
		Port: params.Port,
	}
	if _, err := tcpRuleSearch.Find(c, lib.GORMDefaultPool, tcpRuleSearch); err == nil {
		middleware.ResponseError(c, 2003, errors.New("服务端口被占用，请重新输入"))
		return
	}
	grpcRuleSearch := &dao.GrpcRule{
		Port: params.Port,
	}
	if _, err := grpcRuleSearch.Find(c, lib.GORMDefaultPool, grpcRuleSearch); err == nil {
		middleware.ResponseError(c, 2004, errors.New("服务端口被占用，请重新输入"))
		return
	}

	//ip与权重数量一致
	if len(strings.Split(params.IpList, ",")) != len(strings.Split(params.WeightList, ",")) {
		middleware.ResponseError(c, 2005, errors.New("ip列表与权重设置不匹配"))
		return
	}

	tx := lib.GORMDefaultPool.Begin()
	info := &dao.ServiceInfo{
		LoadType:    public.LoadTypeTCP,
		ServiceName: params.ServiceName,
		ServiceDesc: params.ServiceDesc,
	}
	if err := info.Save(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 2006, err)
		return
	}
	loadBalance := &dao.LoadBalance{
		ServiceID:  info.Id,
		RoundType:  params.RoundType,
		IpList:     params.IpList,
		WeightList: params.WeightList,
		ForbidList: params.ForbidList,
	}
	if err := loadBalance.Save(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 2007, err)
		return
	}

	httpRule := &dao.TcpRule{
		ServiceID: info.Id,
		Port:      params.Port,
	}
	if err := httpRule.Save(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 2008, err)
		return
	}

	accessControl := &dao.AccessControl{
		ServiceID:         info.Id,
		OpenAuth:          params.OpenAuth,
		BlackList:         params.BlackList,
		WhiteList:         params.WhiteList,
		WhiteHostName:     params.WhiteHostName,
		ClientIPFlowLimit: params.ClientIPFlowLimit,
		ServiceFlowLimit:  params.ServiceFlowLimit,
	}
	if err := accessControl.Save(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 2009, err)
		return
	}
	tx.Commit()
	middleware.ResponseSuccess(c, "创建TCP成功")
	return
}

// ServiceUpdateTcp godoc
// @Summary tcp服务更新
// @Description tcp服务更新
// @Tags 服务管理
// @ID /service/service_update_tcp
// @Accept  json
// @Produce  json
// @Param body body dto.ServiceUpdateTcpInput true "body"
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /service/service_update_tcp [post]
func (admin *ServiceController) ServiceUpdateTcp(c *gin.Context) {
	params := &dto.ServiceUpdateTcpInput{}
	if err := params.GetValidParams(c); err != nil {
		middleware.ResponseError(c, 2001, err)
		return
	}

	//ip与权重数量一致
	if len(strings.Split(params.IpList, ",")) != len(strings.Split(params.WeightList, ",")) {
		middleware.ResponseError(c, 2002, errors.New("ip列表与权重设置不匹配"))
		return
	}

	tx := lib.GORMDefaultPool.Begin()

	service := &dao.ServiceInfo{
		Id: params.ID,
	}
	detail, err := service.ServiceDetail(c, lib.GORMDefaultPool, service)
	if err != nil {
		middleware.ResponseError(c, 2002, err)
		return
	}

	info := detail.Info
	info.ServiceDesc = params.ServiceDesc
	if err := info.Save(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 2003, err)
		return
	}

	loadBalance := &dao.LoadBalance{}
	if detail.LoadBalance != nil {
		loadBalance = detail.LoadBalance
	}
	loadBalance.ServiceID = info.Id
	loadBalance.RoundType = params.RoundType
	loadBalance.IpList = params.IpList
	loadBalance.WeightList = params.WeightList
	loadBalance.ForbidList = params.ForbidList
	if err := loadBalance.Save(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 2004, err)
		return
	}

	tcpRule := &dao.TcpRule{}
	if detail.TcpRule != nil {
		tcpRule = detail.TcpRule
	}
	tcpRule.ServiceID = info.Id
	tcpRule.Port = params.Port
	if err := tcpRule.Save(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 2005, err)
		return
	}

	accessControl := &dao.AccessControl{}
	if detail.AccessControl != nil {
		accessControl = detail.AccessControl
	}
	accessControl.ServiceID = info.Id
	accessControl.OpenAuth = params.OpenAuth
	accessControl.BlackList = params.BlackList
	accessControl.WhiteList = params.WhiteList
	accessControl.WhiteHostName = params.WhiteHostName
	accessControl.ClientIPFlowLimit = params.ClientIPFlowLimit
	accessControl.ServiceFlowLimit = params.ServiceFlowLimit
	if err := accessControl.Save(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 2006, err)
		return
	}
	tx.Commit()
	middleware.ResponseSuccess(c, "")
	return
}

// ServiceAddGrpc godoc
// @Summary grpc服务添加
// @Description grpc服务添加
// @Tags 服务管理
// @ID /service/service_add_grpc
// @Accept  json
// @Produce  json
// @Param body body dto.ServiceAddGrpcInput true "body"
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /service/service_add_grpc [post]
func (admin *ServiceController) ServiceAddGrpc(c *gin.Context) {
	params := &dto.ServiceAddGrpcInput{}
	if err := params.GetValidParams(c); err != nil {
		middleware.ResponseError(c, 2001, err)
		return
	}

	//验证 service_name 是否被占用
	infoSearch := &dao.ServiceInfo{
		ServiceName: params.ServiceName,
		IsDelete:    0,
	}
	if _, err := infoSearch.Find(c, lib.GORMDefaultPool, infoSearch); err == nil {
		middleware.ResponseError(c, 2002, errors.New("服务名被占用，请重新输入"))
		return
	}

	//验证端口是否被占用?
	tcpRuleSearch := &dao.TcpRule{
		Port: params.Port,
	}
	if _, err := tcpRuleSearch.Find(c, lib.GORMDefaultPool, tcpRuleSearch); err == nil {
		middleware.ResponseError(c, 2003, errors.New("服务端口被占用，请重新输入"))
		return
	}
	grpcRuleSearch := &dao.GrpcRule{
		Port: params.Port,
	}
	if _, err := grpcRuleSearch.Find(c, lib.GORMDefaultPool, grpcRuleSearch); err == nil {
		middleware.ResponseError(c, 2004, errors.New("服务端口被占用，请重新输入"))
		return
	}

	//ip与权重数量一致
	if len(strings.Split(params.IpList, ",")) != len(strings.Split(params.WeightList, ",")) {
		middleware.ResponseError(c, 2005, errors.New("ip列表与权重设置不匹配"))
		return
	}

	tx := lib.GORMDefaultPool.Begin()
	info := &dao.ServiceInfo{
		LoadType:    public.LoadTypeGRPC,
		ServiceName: params.ServiceName,
		ServiceDesc: params.ServiceDesc,
	}
	if err := info.Save(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 2006, err)
		return
	}

	loadBalance := &dao.LoadBalance{
		ServiceID:  info.Id,
		RoundType:  params.RoundType,
		IpList:     params.IpList,
		WeightList: params.WeightList,
		ForbidList: params.ForbidList,
	}
	if err := loadBalance.Save(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 2007, err)
		return
	}

	grpcRule := &dao.GrpcRule{
		ServiceID:      info.Id,
		Port:           params.Port,
		HeaderTransfor: params.HeaderTransfor,
	}
	if err := grpcRule.Save(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 2008, err)
		return
	}

	accessControl := &dao.AccessControl{
		ServiceID:         info.Id,
		OpenAuth:          params.OpenAuth,
		BlackList:         params.BlackList,
		WhiteList:         params.WhiteList,
		WhiteHostName:     params.WhiteHostName,
		ClientIPFlowLimit: params.ClientIPFlowLimit,
		ServiceFlowLimit:  params.ServiceFlowLimit,
	}
	if err := accessControl.Save(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 2009, err)
		return
	}
	tx.Commit()
	middleware.ResponseSuccess(c, "创建GRPC成功")
	return
}

// ServiceUpdateGrpc godoc
// @Summary grpc服务更新
// @Description grpc服务更新
// @Tags 服务管理
// @ID /service/service_update_grpc
// @Accept  json
// @Produce  json
// @Param body body dto.ServiceUpdateGrpcInput true "body"
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /service/service_update_grpc [post]
func (admin *ServiceController) ServiceUpdateGrpc(c *gin.Context) {
	params := &dto.ServiceUpdateGrpcInput{}
	if err := params.GetValidParams(c); err != nil {
		middleware.ResponseError(c, 2001, err)
		return
	}

	//ip与权重数量一致
	if len(strings.Split(params.IpList, ",")) != len(strings.Split(params.WeightList, ",")) {
		middleware.ResponseError(c, 2002, errors.New("ip列表与权重设置不匹配"))
		return
	}

	tx := lib.GORMDefaultPool.Begin()

	service := &dao.ServiceInfo{
		Id: params.ID,
	}
	detail, err := service.ServiceDetail(c, lib.GORMDefaultPool, service)
	if err != nil {
		middleware.ResponseError(c, 2003, err)
		return
	}

	info := detail.Info
	info.ServiceDesc = params.ServiceDesc
	if err := info.Save(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 2004, err)
		return
	}

	loadBalance := &dao.LoadBalance{}
	if detail.LoadBalance != nil {
		loadBalance = detail.LoadBalance
	}
	loadBalance.ServiceID = info.Id
	loadBalance.RoundType = params.RoundType
	loadBalance.IpList = params.IpList
	loadBalance.WeightList = params.WeightList
	loadBalance.ForbidList = params.ForbidList
	if err := loadBalance.Save(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 2005, err)
		return
	}

	grpcRule := &dao.GrpcRule{}
	if detail.GrpcRule != nil {
		grpcRule = detail.GrpcRule
	}
	grpcRule.ServiceID = info.Id
	//grpcRule.Port = params.Port
	grpcRule.HeaderTransfor = params.HeaderTransfor
	if err := grpcRule.Save(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 2006, err)
		return
	}

	accessControl := &dao.AccessControl{}
	if detail.AccessControl != nil {
		accessControl = detail.AccessControl
	}
	accessControl.ServiceID = info.Id
	accessControl.OpenAuth = params.OpenAuth
	accessControl.BlackList = params.BlackList
	accessControl.WhiteList = params.WhiteList
	accessControl.WhiteHostName = params.WhiteHostName
	accessControl.ClientIPFlowLimit = params.ClientIPFlowLimit
	accessControl.ServiceFlowLimit = params.ServiceFlowLimit
	if err := accessControl.Save(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 2007, err)
		return
	}
	tx.Commit()
	middleware.ResponseSuccess(c, "")
	return
}
