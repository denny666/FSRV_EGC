package collect

import (
	"FSRV_Edge/models/machineio"
)

// Info EG採集所需資訊
type Info struct {
	//MachineNumber machine資料表的 machine number
	MachineNumber string `json:"machineNumber"`
	// MacAddress DC的mac address
	MacAddress      string `json:"macAddress"`
	PutTimeInterval int    `json:"putTimeInterval"`
	IdleTime        int    `json:"idleTime"`
	DcAuthorization string `json:"dcAuthorization"`
}

// DatacollectRes 從SRV 回傳回來的資料
type DatacollectRes struct {
	Info     []Info `json:"info"`
	Response string `json:"response"`
}

// FetchTimeInfo 抓取時間範圍資訊
type FetchTimeInfo struct {
	Min      int64  `json:"min"`
	Max      int64  `json:"max"`
	Response string `json:"response"`
}

// LastDataInfo 最後一筆DC Data 的資訊
type LastDataInfo struct {
	Data     machineio.MachineIO `json:"data"`
	Response string              `json:"response"`
}

// Response Response
type Response struct {
	Response string `json:"response"`
}

// RD 回報裝置
type RD struct {
	MacAddress string `json:"macAddress"`
}

// DS  看板裝置 struct
type DS struct {
	ID             int64  `json:"id"`
	URL            string `json:"url"`
	Status         string `json:"status"`
	CreateTime     int64  `json:"createTime"`
	MacAddress     string `json:"macAddress"`
	WorkShopNumber string `json:"workShopNumber"`
}
