package global

import (
	"FSRV_Edge/nodeattr"
	"net/http"
	"sync"
)

const (
	Token        = "_jwt"
	loginN       = 1
	logoutN      = 2
	abnormal     = 0
	maxAge       = 3600
	OpcStr       = 0
	HttpStr      = 1
	UniqueErrMsg = "UNIQUE constraint failed"
)

var (
	WorkShopNumber   string
	MainIP           string
	BackUpIP         string
	PingWiseSetting  int
	WriteDBSetting   int
	BuildNodeSetting int
)
var DcTr *http.Transport

// SCTMap 最後SCT資料<machineNumber,data[]>
var SCTMap sync.Map

// LastFetchTimeMap 最後fetch 時間 <machineNumber,timestamp>
var LastFetchTimeMap sync.Map

// LastStatusMap  最後一筆狀態資料 <machineNumber,status>
var LastStatusMap sync.Map

// FirstStatusMap  第一筆狀態資料 <machineNumber,status>
var FirstStatusMap sync.Map

// LastDataMap  最後一筆DC資料 <machineNumber,DCIO>
var LastDataMap sync.Map

// LastScheduleMap 最後一筆排程(排程開始的時候才會寫入)
var LastScheduleMap sync.Map

// TaskQTYMap 紀錄所有的QTY緩存
var TaskQTYMap sync.Map

// LastNormalTimeMap 最後正常抓取的時間
var LastNormalTimeMap sync.Map

// Devs 所有設備
var Devs []nodeattr.DevInfo

// Cons 所有連線
var Cons []nodeattr.ConInfo

// TaskQTYInfo 每台設備任務的QTY快取
type TaskQTYInfo struct {
	StartTime int64
	QTY       int64
}

// StatusTime 每台設備狀態時間
type StatusTime struct {
	Timestamp int64
	Status    int
}
