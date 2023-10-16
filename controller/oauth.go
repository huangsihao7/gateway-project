package controller

import (
	"encoding/base64"
	"gateway-project/dao"
	"gateway-project/dto"
	"gateway-project/middleware"
	"gateway-project/public"
	"github.com/dgrijalva/jwt-go"
	"github.com/e421083458/golang_common/lib"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"strings"
	"time"
)

type OAuthController struct{}

func OAuthControllerRegister(group *gin.RouterGroup) {
	oauth := &OAuthController{}
	group.POST("/tokens", oauth.Tokens)
}

// Tokens godoc
// @Summary 获取TOKEN
// @Description 获取TOKEN
// @Tags OAUTH
// @ID /oauth/tokens
// @Accept  json
// @Produce  json
// @Param body body dto.TokensInput true "body"
// @Success 200 {object} middleware.Response{data=dto.TokensOutput} "success"
// @Router /oauth/tokens [post]
func (oauth *OAuthController) Tokens(c *gin.Context) {
	//参数获取
	params := &dto.TokensInput{}
	if err := params.BindValidParam(c); err != nil {
		middleware.ResponseError(c, 2000, err)
		return
	}
	//校验Header
	splits := strings.Split(c.GetHeader("Authorization"), " ")
	if len(splits) != 2 {
		middleware.ResponseError(c, 2001, errors.New("用户名或密码格式错误"))
		return
	}
	//base64解码
	appSecret, err := base64.StdEncoding.DecodeString(splits[1])
	if err != nil {
		middleware.ResponseError(c, 2002, err)
		return
	}
	//用户和密码比对
	parts := strings.Split(string(appSecret), ":")
	if err != nil {
		middleware.ResponseError(c, 2003, err)
		return
	}
	if len(parts) != 2 {
		middleware.ResponseError(c, 2002, errors.New("用户名或密码格式错误"))
		return
	}
	//取出appList 校验密码
	appList := dao.AppManagerHandler.GetAppList()
	for _, appItem := range appList {
		tmpItem := appItem
		if tmpItem.AppID == parts[0] && appItem.Secret == parts[1] {
			claims := jwt.StandardClaims{
				Issuer:    tmpItem.AppID,
				ExpiresAt: time.Now().Add(public.JwtExpires * time.Second).In(lib.TimeLocation).Unix(),
			}
			token, err := public.JwtEncode(claims)
			if err != nil {
				middleware.ResponseError(c, 2004, err)
				return
			}
			output := &dto.TokensOutput{
				AccessToken: token,
				ExpiresIn:   public.JwtExpires,
				Scope:       "read_write",
				TokenType:   "Bearer",
			}
			middleware.ResponseSuccess(c, output)
			return
		}
	}

	middleware.ResponseError(c, 2005, errors.New("未匹配正确APP信息"))
}

// AdminLoginOut godoc
// @Summary 管理员退出
// @Description 管理员退出
// @Tags 管理员接口
// @ID /admin_login/logout
// @Accept  json
// @Produce  json
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /admin_login/logout [get]
func (adminlogin *OAuthController) AdminLoginOut(c *gin.Context) {
	sess := sessions.Default(c)
	sess.Delete(public.AdminSessionInfoKey)
	sess.Save()
	middleware.ResponseSuccess(c, "")
}
