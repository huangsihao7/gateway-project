package dao

import (
	"errors"
	"gateway-project/dto"
	"gateway-project/public"
	"github.com/e421083458/golang_common/lib"
	"github.com/gin-gonic/gin"
	"net/http/httptest"
	"strings"
	"sync"
)

type ServiceDetail struct {
	Info          *ServiceInfo   `json:"info" description:"基本信息"`
	HttpRule      *HttpRule      `json:"http" description:"http_rule"`
	TcpRule       *TcpRule       `json:"tcp" description:"tcp_rule"`
	GrpcRule      *GrpcRule      `json:"grpc" description:"grpc_rule"`
	LoadBalance   *LoadBalance   `json:"loadBalance" description:"loadBalance"`
	AccessControl *AccessControl `json:"accessControl" description:"accessControl"`
}

var ServiceManagerHandler *ServiceManager

func init() {
	ServiceManagerHandler = NewServiceManager()
}

type ServiceManager struct {
	//通过ServiceName 找到detail
	ServiceMap   map[string]*ServiceDetail
	ServiceSlice []*ServiceDetail
	Locker       sync.RWMutex
	init         sync.Once
	err          error
}

func NewServiceManager() *ServiceManager {
	return &ServiceManager{
		ServiceMap:   map[string]*ServiceDetail{},
		ServiceSlice: []*ServiceDetail{},
		Locker:       sync.RWMutex{},
		init:         sync.Once{},
	}
}

func (s *ServiceManager) LoadOnce() error {
	s.init.Do(func() {
		serviceInfo := &ServiceInfo{}
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		tx, err := lib.GetGormPool("default")
		if err != nil {
			s.err = err
			return
		}
		params := &dto.ServiceListInput{PageSize: 99999, PageNum: 1}
		list, _, err := serviceInfo.PageList(c, tx, params)
		if err != nil {
			s.err = err
			return
		}

		s.Locker.Lock()
		defer s.Locker.Unlock()
		for _, listItem := range list {
			tmpItem := listItem
			serviceDetail, err := tmpItem.ServiceDetail(c, tx, &tmpItem)
			if err != nil {
				s.err = err
				return
			}
			s.ServiceMap[tmpItem.ServiceName] = serviceDetail
			s.ServiceSlice = append(s.ServiceSlice, serviceDetail)
		}
	})

	return s.err
}

func (s *ServiceManager) HTTPAccessMode(c *gin.Context) (*ServiceDetail, error) {
	//1、前缀匹配 /abc ==> serviceSlice.rule
	//2、域名匹配 www.test.com ==> serviceSlice.rule
	//host c.Request.Host
	//path c.Request.URL.Path
	host := c.Request.Host
	host = host[0:strings.Index(host, ":")]
	path := c.Request.URL.Path
	for _, serviceItem := range s.ServiceSlice {
		if serviceItem.Info.LoadType != public.LoadTypeHTTP {
			continue
		}
		if serviceItem.HttpRule.RuleType == public.HTTPRuleTypeDomain {
			if serviceItem.HttpRule.Rule == host {
				return serviceItem, nil
			}
		}
		if serviceItem.HttpRule.RuleType == public.HTTPRuleTypePrefixURL {
			if strings.HasPrefix(path, serviceItem.HttpRule.Rule) {
				return serviceItem, nil
			}
		}
	}
	return nil, errors.New("没有匹配到服务")
}
