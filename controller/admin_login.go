package controller

import (
	"encoding/json"
	"gateway-project/dao"
	"gateway-project/dto"
	"gateway-project/middleware"
	"gateway-project/public"
	"github.com/e421083458/golang_common/lib"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	"time"
)

type AdminLoginController struct {
}

func AdminLoginRegister(group *gin.RouterGroup) {
	admin := &AdminLoginController{}
	group.POST("/login", admin.AdminLogin)
	group.GET("/logout", admin.AdminLogout)
}

// AdminLogin godoc
// @Summary 管理员登录
// @Description 管理员登录
// @Tags 管理员接口
// @ID /admin_login/login
// @Accept  json
// @Produce  json
// @Param body body dto.AdminLoginInput true "body"
// @Success 200 {object} middleware.Response{data=dto.AdminLoginOutput} "success"
// @Router /admin_login/login [post]
func (adminLogin *AdminLoginController) AdminLogin(c *gin.Context) {
	params := &dto.AdminLoginInput{}
	if err := params.BindValidParam(c); err != nil {
		middleware.ResponseError(c, 2000, err)
	}
	tx, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(c, 2001, err)
		return
	}
	admin := &dao.Admin{}
	admin, err = admin.LoginCheck(c, tx, params)
	if err != nil {
		middleware.ResponseError(c, 2002, err)
		return
	}
	//设置session 存入redis
	sessInfo := &dto.AdminSessionInfo{
		Id:        admin.Id,
		UserName:  admin.UserName,
		LoginTime: time.Now(),
	}
	sessBts, err := json.Marshal(sessInfo)
	if err != nil {
		middleware.ResponseError(c, 2003, err)
		return
	}
	sess := sessions.Default(c)
	sess.Set(public.AdminSessionInfoKey, string(sessBts))
	err = sess.Save()
	if err != nil {
		middleware.ResponseError(c, 2004, err)
		return
	}

	out := &dto.AdminLoginOutput{Token: admin.UserName}
	middleware.ResponseSuccess(c, out)
}

// AdminLogout godoc
// @Summary 管理员退出
// @Description 管理员退出
// @Tags 管理员接口
// @ID /admin/logout
// @Accept  json
// @Produce  json
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /admin_login/logout [get]
func (adminLogin *AdminLoginController) AdminLogout(c *gin.Context) {

	sess := sessions.Default(c)
	sess.Delete(public.AdminSessionInfoKey)
	err := sess.Save()
	if err != nil {
		middleware.ResponseError(c, 2000, err)
		return
	}

	middleware.ResponseSuccess(c, "")
}
