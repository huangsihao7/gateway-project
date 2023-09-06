package dao

import (
	"gateway-project/dto"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"time"
)

type ServiceInfo struct {
	Id          int       `json:"id" gorm:"primary_key" description:"自增主键"`
	LoadType    int       `json:"load_type" gorm:"column:load_type" description:"负载类型 0=http 1=tcp 2=grpc"`
	ServiceName string    `json:"service_name" gorm:"column:service_name" description:"服务名称"`
	ServiceDesc string    `json:"service_desc" gorm:"column:service_desc" description:"服务描述"`
	CreatedAt   time.Time `json:"create_at" gorm:"column:create_at" description:"创建时间"`
	UpdatedAt   time.Time `json:"update_at" gorm:"column:update_at" description:"更新时间"`
	IsDelete    int       `json:"is_delete" gorm:"column:is_delete" description:"是否删除"`
}

func (t *ServiceInfo) TableName() string {
	return "gateway_service_info"
}

func (t *ServiceInfo) PageList(c *gin.Context, tx *gorm.DB, param *dto.ServiceListInput) ([]ServiceInfo, int64, error) {

	var total int64 = 0
	var list []ServiceInfo
	offset := (param.PageNum - 1) * param.PageSize
	query := tx.WithContext(c)
	query.Table(t.TableName()).Where("is_delete = 0")

	if param.Info != "" {
		query = query.Where("(service_name like %?% or service_desc like %?%)", param.Info, param.Info)
	}
	if err := query.Limit(param.PageSize).Offset(offset).Find(&list).Error; err != nil && err != gorm.ErrRecordNotFound {
		return nil, 0, err
	}
	total = int64(len(list))
	return list, total, nil
}

func (t *ServiceInfo) Find(c *gin.Context, tx *gorm.DB, search *ServiceInfo) (*ServiceInfo, error) {

	out := &ServiceInfo{}
	err := tx.WithContext(c).Where(search).Find(out).Error
	if err != nil {
		return nil, err
	}

	return out, err
}

func (t *ServiceInfo) Save(c *gin.Context, tx *gorm.DB) error {
	return tx.WithContext(c).Save(t).Error
}
