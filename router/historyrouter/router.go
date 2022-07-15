package historyrouter

import (
	"FSRV_Edge/global"
	"FSRV_Edge/influx"
	"FSRV_Edge/lib"
	"FSRV_Edge/nodeattr"
	"FSRV_Edge/service/edgeservice"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type errMsg nodeattr.ErrMsg

// AddHistoryRouter 歷史資料專用router
func AddHistoryRouter(router *gin.Engine) {
	api := router.Group("edge")
	api.Use(lib.AuthInterceptor())
	// no authentication endpoints
	{
		api.GET("/history", getAllHDA)
		api.GET("/level0history", getL0HDA)
		api.GET("/level1history", getL1HDA)
		api.GET("/level2history", getL2HDA)
		api.GET("/level3history", getL3HDA)
		api.GET("/level4history", getL4HDA)
	}
}

type hdaData []map[string]interface{}

// @取得所有歷史資料
// @Tags 歷史資料
// @Description 取得所有歷史資料<br> 若protocol="OPC UA" Machine_Status(value=1:全自動,2:半自動,3:手動,4:設置,其他數字:Unknown) <br> Machine_Motor(value=1:ON,0:OFF)<br> 若protocol="HTTP" Machine_Status(value=2:停機,3:閒置,4:異常,5:運轉,7:不良品,其他數字:Unknown)
// @Accept  json
// @Produce  json
// @Param id header string true "設備ID"
// @Param selectTime header string true "選擇時間"
// @Param columns header string true "顯示欄位"
// @Success 200 {object} historyrouter.hdaData
// @Failure 500 {object} historyrouter.errMsg
// @Router /edge/history [get]
func getAllHDA(gCtx *gin.Context) {
	var errMsg nodeattr.ErrMsg
	var dev nodeattr.DevInfo
	idStr := gCtx.Request.Header.Get("id")
	selectT := gCtx.Request.Header.Get("selectTime")
	cols := gCtx.Request.Header.Get("columns")
	var arr []string
	err := json.Unmarshal([]byte(cols), &arr)
	if err != nil {
		errMsg.Msg = err.Error()
		gCtx.JSON(http.StatusInternalServerError, errMsg)
		return
	}
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
		if hdaData, err := influx.ReadData(influx.DatabaseName, idStr, selectT, arr, dev.Protocol, true); err != nil {
			errMsg.Msg = err.Error()
			gCtx.JSON(http.StatusInternalServerError, errMsg)
			return
		} else {
			if len(arr) == 0 {
				gCtx.JSON(http.StatusOK, nil)
			} else {
				gCtx.JSON(http.StatusOK, hdaData)
			}
		}

	} else if dev.Protocol == global.HttpStr {
		if hdaData, err := influx.GetAllDcData(idStr, selectT, arr); err != nil {
			errMsg.Msg = err.Error()
			gCtx.JSON(http.StatusInternalServerError, errMsg)
			return
		} else {
			if len(hdaData) == 0 {
				gCtx.JSON(http.StatusOK, nil)
			} else {
				gCtx.JSON(http.StatusOK, hdaData)
			}
		}
	}
	// 新增colums Header，用Json lib轉型成Array
}

type hdaLevelData []edgeservice.NodeStruct

// @取得level0歷史資料
// @Tags 歷史資料
// @Description 取得level0歷史資料
// @Accept  json
// @Produce  json
// @Param id header string true "設備ID"
// @Param selectTime header string true "選擇時間"
// @Success 200 {object} historyrouter.hdaLevelData
// @Failure 500 {object} historyrouter.errMsg
// @Router /edge/level0history [get]
func getL0HDA(gCtx *gin.Context) {
	var errMsg nodeattr.ErrMsg
	idStr := gCtx.Request.Header.Get("id")
	selectT := gCtx.Request.Header.Get("selectTime")
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
		if hdaData, err := influx.ReadData(influx.DatabaseName, idStr, selectT, nil, dev.Protocol, false); err != nil {
			errMsg.Msg = err.Error()
			gCtx.JSON(http.StatusInternalServerError, errMsg)
			return
		} else {
			if len(hdaData) == 0 {
				gCtx.JSON(http.StatusOK, nil)
			} else {
				data := edgeservice.GetLevelHDA(nodeattr.Level0Str, hdaData)
				gCtx.JSON(http.StatusOK, data)
			}
			return
		}
	} else if dev.Protocol == global.HttpStr {
		if hdaData, err := influx.ReadData(influx.DatabaseName, idStr, selectT, nil, dev.Protocol, false); err != nil {
			errMsg.Msg = err.Error()
			gCtx.JSON(http.StatusInternalServerError, errMsg)
		} else {
			if len(hdaData) == 0 {
				gCtx.JSON(http.StatusOK, nil)
			} else {
				data := edgeservice.GetLevelHDA(nodeattr.Level0Str, hdaData)
				gCtx.JSON(http.StatusOK, data)
			}
		}
	}
}

// @取得level1歷史資料
// @Tags 歷史資料
// @Description 取得level1歷史資料<br> 若protocol="OPC UA" Machine_Status(value=1:全自動,2:半自動,3:手動,4:設置,其他數字:Unknown) <br> Machine_Motor(value=1:ON,0:OFF)<br> 若protocol="HTTP" Machine_Status(value=2:停機,3:閒置,4:異常,5:運轉,7:不良品,其他數字:Unknown)
// @Accept  json
// @Produce  json
// @Param id header string true "設備ID"
// @Param selectTime header string true "選擇時間"
// @Success 200 {object} historyrouter.hdaLevelData
// @Failure 500 {object} historyrouter.errMsg
// @Router /edge/level1history [get]
func getL1HDA(gCtx *gin.Context) {
	var errMsg nodeattr.ErrMsg
	idStr := gCtx.Request.Header.Get("id")
	selectT := gCtx.Request.Header.Get("selectTime")

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
		if hdaData, err := influx.ReadData(influx.DatabaseName, idStr, selectT, nil, dev.Protocol, false); err != nil {
			errMsg.Msg = err.Error()
			gCtx.JSON(http.StatusInternalServerError, errMsg)
			return
		} else {
			if len(hdaData) == 0 {
				gCtx.JSON(http.StatusOK, nil)
			} else {
				data := edgeservice.GetLevelHDA(nodeattr.Level1Str, hdaData)
				gCtx.JSON(http.StatusOK, data)
			}
			return
		}
	} else if dev.Protocol == global.HttpStr {
		if hdaData, err := influx.ReadData(influx.DatabaseName, idStr, selectT, nil, dev.Protocol, false); err != nil {
			errMsg.Msg = err.Error()
			gCtx.JSON(http.StatusInternalServerError, errMsg)
			return
		} else {
			if len(hdaData) == 0 {
				gCtx.JSON(http.StatusOK, nil)
			} else {
				data := edgeservice.GetLevelHDA(nodeattr.Level1Str, hdaData)

				gCtx.JSON(http.StatusOK, data)
			}
		}
	}
}

// @取得level2歷史資料
// @Tags 歷史資料
// @Description 取得level2歷史資料
// @Accept  json
// @Produce  json
// @Param id header string true "設備ID"
// @Param selectTime header string true "選擇時間"
// @Success 200 {object} historyrouter.hdaLevelData
// @Failure 500 {object} historyrouter.errMsg
// @Router /edge/level2history [get]
func getL2HDA(gCtx *gin.Context) {
	var errMsg nodeattr.ErrMsg
	idStr := gCtx.Request.Header.Get("id")
	selectT := gCtx.Request.Header.Get("selectTime")

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
		if hdaData, err := influx.ReadData(influx.DatabaseName, idStr, selectT, nil, dev.Protocol, false); err != nil {
			errMsg.Msg = err.Error()
			gCtx.JSON(http.StatusInternalServerError, errMsg)
			return
		} else {
			if len(hdaData) == 0 {
				gCtx.JSON(http.StatusOK, nil)
			} else {
				data := edgeservice.GetLevelHDA(nodeattr.Level2Str, hdaData)
				gCtx.JSON(http.StatusOK, data)
			}
			return
		}
	} else if dev.Protocol == global.HttpStr {
		if hdaData, err := influx.ReadData(influx.DatabaseName, idStr, selectT, nil, dev.Protocol, false); err != nil {
			errMsg.Msg = err.Error()
			gCtx.JSON(http.StatusInternalServerError, errMsg)
		} else {
			if len(hdaData) == 0 {
				gCtx.JSON(http.StatusOK, nil)
			} else {
				data := edgeservice.GetLevelHDA(nodeattr.Level2Str, hdaData)
				gCtx.JSON(http.StatusOK, data)
			}
		}
	}
}

// @取得level3歷史資料
// @Tags 歷史資料
// @Description 取得level3歷史資料
// @Accept  json
// @Produce  json
// @Param id header string true "設備ID"
// @Param selectTime header string true "選擇時間"
// @Success 200 {object} historyrouter.hdaLevelData
// @Failure 500 {object} historyrouter.errMsg
// @Router /edge/level3history [get]
func getL3HDA(gCtx *gin.Context) {
	var errMsg nodeattr.ErrMsg
	idStr := gCtx.Request.Header.Get("id")
	selectT := gCtx.Request.Header.Get("selectTime")

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
		if hdaData, err := influx.ReadData(influx.DatabaseName, idStr, selectT, nil, dev.Protocol, false); err != nil {
			errMsg.Msg = err.Error()
			gCtx.JSON(http.StatusInternalServerError, errMsg)
			return
		} else {
			if len(hdaData) == 0 {
				gCtx.JSON(http.StatusOK, nil)
			} else {
				data := edgeservice.GetLevelHDA(nodeattr.Level3Str, hdaData)
				gCtx.JSON(http.StatusOK, data)
			}
			return
		}
	} else if dev.Protocol == global.HttpStr {
		if hdaData, err := influx.ReadData(influx.DatabaseName, idStr, selectT, nil, dev.Protocol, false); err != nil {
			errMsg.Msg = err.Error()
			gCtx.JSON(http.StatusInternalServerError, errMsg)
		} else {
			if len(hdaData) == 0 {
				gCtx.JSON(http.StatusOK, nil)
			} else {
				data := edgeservice.GetLevelHDA(nodeattr.Level3Str, hdaData)
				gCtx.JSON(http.StatusOK, data)
			}
		}
	}
}

// @取得level4歷史資料
// @Tags 歷史資料
// @Description 取得level4歷史資料
// @Accept  json
// @Produce  json
// @Param id header string true "設備ID"
// @Param selectTime header string true "選擇時間"
// @Success 200 {object} historyrouter.hdaLevelData
// @Failure 500 {object} historyrouter.errMsg
// @Router /edge/level4history [get]
func getL4HDA(gCtx *gin.Context) {
	var errMsg nodeattr.ErrMsg
	idStr := gCtx.Request.Header.Get("id")
	selectT := gCtx.Request.Header.Get("selectTime")

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
		if hdaData, err := influx.ReadData(influx.DatabaseName, idStr, selectT, nil, dev.Protocol, false); err != nil {
			errMsg.Msg = err.Error()
			gCtx.JSON(http.StatusInternalServerError, errMsg)
			return
		} else {
			if len(hdaData) == 0 {
				gCtx.JSON(http.StatusOK, nil)
			} else {
				data := edgeservice.GetLevelHDA(nodeattr.Level4Str, hdaData)
				gCtx.JSON(http.StatusOK, data)
			}
			return
		}
	} else if dev.Protocol == global.HttpStr {
		if hdaData, err := influx.ReadData(influx.DatabaseName, idStr, selectT, nil, dev.Protocol, false); err != nil {
			errMsg.Msg = err.Error()
			gCtx.JSON(http.StatusInternalServerError, errMsg)
		} else {
			if len(hdaData) == 0 {
				gCtx.JSON(http.StatusOK, nil)
			} else {
				data := edgeservice.GetLevelHDA(nodeattr.Level4Str, hdaData)
				gCtx.JSON(http.StatusOK, data)
			}
		}
	}
}
