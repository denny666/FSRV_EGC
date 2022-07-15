package loginrouter

import (
	"FSRV_Edge/global"
	"FSRV_Edge/lib"
	"FSRV_Edge/nodeattr"
	"FSRV_Edge/service/routerservice"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type errMsg nodeattr.ErrMsg

const (
	maxAge = 3600
	loginN = 1
)

var (
	errNoRows = errors.New("<QuerySeter> no row found")
)

// AddLoginRouter 登入專用router
func AddLoginRouter(router *gin.Engine) {
	api := router.Group("edge")
	api.Use(lib.RecordInterceptor())
	// no authentication endpoints
	{
		api.GET("/login", loginHandle)
		api.GET("/logout", logoutHandle)
	}
}

// @登入
// @Tags 登入、登出功能
// @Description 登入用
// @Accept  json
// @Produce  json
// @Param account header string true "帳號"
// @Param password header string true "密碼(SHA512格式)"
// @Success 200 {object} loginrouter.errMsg
// @Failure 500 {object} loginrouter.errMsg
// @Router /edge/login [get]
func loginHandle(gCtx *gin.Context) {
	account := gCtx.Request.Header.Get("account")
	password := gCtx.Request.Header.Get("password")
	var errMsg nodeattr.ErrMsg
	user, err := routerservice.GetEdgeAuthByName(account)
	if err != nil {
		if err == errNoRows {
			errMsg.Msg = "user not found"
		} else {
			errMsg.Msg = err.Error()
		}
		gCtx.JSON(http.StatusInternalServerError, errMsg)
	} else {
		if token, err := lib.GeneratorJWT(account); err != nil {
			errMsg.Msg = err.Error()
			gCtx.JSON(http.StatusInternalServerError, errMsg)
		} else {
			if user.Password != password {
				errMsg.Msg = "password not correct"
				gCtx.JSON(http.StatusInternalServerError, errMsg)
			} else {
				var record nodeattr.ConnectRecord
				var accountAuth nodeattr.EdgeAuth
				record.IP = gCtx.ClientIP()
				record.Timestamp = time.Now().UnixNano() / int64(time.Millisecond)
				record.URL = gCtx.Request.RequestURI
				record.Account = account
				record.Type = loginN
				accountAuth.Timestamp = record.Timestamp
				accountAuth.Account = account
				routerservice.UpdateAccountLoginTime(accountAuth) // 更新最後登入時間
				routerservice.InsertConRecordInfo(record)         // 新增登入的連線紀錄
				gCtx.SetCookie(global.Token, token, maxAge, "/", "", false, true)
			}
		}
	}
}

// @登出
// @Tags 登入、登出功能
// @Description 登出用
// @Accept  json
// @Produce  json
// @Param account header string true "帳戶"
// @Success 200 {object} loginrouter.errMsg
// @Failure 500 {object} loginrouter.errMsg
// @Router /edge/logout [get]
func logoutHandle(gCtx *gin.Context) {
	gCtx.SetCookie(global.Token, "", maxAge, "/", "", false, true)
}
