package connectrecordrouter

import (
	"FSRV_Edge/lib"
	"FSRV_Edge/nodeattr"
	"FSRV_Edge/service/routerservice"
	"net/http"

	"github.com/gin-gonic/gin"
)

type errMsg nodeattr.ErrMsg

// AddConnectRouter 連線列表專用router
func AddConnectRecordRouter(router *gin.Engine) {
	api := router.Group("edge")
	api.Use(lib.AuthInterceptor())
	// no authentication endpoints
	{
		api.GET("/connectrecord", getConnectRecord)
	}
}

type records []nodeattr.ConnectRecord

// @取得系統記錄
// @Tags 系統記錄
// @Description 取得系統記錄(type=0:異常,1:登入,2:登出)
// @Accept  json
// @Produce  json
// @Success 200 {object} connectrecordrouter.records
// @Failure 500 {object} connectrecordrouter.errMsg
// @Router /edge/connectrecord [get]
func getConnectRecord(gCtx *gin.Context) {
	var errMsg nodeattr.ErrMsg
	records, err := routerservice.GetConnectRecord()
	if err != nil {
		errMsg.Msg = err.Error()
		gCtx.JSON(http.StatusInternalServerError, errMsg)
		return
	}
	gCtx.JSON(http.StatusOK, records)
}
