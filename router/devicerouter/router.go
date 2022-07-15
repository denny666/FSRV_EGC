package devicerouter

import (
	"FSRV_Edge/global"
	"FSRV_Edge/lib"
	"FSRV_Edge/nodeattr"
	"FSRV_Edge/service/edgeservice"
	"FSRV_Edge/service/opcuaclient"
	"FSRV_Edge/service/routerservice"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type errMsg nodeattr.ErrMsg

// AddDeviceRouter 設備列表專用router
func AddDeviceRouter(router *gin.Engine) {
	api := router.Group("edge")
	api.Use(lib.AuthInterceptor())
	// no authentication endpoints
	{
		api.GET("/device", getDevice)
		api.PUT("/device", editDevice)
		api.DELETE("/device", delDevice)
		api.PUT("/converter", editConvert)
		api.GET("/converter", getConverter)
		api.POST("/converter", copyConvert)
		api.GET("/browse", getAllBrowse)
		api.GET("/value", getValue)
		api.POST("/checkconverter", checkConvert)
	}
}

type devArr []nodeattr.DevInfo

// @取得設備資訊
// @Tags 設備列表
// @Description 取得設備資訊<br> Protocol(value=0:OPC UA,1:HTTP) <br>若protocol=0 status(value=0:異常,1:運轉,2:閒置,3:Unknown) <br>若protocol=1 status(value=2:停機,3:閒置,4:異常,5:運轉,7:不良品,其他數字:Unknown)
// @Accept  json
// @Produce  json
// @Success 200 {object} devicerouter.devArr
// @Failure 500 {object} devicerouter.errMsg
// @Router /edge/device [get]
func getDevice(gCtx *gin.Context) {
	// var err error
	// var errMsg nodeattr.ErrMsg
	// if global.Devs, err = opcuaservice.GetAllDevInfo(); err != nil {
	// 	errMsg.Msg = err.Error()
	// 	gCtx.JSON(http.StatusInternalServerError, errMsg)
	// 	return
	// }
	var devs []nodeattr.DevInfo
	for _, v := range global.Devs {
		convMap := make(map[string]nodeattr.Converter)
		convInfos, _ := routerservice.GetDataConv("Device"+strconv.Itoa(int(v.ID)), v.TempID)
		tmpRels, _ := routerservice.GetTempRel(v.TempID)
		for _, conv := range convInfos {
			convMap[conv.DstBrowse] = conv
		}
		for _, rel := range tmpRels {
			if rel.ConvFunc != convMap[rel.DstBrowse].ConvFunc || rel.RefBrowseName1 != convMap[rel.DstBrowse].RefBrowseName1 || rel.RefBrowseName2 != convMap[rel.DstBrowse].RefBrowseName2 ||
				rel.RefBrowseName3 != convMap[rel.DstBrowse].RefBrowseName3 || rel.SrcBrowse != convMap[rel.DstBrowse].SrcBrowse || rel.SrcNamespace != convMap[rel.DstBrowse].SrcNamespace ||
				rel.SrcNodeid != convMap[rel.DstBrowse].SrcNodeid {
				v.TempName = v.TempName + "*"
				break
			}
		}
		devs = append(devs, v)
	}
	gCtx.JSON(http.StatusOK, devs)
}

type devInfo nodeattr.EditDeviceInfo

// @編輯設備資訊
// @Tags 設備列表
// @Description 編輯設備資訊
// @Accept  json
// @Produce  json
// @Param data body devicerouter.devInfo true "設備編輯資料"
// @Success 200 {object} devicerouter.errMsg
// @Failure 500 {object} devicerouter.errMsg
// @Router /edge/device [put]
func editDevice(gCtx *gin.Context) {
	var errMsg nodeattr.ErrMsg
	var editInfo nodeattr.EditDeviceInfo
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
	updateErr := routerservice.UpdateDevTemplate(editInfo)
	if updateErr != nil {
		errMsg.Msg = updateErr.Error()
		gCtx.JSON(http.StatusInternalServerError, errMsg)
		return
	}
	global.Devs, _ = routerservice.GetAllDevInfo()
	gCtx.JSON(http.StatusOK, errMsg)
}

// @刪除設備資訊
// @Tags 設備列表
// @Description 刪除設備資訊
// @Accept  json
// @Produce  json
// @Param id header string true "設備ID"
// @Success 200 {object} devicerouter.errMsg
// @Failure 500 {object} devicerouter.errMsg
// @Router /edge/device [delete]
func delDevice(gCtx *gin.Context) {
	idStr := gCtx.Request.Header.Get("id")
	var errMsg nodeattr.ErrMsg
	if id, err := strconv.Atoi(idStr); err != nil {
		errMsg.Msg = err.Error()
		gCtx.JSON(http.StatusInternalServerError, errMsg)
	} else {
		if delErr := routerservice.DelDevInfo(id); delErr != nil {
			errMsg.Msg = err.Error()
			gCtx.JSON(http.StatusInternalServerError, errMsg)
			return
		}
		routerservice.DelDataConvDev("Device" + idStr)
	}
	global.Devs, _ = routerservice.GetAllDevInfo()
	gCtx.JSON(http.StatusOK, errMsg)
}

type edConvInfo nodeattr.EditConverter

// @編輯資料轉換表資訊
// @Tags 設備列表
// @Description 編輯資料轉換表資訊
// @Accept  json
// @Produce  json
// @Param data body devicerouter.edConvInfo true "資料轉換表資料"
// @Success 200 {object} devicerouter.errMsg
// @Failure 500 {object} devicerouter.errMsg
// @Router /edge/converter [put]
func editConvert(gCtx *gin.Context) {
	var errMsg nodeattr.ErrMsg
	var editInfo nodeattr.EditConverter
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
	if err := routerservice.UpdateConvter(editInfo); err != nil {
		errMsg.Msg = err.Error()
		gCtx.JSON(http.StatusInternalServerError, errMsg)
		return
	}
	devIdArr := strings.Split(editInfo.Name, "Device")
	if len(devIdArr) > 1 {
		opcuaclient.DevSub.Lock.Lock()
		opcuaclient.DevSub.DataMap[devIdArr[1]] = false
		opcuaclient.DevSub.Lock.Unlock()
		edgeservice.UpdateConv(devIdArr[1])
	}

	global.Devs, _ = routerservice.GetAllDevInfo()
	gCtx.JSON(http.StatusOK, errMsg)
}

type convInfoArr []nodeattr.Converter

// @取得轉換表資訊
// @Tags 設備列表
// @Description 取得轉換表資訊
// @Accept  json
// @Produce  json
// @Param id header string true "設備ID"
// @Success 200 {object} devicerouter.convInfoArr
// @Failure 500 {object} devicerouter.errMsg
// @Router /edge/converter [get]
func getConverter(gCtx *gin.Context) {
	var errMsg nodeattr.ErrMsg
	var dev nodeattr.DevInfo
	idStr := gCtx.Request.Header.Get("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		errMsg.Msg = err.Error()
		gCtx.JSON(http.StatusInternalServerError, errMsg)
		return
	}
	// 檢查設備是否存在
	for _, v := range global.Devs {
		if id == v.ID {
			dev = v
			break
		}
	}
	if dev.ID == 0 {
		errMsg.Msg = "Device not found"
		gCtx.JSON(http.StatusInternalServerError, errMsg)
		return
	}
	if convInfos, err := routerservice.GetDataConv(dev.Name, dev.TempID); err != nil {
		errMsg.Msg = err.Error()
		gCtx.JSON(http.StatusInternalServerError, errMsg)
	} else {
		gCtx.JSON(http.StatusOK, convInfos)
	}
}

// @取得所有Browse name
// @Tags 設備列表
// @Description 取得所有Browse name
// @Accept  json
// @Produce  json
// @Param id header int true "設備ID"
// @Success 200 {object} devicerouter.errMsg
// @Failure 500 {object} devicerouter.errMsg
// @Router /edge/browse [get]
func getAllBrowse(gCtx *gin.Context) {
	idStr := gCtx.Request.Header.Get("id")
	opcuaclient.BrowseNameMap.Lock.Lock()
	bArr := opcuaclient.BrowseNameMap.DataMap[idStr]
	opcuaclient.BrowseNameMap.Lock.Unlock()
	gCtx.JSON(http.StatusOK, bArr)
}

// @檢查轉換表是否被編輯
// @Tags 設備列表
// @Description 檢查轉換表是否被編輯
// @Accept  json
// @Produce  json
// @Param data body devicerouter.edConvInfo true "轉換表資料"
// @Success 200 {object} devicerouter.errMsg
// @Failure 500 {object} devicerouter.errMsg
// @Router /edge/checkconverter [post]
func checkConvert(gCtx *gin.Context) {
	var errMsg nodeattr.ErrMsg
	var editInfo nodeattr.EditConverter
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
	if msg, err := routerservice.CheckConvert(editInfo); err != nil {
		errMsg.Msg = err.Error()
		gCtx.JSON(http.StatusInternalServerError, errMsg)
	} else {
		gCtx.JSON(http.StatusOK, msg)
	}
}

// @複製轉換表為新範本
// @Tags 設備列表
// @Description 複製轉換表為新範本(data需提供轉換表所有欄位資料，name填入新範本名稱)
// @Accept  json
// @Produce  json
// @Param data body devicerouter.edConvInfo true "轉換表所有欄位資料"
// @Success 200 {object} devicerouter.errMsg
// @Failure 500 {object} devicerouter.errMsg
// @Router /edge/converter [post]
func copyConvert(gCtx *gin.Context) {
	var errMsg nodeattr.ErrMsg
	var editInfo nodeattr.EditConverter
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
	if err := routerservice.CopyConverter(editInfo); err != nil {
		if strings.Contains(err.Error(), global.UniqueErrMsg) {
			errMsg.Msg = "Template name is exist"
			gCtx.JSON(http.StatusInternalServerError, errMsg)
		} else {
			errMsg.Msg = err.Error()
			gCtx.JSON(http.StatusInternalServerError, errMsg)
		}
		return
	}
	global.Devs, _ = routerservice.GetAllDevInfo()
	gCtx.JSON(http.StatusOK, errMsg)
}

// @取得欄位實際值
// @Tags 設備列表
// @Description 取得欄位實際值
// @Accept  json
// @Produce  json
// @Param nodeId header int true "欄位NodeId"
// @Param namespace header int true "欄位namespace"
// @Param browseName header int true "欄位browseName"
// @Param devId header int true "設備ID"
// @Success 200 {object} string
// @Failure 500 {object} devicerouter.errMsg
// @Router /edge/value [get]
func getValue(gCtx *gin.Context) {
	var errMsg nodeattr.ErrMsg
	var dev nodeattr.DevInfo
	nodeID := gCtx.Request.Header.Get("nodeId")
	browse := gCtx.Request.Header.Get("browseName")
	nsStr := gCtx.Request.Header.Get("namespace")
	devID := gCtx.Request.Header.Get("devId")
	id, err := strconv.ParseInt(devID, 10, 64)
	if err != nil {
		errMsg.Msg = err.Error()
		gCtx.JSON(http.StatusInternalServerError, errMsg)
		return
	}
	// 檢查設備是否存在
	for _, v := range global.Devs {
		if id == v.ID {
			dev = v
			break
		}
	}
	if dev.ID == 0 {
		errMsg.Msg = "Device not found"
		gCtx.JSON(http.StatusInternalServerError, errMsg)
		return
	}
	if val, err := opcuaclient.GetActVal(devID, nodeID, nsStr, browse, dev); err != nil {
		errMsg.Msg = err.Error()
		gCtx.JSON(http.StatusInternalServerError, errMsg)
		return
	} else {
		gCtx.JSON(http.StatusOK, val)
	}
}
