package dto

import (
	"gateway-project/public"
	"github.com/gin-gonic/gin"
)

type ServiceListInput struct {
	Info     string `json:"info" form:"info" comment:"关键词" example:"网关" validate:""`                  // 关键词
	PageNum  int    `json:"pageNum" form:"pageNum" comment:"页数" example:"1" validate:"required"`      // 页数
	PageSize int    `json:"pageSize" form:"pageSize" comment:"每页数量" example:"20" validate:"required"` // 每页数量
}

func (param *ServiceListInput) BindValidParam(c *gin.Context) error {
	return public.DefaultGetValidParams(c, param)
}

type ServiceItemOutput struct {
	ID          int64  `json:"id" form:"id"`                     // id
	ServiceName string `json:"service_name" form:"service_name"` // 服务名称
	ServiceDesc string `json:"service_desc" form:"service_desc"` // 服务描述
	LoadType    int    `json:"load_type" form:"load_type"`       // 类型
	ServiceAddr string `json:"service_addr" form:"service_addr"` // 服务地址
	Qps         int64  `json:"qps" form:"qps"`                   // qps
	Qpd         int64  `json:"qpd" form:"qpd"`                   // qpd
	TotalNode   int    `json:"totalNode" form:"totalNode"`       // 节点数
}
type ServiceOutput struct {
	Total int64               `json:"total" form:"total" comment:"总数" example:"" validate:""` // 总数
	List  []ServiceItemOutput `json:"list" form:"list" comment:"列表" example:"" validate:""`   // 列表
}
