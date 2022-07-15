package accountrouter

import (
	"FSRV_Edge/lib"
	"FSRV_Edge/nodeattr"
	"FSRV_Edge/service/routerservice"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type errMsg nodeattr.ErrMsg

// AddAccountRouter 帳戶列表專用router
func AddAccountRouter(router *gin.Engine) {
	api := router.Group("edge")
	api.Use(lib.AuthInterceptor())
	// no authentication endpoints
	{
		api.GET("/account", getAccount)
		api.PUT("/account", editAccount)
		api.GET("/permission", getPermission)
		api.PUT("/permission", editPermission)
	}
}

type accountInfos []nodeattr.EdgeAuth

// @取得帳戶資訊
// @Tags 帳戶列表
// @Description 取得帳戶資訊
// @Accept  json
// @Produce  json
// @Success 200 {object} accountrouter.accountInfos
// @Failure 500 {object} accountrouter.errMsg
// @Router /edge/account [get]
func getAccount(gCtx *gin.Context) {
	var errMsg nodeattr.ErrMsg
	accounts, err := routerservice.GetEdgeAuth()
	if err != nil {
		errMsg.Msg = err.Error()
		gCtx.JSON(http.StatusInternalServerError, errMsg)
		return
	}
	gCtx.JSON(http.StatusOK, accounts)
}

type editAccountData nodeattr.EditEdgeAuth

// @編輯帳戶列表
// @Tags 帳戶列表
// @Description 編輯帳戶列表
// @Accept  json
// @Produce  json
// @Param data body accountrouter.editAccountData true "帳戶列表資料"
// @Success 200 {object} accountrouter.errMsg
// @Failure 500 {object} accountrouter.errMsg
// @Router /edge/account [put]
func editAccount(gCtx *gin.Context) {
	var errMsg nodeattr.ErrMsg
	var editInfo nodeattr.EditEdgeAuth
	defer gCtx.Request.Body.Close()
	byteArr, readErr := ioutil.ReadAll(gCtx.Request.Body)
	if readErr != nil {
		errMsg.Msg = readErr.Error()
		gCtx.JSON(http.StatusInternalServerError, errMsg)
		return
	}
	unmarshalErr := json.Unmarshal(byteArr, &editInfo)
	if unmarshalErr != nil {
		errMsg.Msg = unmarshalErr.Error()
		gCtx.JSON(http.StatusInternalServerError, errMsg)
		return
	}
	updateErr := routerservice.UpdateAccountPwd(editInfo)
	if updateErr != nil {
		errMsg.Msg = updateErr.Error()
		gCtx.JSON(http.StatusInternalServerError, errMsg)
		return
	}
	gCtx.JSON(http.StatusOK, errMsg)
}

type permissionInfos []nodeattr.AuthPermissopn

// @取得帳戶權限
// @Tags 帳戶列表
// @Description 取得帳戶權限 <br> action(value=0:None,1:Read,2:Write)
// @Accept  json
// @Produce  json
// @Param id header int true "帳戶ID"
// @Success 200 {object} accountrouter.permissionInfos
// @Failure 500 {object} accountrouter.errMsg
// @Router /edge/permission [get]
func getPermission(gCtx *gin.Context) {
	var errMsg nodeattr.ErrMsg
	var permissions []nodeattr.AuthPermissopn
	idStr := gCtx.Request.Header.Get("id")
	if id, err := strconv.Atoi(idStr); err != nil {
		errMsg.Msg = err.Error()
		gCtx.JSON(http.StatusInternalServerError, errMsg)
	} else {
		permissions, err = routerservice.GetPermission(int64(id))
		if err != nil {
			errMsg.Msg = err.Error()
			gCtx.JSON(http.StatusInternalServerError, errMsg)
			return
		}
	}
	gCtx.JSON(http.StatusOK, permissions)
}

type editPermissionInfo nodeattr.EditPermission

// @編輯帳戶權限
// @Tags 帳戶列表
// @Description 編輯帳戶權限 <br> action(value=0:None,1:Read,2:Write)
// @Accept  json
// @Produce  json
// @Param data body accountrouter.editPermissionInfo true "帳戶權限資料"
// @Success 200 {object} accountrouter.errMsg
// @Failure 500 {object} accountrouter.errMsg
// @Router /edge/permission [put]
func editPermission(gCtx *gin.Context) {
	var errMsg nodeattr.ErrMsg
	var editInfo nodeattr.EditPermission
	defer gCtx.Request.Body.Close()
	byteArr, readErr := ioutil.ReadAll(gCtx.Request.Body)
	if readErr != nil {
		errMsg.Msg = readErr.Error()
		gCtx.JSON(http.StatusInternalServerError, errMsg)
		return
	}
	unmarshalErr := json.Unmarshal(byteArr, &editInfo)
	if unmarshalErr != nil {
		errMsg.Msg = unmarshalErr.Error()
		gCtx.JSON(http.StatusInternalServerError, errMsg)
		return
	}
	updateErr := routerservice.UpdatePermission(editInfo)
	if updateErr != nil {
		errMsg.Msg = updateErr.Error()
		gCtx.JSON(http.StatusInternalServerError, errMsg)
		return
	}
	gCtx.JSON(http.StatusOK, errMsg)
}
