package templaterouter

import (
	"FSRV_Edge/global"
	"FSRV_Edge/lib"
	"FSRV_Edge/nodeattr"
	"FSRV_Edge/service/routerservice"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type errMsg nodeattr.ErrMsg

const (
	tmpExistMsg     = "template name exist"
	tempNotExistMsg = "template is not exist"
)

// AddTemplateRouter 範本專用router
func AddTemplateRouter(router *gin.Engine) {
	api := router.Group("edge")
	api.Use(lib.AuthInterceptor())
	// no authentication endpoints
	{
		api.GET("/template", getTemp)
		api.POST("/template", addTemp)
		api.PUT("/template", editTemp)
		api.DELETE("/template", delTemp)
		api.GET("/temprel", getTempRel)
		api.PUT("/temprel", editTempRel)
		api.GET("/model", getModel)
		api.POST("/importtemplate", importTemp)
	}
}

type tmp []nodeattr.Template
type modelArr []string

var templateArr []nodeattr.Template

// @取得範本列表
// @Tags 範本列表
// @Description 取得範本列表
// @Accept  json
// @Produce  json
// @Success 200 {object} templaterouter.tmp
// @Failure 500 {object} templaterouter.errMsg
// @Router /edge/template [get]
func getTemp(gCtx *gin.Context) {
	var errMsg nodeattr.ErrMsg
	var err error
	templateArr, err = routerservice.GetAllTemplate()
	if err != nil {
		errMsg.Msg = err.Error()
		gCtx.JSON(http.StatusInternalServerError, errMsg)
		return
	}
	gCtx.JSON(http.StatusOK, templateArr)
}

// @取得設備模型資料
// @Tags 範本列表
// @Description 取得設備模型資料
// @Accept  json
// @Produce  json
// @Success 200 {object} templaterouter.modelArr
// @Failure 500 {object} templaterouter.errMsg
// @Router /edge/model [get]
func getModel(gCtx *gin.Context) {
	var errMsg nodeattr.ErrMsg
	model, err := routerservice.GetDataConvDev()
	if err != nil {
		errMsg.Msg = err.Error()
		gCtx.JSON(http.StatusInternalServerError, errMsg)
		return
	}
	gCtx.JSON(http.StatusOK, model)
}

type tmpInfo nodeattr.Template

// @新增範本列表
// @Tags 範本列表
// @Description 新增範本列表 <br>使用複製範本功能，id填入來源範本id，name填入新範本名稱
// @Accept  json
// @Produce  json
// @Param data body templaterouter.tmpInfo true "範本資料"
// @Success 200 {object} templaterouter.errMsg
// @Failure 500 {object} templaterouter.errMsg
// @Router /edge/template [post]
func addTemp(gCtx *gin.Context) {
	// 根據前端提供的範本名稱建立 TempName 和 Template_id 的 Map
	var errMsg nodeattr.ErrMsg
	var tmp nodeattr.Template
	defer gCtx.Request.Body.Close()
	byteArr, readErr := ioutil.ReadAll(gCtx.Request.Body)
	if readErr != nil {
		errMsg.Msg = readErr.Error()
		gCtx.JSON(http.StatusInternalServerError, errMsg)
		return
	}
	unmarshalErr := json.Unmarshal(byteArr, &tmp)
	if unmarshalErr != nil {
		errMsg.Msg = unmarshalErr.Error()
		gCtx.JSON(http.StatusInternalServerError, errMsg)
		return
	}

	if err := routerservice.InsertTemp(tmp, templateArr); err != nil {
		errMsg.Msg = err.Error()
		gCtx.JSON(http.StatusInternalServerError, errMsg)
		return
	}
	gCtx.JSON(http.StatusOK, errMsg)
}

type updateTemp nodeattr.EditTemplate

// @編輯範本列表
// @Tags 範本列表
// @Description 編輯範本列表
// @Accept  json
// @Produce  json
// @Param editinfo body templaterouter.updateTemp true "編輯資料"
// @Success 200 {object} templaterouter.errMsg
// @Failure 500 {object} templaterouter.errMsg
// @Router /edge/template [put]
func editTemp(gCtx *gin.Context) {
	var errMsg nodeattr.ErrMsg
	var editInfo nodeattr.EditTemplate
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
	// 檢查範本名稱是否重複
	for _, v := range templateArr {
		if editInfo.EditContent.Name == v.Name { // 名稱重複
			errMsg.Msg = tmpExistMsg
			gCtx.JSON(http.StatusInternalServerError, errMsg)
			return
		}
	}
	editInfo.EditContent.ModifyTime = time.Now().UnixNano() / int64(time.Millisecond)
	updateErr := routerservice.UpdateTemplate(editInfo)
	if updateErr != nil {
		errMsg.Msg = updateErr.Error()
		gCtx.JSON(http.StatusInternalServerError, errMsg)
		return
	}
	gCtx.JSON(http.StatusOK, errMsg)
}

// @刪除範本列表
// @Tags 範本列表
// @Description 刪除範本列表
// @Accept  json
// @Produce  json
// @Param id header int true "範本名稱"
// @Success 200 {object} templaterouter.errMsg
// @Failure 500 {object} templaterouter.errMsg
// @Router /edge/template [delete]
func delTemp(gCtx *gin.Context) {

	idStr := gCtx.Request.Header.Get("id")
	var errMsg nodeattr.ErrMsg
	if id, err := strconv.Atoi(idStr); err != nil {
		errMsg.Msg = err.Error()
		gCtx.JSON(http.StatusInternalServerError, errMsg)
	} else {
		for _, v := range global.Devs {
			if int64(id) == v.TempID { // 範本已被套用
				errMsg.Msg = "Template is used"
				gCtx.JSON(http.StatusInternalServerError, errMsg)
				return
			}
		}
		if delErr := routerservice.DelTemplate(id); delErr != nil {
			errMsg.Msg = delErr.Error()
			gCtx.JSON(http.StatusInternalServerError, errMsg)
			return
		}
		if delErr := routerservice.DelTempRelByTempID(int64(id)); delErr != nil {
			errMsg.Msg = delErr.Error()
			gCtx.JSON(http.StatusInternalServerError, errMsg)
			return
		}
	}
	gCtx.JSON(http.StatusOK, errMsg)
}

type tmpRels []nodeattr.TemplateRel

// @取得範本關聯資訊
// @Tags 範本列表
// @Description 取得範本關聯資訊
// @Accept  json
// @Produce  json
// @Param id header int true "範本ID"
// @Success 200 {object} templaterouter.tmpRels
// @Failure 500 {object} templaterouter.errMsg
// @Router /edge/temprel [get]
func getTempRel(gCtx *gin.Context) {
	var errMsg nodeattr.ErrMsg
	idStr := gCtx.Request.Header.Get("id")
	if id, err := strconv.Atoi(idStr); err != nil {
		errMsg.Msg = err.Error()
		gCtx.JSON(http.StatusInternalServerError, errMsg)
	} else {
		if tmpRels, err := routerservice.GetTempRel(int64(id)); err != nil {
			errMsg.Msg = err.Error()
			gCtx.JSON(http.StatusInternalServerError, errMsg)
		} else {
			gCtx.JSON(http.StatusOK, tmpRels)
		}
	}
}

type editTmpRel nodeattr.EditTemplateRel

// @編輯範本關聯資訊
// @Tags 範本列表
// @Description 編輯範本關聯資訊
// @Accept  json
// @Produce  json
// @Param data body templaterouter.editTmpRel true "範本關聯表資料"
// @Success 200 {object} templaterouter.errMsg
// @Failure 500 {object} templaterouter.errMsg
// @Router /edge/temprel [put]
func editTempRel(gCtx *gin.Context) {
	var errMsg nodeattr.ErrMsg
	var editInfo nodeattr.EditTemplateRel
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
	if err := routerservice.UpdateTempRel(editInfo, templateArr); err != nil {
		errMsg.Msg = err.Error()
		gCtx.JSON(http.StatusInternalServerError, errMsg)
		return
	}
	gCtx.JSON(http.StatusOK, errMsg)
}

// @匯入範本
// @Tags 範本列表
// @Description 匯入範本
// @Accept  json
// @Produce  json
// @Param data body templaterouter.editTmpRel true "範本關聯表資料"
// @Success 200 {object} templaterouter.errMsg
// @Failure 500 {object} templaterouter.errMsg
// @Router /edge/importtemplate [post]
func importTemp(gCtx *gin.Context) {
	var errMsg nodeattr.ErrMsg
	var editInfo nodeattr.EditTemplateRel
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
	var tmp nodeattr.Template
	tmp.Name = editInfo.Name
	if err := routerservice.ImportTemp(tmp, templateArr, editInfo); err != nil {
		errMsg.Msg = err.Error()
		gCtx.JSON(http.StatusInternalServerError, errMsg)
		return
	}
	gCtx.JSON(http.StatusOK, errMsg)
}
