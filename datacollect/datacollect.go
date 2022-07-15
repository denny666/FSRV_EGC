package datacollect

import "github.com/cbrake/influxdbhelper/v2"

// 建立 data_di_mac、data_status_mac
// SettingBody 設定資訊
type SettingBody struct {
	UID   int   `json:"UID"`
	MAC   int   `json:"MAC"`
	TmF   int   `json:"TmF"`
	Fltr  int   `json:"Fltr"`
	TSt   int64 `json:"TSt"`
	TEnd  int64 `json:"TEnd"`
	Amt   int   `json:"Amt"`
	Total int   `json:"Total"`
	TLst  int64 `json:"TLst"`
	TFst  int64 `json:"TFst"`
}

// WiseMsg DC採集資料
type WiseMsg struct {
	LogMsg []LogMsg
	Err    int
	Msg    string
}

// LogMsg 設備資料
type LogMsg struct {
	PE      int
	UID     string
	TIM     string
	Record  [8][4]int
	IntTime int64 `json:"-"`
}

// DcSetting DC一次清洗 & 緩存設定
type DcSetting struct {
	// Clean = 0 清洗設定
	Clean bool `json:"clean"`
	// Cache 緩存設定
	Cache bool `json:"cache"`
	// Fq 採集頻率(10~3000)
	Fq int64 `json:"fq"`
	// IdleTime 閒置時間
	IdleTime int64 `json:"idleTime"`
	// DcAbnormalTime 閒置時間
	DcAbnormalTime int64 `json:"dcAbnormalTime"`
}

// SettingField DC設定參數欄位
type SettingField struct {
	Key   string `db:"key"`
	Value string `db:"value"`
}

// MachineIO machine 一次清洗資料
type MachineIO struct {
	InfluxMeasurement influxdbhelper.Measurement
	Timestamp         int64 `db:"timestamp" influx:"timestamp"`
	ID                int64 `db:"id" influx:"-"`
	Di0               int   `db:"di0" influx:"di0"`
	Di1               int   `db:"di1" influx:"di1"`
	Di2               int   `db:"di2" influx:"di2"`
	Di3               int   `db:"di3" influx:"di3"`
	Di4               int   `db:"di4" influx:"di4"`
	Di5               int   `db:"di5" influx:"di5"`
	Di6               int   `db:"di6" influx:"di6"`
	Di7               int   `db:"di7" influx:"di7"`
	Analyzed          int   `db:"analyzed" influx:"analyzed"`
}

// Status machine analyze status
type Status struct {
	InfluxMeasurement influxdbhelper.Measurement
	ID                int64   `db:"id" influx:"-"`
	CycleTime         float64 `db:"cycle_time" influx:"cycle_time"`
	Status            int     `db:"status" influx:"status"`
	GYR               string  `db:"gry" influx:"gry"`
	SCT               float64 `db:"sct" influx:"sct"`
	Timestamp         int64   `db:"timestamp" influx:"timestamp"`
}

// Response 回應
type Response struct {
	StatusCode int    `json:"statusCode"`
	Response   string `json:"response"`
}

// StatusLog 狀態的Log
type StatusLog struct {
	StatusInfo []Status `json:"StatusInfo"`
	Msg        string   `json:"Msg"`
	Err        int      `json:"Err"`
}

const (
	// Stop 停機
	Stop int = 2
	//Idle 閒置
	Idle int = 3
	// Abnormal 異常
	Abnormal int = 4
	// Running 正常(2,1,2)
	Running int = 5
	// NG 不良品
	NG int = 7
	// Sync 同步中
	Sync int = 9
	// OFFLine EG異常
	OFFLine int = 10
	// DcAbnormal 算不出DC
	DcAbnormal int = 11
)
