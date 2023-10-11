package controller

import (
	"encoding/json"
	"fmt"
	"gateway-project/dao"
	"gateway-project/dto"
	"gateway-project/middleware"
	"gateway-project/public"
	"github.com/e421083458/golang_common/lib"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
)

type AdminController struct {
}

func AdminRegister(group *gin.RouterGroup) {
	admin := &AdminController{}
	group.GET("/admin_info", admin.AdminInfo)
	group.POST("/change_pwd", admin.ChangePwd)

}

// AdminInfo godoc
// @Summary 管理员信息
// @Description 管理员信息
// @Tags 管理员接口
// @ID /admin/admin_info
// @Accept  json
// @Produce  json
// @Success 200 {object} middleware.Response{data=dto.AdminInfoOutput} "success"
// @Router /admin/admin_info [get]
func (admin *AdminController) AdminInfo(c *gin.Context) {

	//1.读取session的json 然后转换为结构体
	sess := sessions.Default(c)
	sessInfo := sess.Get(public.AdminSessionInfoKey)
	adminSessInfo := &dto.AdminSessionInfo{}
	if err := json.Unmarshal([]byte(fmt.Sprint(sessInfo)), adminSessInfo); err != nil {
		middleware.ResponseError(c, 2000, err)
		return
	}
	//2.取出数据封装输出

	out := &dto.AdminInfoOutput{
		Id:           adminSessInfo.Id,
		Name:         adminSessInfo.UserName,
		LoginTime:    adminSessInfo.LoginTime,
		Avatar:       "https://img.touxiangwu.com/zb_users/upload/2023/05/202305121683874652100421.jpg",
		Introduction: "你好 我是管理员",
		Roles:        []string{"admin"},
	}
	middleware.ResponseSuccess(c, out)
}

// ChangePwd godoc
// @Summary 修改密码
// @Description 修改密码
// @Tags 管理员接口
// @ID /admin/change_pwd
// @Accept  json
// @Produce  json
// @Param body body dto.ChangePwdInput true "body"
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /admin/change_pwd [post]
func (admin *AdminController) ChangePwd(c *gin.Context) {

	params := &dto.ChangePwdInput{}
	if err := params.BindValidParam(c); err != nil {
		middleware.ResponseError(c, 2000, err)
	}

	//1. 在session读取用户信息到结构体 sessInfo
	sess := sessions.Default(c)
	sessInfo := sess.Get(public.AdminSessionInfoKey)
	adminSessInfo := &dto.AdminSessionInfo{}
	if err := json.Unmarshal([]byte(fmt.Sprint(sessInfo)), adminSessInfo); err != nil {
		middleware.ResponseError(c, 2000, err)
		return
	}

	//2. 根据ID读取数据库信息  adminInfo
	tx, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(c, 2001, err)
		return
	}
	adminInfo := &dao.Admin{}
	adminInfo, err = adminInfo.Find(c, tx, &dao.Admin{UserName: adminSessInfo.UserName})
	if err != nil {
		middleware.ResponseError(c, 2002, err)
		return
	}

	//3. param.password + adminInfo.salt 数据库的盐值  sha256 saltPassword
	saltPassword := public.GenSaltPassword(adminInfo.Salt, params.Password)
	adminInfo.Password = saltPassword

	//4. 新密码保存到数据库
	if err := adminInfo.Save(c, tx); err != nil {
		middleware.ResponseError(c, 2003, err)
		return
	}

	middleware.ResponseSuccess(c, "")
}
