package dcinfo

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

// WiseMessage DC 回傳的資料格式
type WiseMessage struct {
	LogMsg []LogMsg `json:"LogMsg"`
	Err    int      `json:"Err"`
	Msg    string   `json:"Msg"`
}

// LogMsg DI 資料 二維陣列
type LogMsg struct {
	PE     int     `json:"PE"`
	UID    string  `json:"UID"`
	TIM    string  `json:"TIM"`
	Record [][]int `json:"Record"`
	SysTk  int64   `json:"SysTk"`
}
