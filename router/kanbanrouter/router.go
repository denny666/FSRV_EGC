package kanbanrouter

import (
	"FSRV_Edge/dao/settingdao"
	"FSRV_Edge/global"
	"FSRV_Edge/influx"
	"FSRV_Edge/lib"
	"FSRV_Edge/nodeattr"
	"FSRV_Edge/service/edgeservice"
	"FSRV_Edge/service/opcuaclient"
	"FSRV_Edge/service/wiseservice"

	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type errMsg nodeattr.ErrMsg

// AddKanbanRouter 即時看板專用router
func AddKanbanRouter(router *gin.Engine) {
	api := router.Group("edge")
	api.Use(lib.AuthInterceptor())
	// no authentication endpoints
	{
		api.GET("/level0info", getLevel0info)
		api.GET("/level1info", getLevel1info)
		api.GET("/level2info", getLevel2info)
		api.GET("/level3info", getLevel3info)
		api.GET("/level4info", getLevel4info)
	}
}

type levelinfo []edgeservice.NodeStruct

// @取得 level 0 資訊
// @Tags 即時看板
// @Description 取得設備 level 0 資訊
// @Accept  json
// @Produce  json
// @Param id header string true "設備ID"
// @Success 200 {object} kanbanrouter.levelinfo
// @Failure 500 {object} kanbanrouter.errMsg
// @Router /edge/level0info [get]
func getLevel0info(gCtx *gin.Context) {
	var nodeArr []edgeservice.NodeStruct
	var node edgeservice.NodeStruct
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
	if dev.Protocol == global.OpcStr {
		opcuaclient.AllNodeData.Lock.Lock()
		dataArr := opcuaclient.AllNodeData.DataMap[idStr][nodeattr.Level0Str]
		opcuaclient.AllNodeData.Lock.Unlock()
		for _, val := range dataArr {
			node.Name = val.Name
			node.Value = val.Value
			nodeArr = append(nodeArr, node)
		}
		if len(nodeArr) == 0 {
			edgeservice.DefaultNode.Lock.Lock()
			data := edgeservice.DefaultNode.DataMap
			edgeservice.DefaultNode.Lock.Unlock()
			var dataArr []map[string]interface{}
			dataArr = append(dataArr, data)
			tmp := edgeservice.GetLevelHDA(nodeattr.Level0Str, dataArr)
			gCtx.JSON(http.StatusOK, tmp)
		} else {
			gCtx.JSON(http.StatusOK, nodeArr)
		}
	} else if dev.Protocol == global.HttpStr {
		data := make(map[string]interface{})
		var dataArr []map[string]interface{}
		dataArr = append(dataArr, data)
		tmp := edgeservice.GetLevelHDA(nodeattr.Level0Str, dataArr)
		gCtx.JSON(http.StatusOK, tmp)
	}
}

// @取得 level 1 資訊
// @Tags 即時看板
// @Description 取得level1即時資料<br> 若protocol="OPC UA" Machine_Status(value=0：手動、1：半自動、2：全自動(電眼)、3：全自動(操作狀態Operation mode)) <br> Machine_Motor(value=1:ON,0:OFF)<br> 若protocol="HTTP" Machine_Status(value=2:停機,3:閒置,4:異常,5:運轉,7:不良品,其他數字:Unknown)
// @Accept  json
// @Produce  json
// @Param id header string true "設備ID"
// @Success 200 {object} kanbanrouter.levelinfo
// @Failure 500 {object} kanbanrouter.errMsg
// @Router /edge/level1info [get]
func getLevel1info(gCtx *gin.Context) {
	var nodeArr []edgeservice.NodeStruct
	var node edgeservice.NodeStruct
	var errMsg nodeattr.ErrMsg
	idStr := gCtx.Request.Header.Get("id")
	var dev nodeattr.DevInfo
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

	if dev.Protocol == global.OpcStr {
		opcuaclient.AllNodeData.Lock.Lock()
		dataArr := opcuaclient.AllNodeData.DataMap[idStr][nodeattr.Level1Str]
		opcuaclient.AllNodeData.Lock.Unlock()
		for _, val := range dataArr {
			node.Name = val.Name
			node.Value = val.Value
			nodeArr = append(nodeArr, node)
		}
		if len(nodeArr) == 0 {
			edgeservice.DefaultNode.Lock.Lock()
			data := edgeservice.DefaultNode.DataMap
			edgeservice.DefaultNode.Lock.Unlock()
			var dataArr []map[string]interface{}
			dataArr = append(dataArr, data)
			tmp := edgeservice.GetLevelHDA(nodeattr.Level1Str, dataArr)
			gCtx.JSON(http.StatusOK, tmp)
		} else {
			gCtx.JSON(http.StatusOK, nodeArr)
		}
	} else if dev.Protocol == global.HttpStr {
		settingdao.MacData.Lock.Lock()
		mac := settingdao.MacData.DataMap[idStr]
		settingdao.MacData.Lock.Unlock()
		wiseservice.AllDcData.Lock.Lock()
		v := wiseservice.AllDcData.DataMap[mac]
		wiseservice.AllDcData.Lock.Unlock()
		data := make(map[string]interface{})
		var dataArr []map[string]interface{}
		data[nodeattr.StatusStr] = dev.Status
		data[nodeattr.CycleTimeStr] = v.CycleTime
		selectT := time.Now().UnixNano() / int64(time.Millisecond)
		num, _ := influx.GetMoldCount(strconv.FormatInt(dev.ID, 10), strconv.FormatInt(selectT, 10))
		data[nodeattr.MoldCountStr] = num
		dataArr = append(dataArr, data)
		tmp := edgeservice.GetLevelHDA(nodeattr.Level1Str, dataArr)
		gCtx.JSON(http.StatusOK, tmp)
	}
}

// @取得 level 2 資訊
// @Tags 即時看板
// @Description 取得設備 level 2 資訊
// @Accept  json
// @Produce  json
// @Param id header string true "設備ID"
// @Success 200 {object} kanbanrouter.levelinfo
// @Failure 500 {object} kanbanrouter.errMsg
// @Router /edge/level2info [get]
func getLevel2info(gCtx *gin.Context) {
	var nodeArr []edgeservice.NodeStruct
	var node edgeservice.NodeStruct
	var errMsg nodeattr.ErrMsg
	idStr := gCtx.Request.Header.Get("id")
	var dev nodeattr.DevInfo
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
	if dev.Protocol == global.OpcStr {
		opcuaclient.AllNodeData.Lock.Lock()
		dataArr := opcuaclient.AllNodeData.DataMap[idStr][nodeattr.Level2Str]
		opcuaclient.AllNodeData.Lock.Unlock()
		for _, val := range dataArr {
			node.Name = val.Name
			node.Value = val.Value
			nodeArr = append(nodeArr, node)
		}
		if len(nodeArr) == 0 {
			edgeservice.DefaultNode.Lock.Lock()
			data := edgeservice.DefaultNode.DataMap
			edgeservice.DefaultNode.Lock.Unlock()
			var dataArr []map[string]interface{}
			dataArr = append(dataArr, data)
			tmp := edgeservice.GetLevelHDA(nodeattr.Level2Str, dataArr)
			gCtx.JSON(http.StatusOK, tmp)
		} else {
			gCtx.JSON(http.StatusOK, nodeArr)
		}
	} else if dev.Protocol == global.HttpStr {
		edgeservice.DefaultNode.Lock.Lock()
		data := edgeservice.DefaultNode.DataMap
		edgeservice.DefaultNode.Lock.Unlock()
		var dataArr []map[string]interface{}
		dataArr = append(dataArr, data)
		tmp := edgeservice.GetLevelHDA(nodeattr.Level2Str, dataArr)
		gCtx.JSON(http.StatusOK, tmp)
	}

}

// @取得 level 3 資訊
// @Tags 即時看板
// @Description 取得設備 level 3 資訊
// @Accept  json
// @Produce  json
// @Param id header string true "設備ID"
// @Success 200 {object} kanbanrouter.levelinfo
// @Failure 500 {object} kanbanrouter.errMsg
// @Router /edge/level3info [get]
func getLevel3info(gCtx *gin.Context) {
	var nodeArr []edgeservice.NodeStruct
	var node edgeservice.NodeStruct
	var errMsg nodeattr.ErrMsg
	idStr := gCtx.Request.Header.Get("id")
	var dev nodeattr.DevInfo
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
	if dev.Protocol == global.OpcStr {
		opcuaclient.AllNodeData.Lock.Lock()
		dataArr := opcuaclient.AllNodeData.DataMap[idStr][nodeattr.Level3Str]
		opcuaclient.AllNodeData.Lock.Unlock()
		for _, val := range dataArr {
			node.Name = val.Name
			node.Value = val.Value
			nodeArr = append(nodeArr, node)
		}
		if len(nodeArr) == 0 {
			edgeservice.DefaultNode.Lock.Lock()
			data := edgeservice.DefaultNode.DataMap
			edgeservice.DefaultNode.Lock.Unlock()
			var dataArr []map[string]interface{}
			dataArr = append(dataArr, data)
			tmp := edgeservice.GetLevelHDA(nodeattr.Level3Str, dataArr)
			gCtx.JSON(http.StatusOK, tmp)
		} else {
			gCtx.JSON(http.StatusOK, nodeArr)
		}
	} else if dev.Protocol == global.HttpStr {
		edgeservice.DefaultNode.Lock.Lock()
		data := edgeservice.DefaultNode.DataMap
		edgeservice.DefaultNode.Lock.Unlock()
		var dataArr []map[string]interface{}
		dataArr = append(dataArr, data)
		tmp := edgeservice.GetLevelHDA(nodeattr.Level3Str, dataArr)
		gCtx.JSON(http.StatusOK, tmp)
	}

}

// @取得 level 4 資訊
// @Tags 即時看板
// @Description 取得設備 level 4 資訊
// @Accept  json
// @Produce  json
// @Param id header string true "設備ID"
// @Success 200 {object} kanbanrouter.levelinfo
// @Failure 500 {object} kanbanrouter.errMsg
// @Router /edge/level4info [get]
func getLevel4info(gCtx *gin.Context) {
	var nodeArr []edgeservice.NodeStruct
	var node edgeservice.NodeStruct
	var errMsg nodeattr.ErrMsg
	idStr := gCtx.Request.Header.Get("id")
	var dev nodeattr.DevInfo
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
	if dev.Protocol == global.OpcStr {
		opcuaclient.AllNodeData.Lock.Lock()
		dataArr := opcuaclient.AllNodeData.DataMap[idStr][nodeattr.Level4Str]
		opcuaclient.AllNodeData.Lock.Unlock()
		for _, val := range dataArr {
			node.Name = val.Name
			node.Value = val.Value
			nodeArr = append(nodeArr, node)
		}
		if len(nodeArr) == 0 {
			edgeservice.DefaultNode.Lock.Lock()
			data := edgeservice.DefaultNode.DataMap
			edgeservice.DefaultNode.Lock.Unlock()
			var dataArr []map[string]interface{}
			dataArr = append(dataArr, data)
			tmp := edgeservice.GetLevelHDA(nodeattr.Level4Str, dataArr)
			gCtx.JSON(http.StatusOK, tmp)
		} else {
			gCtx.JSON(http.StatusOK, nodeArr)
		}
	} else if dev.Protocol == global.HttpStr {
		edgeservice.DefaultNode.Lock.Lock()
		data := edgeservice.DefaultNode.DataMap
		edgeservice.DefaultNode.Lock.Unlock()
		var dataArr []map[string]interface{}
		dataArr = append(dataArr, data)
		tmp := edgeservice.GetLevelHDA(nodeattr.Level4Str, dataArr)
		gCtx.JSON(http.StatusOK, tmp)
	}
}
