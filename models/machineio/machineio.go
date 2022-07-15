package machineio

// MachineIO machine 一次清洗資料
type MachineIO struct {
	Timestamp  int64 `json:"timestamp"`
	Di0        int   `json:"di0"`
	Di1        int   `json:"di1"`
	Di2        int   `json:"di2"`
	Di3        int   `json:"di3"`
	Di4        int   `json:"di4"`
	Di5        int   `json:"di5"`
	Di6        int   `json:"di6"`
	Di7        int   `json:"di7"`
	Analyzed   int   `json:"analyzed"`
	CreateTime int64 `json:"createTime"`
}
