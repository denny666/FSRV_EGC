package connectrouter

import (
	"FSRV_Edge/global"
	"FSRV_Edge/lib"
	"FSRV_Edge/nodeattr"
	"FSRV_Edge/service/edgeservice"
	"FSRV_Edge/service/routerservice"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type errMsg nodeattr.ErrMsg

// AddConnectRouter 連線列表專用router
func AddConnectRouter(router *gin.Engine) {
	api := router.Group("edge")
	api.Use(lib.AuthInterceptor())
	// no authentication endpoints
	{
		api.GET("/connect", getConnect)
		api.POST("/connect", addConnect)
		api.PUT("/connect", editConnect)
		api.DELETE("/connect", delConnect)
		api.GET("/scanip", scanIP)
	}
}

type conInfos []nodeattr.ConInfo

// @取得連線資訊
// @Tags 連線列表
// @Description 取得連線資訊<br> status(value=0:異常,1:連線,2:斷線)<br> Protocol(value=0:OPC UA,1:HTTP)
// @Accept  json
// @Produce  json
// @Success 200 {object} connectrouter.conInfos
// @Failure 500 {object} connectrouter.errMsg
// @Router /edge/connect [get]
func getConnect(gCtx *gin.Context) {
	global.Cons, _ = routerservice.GetAllConInfoNoPwd()
	gCtx.JSON(http.StatusOK, global.Cons)
}

type conInfo []nodeattr.ConInfo

// @新增連線資訊
// @Tags 連線列表
// @Description 新增連線資訊<br> Protocol(value=0:OPC UA,1:HTTP)
// @Accept  json
// @Produce  json
// @Param data body connectrouter.conInfo true "連線資料"
// @Success 200 {object} connectrouter.errMsg
// @Failure 500 {object} connectrouter.errMsg
// @Router /edge/connect [post]
func addConnect(gCtx *gin.Context) {
	var errMsg nodeattr.ErrMsg
	var con []nodeattr.ConInfo
	defer gCtx.Request.Body.Close()
	byteArr, readErr := ioutil.ReadAll(gCtx.Request.Body)
	if readErr != nil {
		errMsg.Msg = readErr.Error()
		gCtx.JSON(http.StatusInternalServerError, errMsg)
		return
	}
	unmarshalErr := json.Unmarshal(byteArr, &con)
	if unmarshalErr != nil {
		errMsg.Msg = unmarshalErr.Error()
		gCtx.JSON(http.StatusInternalServerError, errMsg)
		return
	}
	for i := 0; i < len(con); i++ {
		con[i].Exist = 1
	}
	err := routerservice.InsertMultiConInfo(con)
	if err != nil {
		errMsg.Msg = err.Error()
		gCtx.JSON(http.StatusInternalServerError, errMsg)
		return
	}
	global.Devs, _ = routerservice.GetAllDevInfo()
	gCtx.JSON(http.StatusOK, errMsg)
}

type updateCon nodeattr.EditConInfo

// @編輯連線資訊
// @Tags 連線列表
// @Description 編輯連線資訊
// @Accept  json
// @Produce  json
// @Param editinfo body connectrouter.updateCon true "編輯資料"
// @Success 200 {object} connectrouter.errMsg
// @Failure 500 {object} connectrouter.errMsg
// @Router /edge/connect [put]
func editConnect(gCtx *gin.Context) {
	var errMsg nodeattr.ErrMsg
	var editInfo nodeattr.EditConInfo
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
	if err := edgeservice.CheckIPAddressType(editInfo.EditContent.IP); err != nil {
		errMsg.Msg = err.Error()
		gCtx.JSON(http.StatusInternalServerError, errMsg)
		return
	}
	updateErr := routerservice.UpdateDevAuth(editInfo)
	if updateErr != nil {
		errMsg.Msg = updateErr.Error()
		gCtx.JSON(http.StatusInternalServerError, errMsg)
		return
	}
	global.Devs, _ = routerservice.GetAllDevInfo()
	gCtx.JSON(http.StatusOK, errMsg)
}

// @刪除連線資訊
// @Tags 連線列表
// @Description 刪除連線資訊
// @Accept  json
// @Produce  json
// @Param id header int true "連線ID"
// @Param name header string true "連線名稱"
// @Success 200 {object} connectrouter.errMsg
// @Failure 500 {object} connectrouter.errMsg
// @Router /edge/connect [delete]
func delConnect(gCtx *gin.Context) {
	var errMsg nodeattr.ErrMsg
	idStr := gCtx.Request.Header.Get("id")
	if id, err := strconv.Atoi(idStr); err != nil {
		errMsg.Msg = err.Error()
		gCtx.JSON(http.StatusInternalServerError, errMsg)
	} else {
		if delErr := routerservice.DelConInfo(id); delErr != nil {
			errMsg.Msg = delErr.Error()
			gCtx.JSON(http.StatusInternalServerError, errMsg)
			return
		}
		gCtx.JSON(http.StatusOK, errMsg)
	}
}

// @取得所有連線
// @Tags 連線列表
// @Description 取得所有連線
// @Accept  json
// @Produce  json
// @Param ip header string true "ip範圍或單一IP"
// @Success 200 {object} connectrouter.conInfos
// @Failure 500 {object} connectrouter.errMsg
// @Router /edge/scanip [get]
func scanIP(gCtx *gin.Context) {
	ip := gCtx.Request.Header.Get("ip")
	var errMsg nodeattr.ErrMsg
	if err := edgeservice.CheckIPAddressType(ip); err != nil {
		errMsg.Msg = err.Error()
		gCtx.JSON(http.StatusInternalServerError, errMsg)
		return
	}
	ipData := edgeservice.ScanIP(ip)
	gCtx.JSON(http.StatusOK, ipData)
}
