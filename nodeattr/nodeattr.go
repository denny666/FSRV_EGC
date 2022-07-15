package nodeattr

import (
	"flag"
	"strings"
	"time"

	"github.com/cbrake/influxdbhelper/v2"
	"github.com/gopcua/opcua/ua"
)

var (
	Certfile       = flag.String("cert", "", "Path to certificate file")
	Keyfile        = flag.String("key", "", "Path to PEM Private Key file")
	Gencert        = flag.Bool("gen-cert", false, "Generate a new certificate")
	Policy         = flag.String("sec-policy", "None", "Security Policy URL or one of None, Basic128Rsa15, Basic256, Basic256Sha256")
	Mode           = flag.String("sec-mode", "Sign", "Security Mode: one of None, Sign, SignAndEncrypt")
	Auth           = flag.String("auth-mode", "UserName", "Authentication Mode: one of Anonymous, UserName, Certificate")
	Appuri         = flag.String("app-uri", "urn:gopcua:client", "Application URI")
	List           = flag.Bool("list", false, "List the policies supported by the endpoint and exit")
	NonePolicy     = "None"
	Basic128Rsa15  = "Basic128Rsa15"
	Basic256       = "Basic256"
	Basic256Sha256 = "Basic256Sha256"
	// NodeidMap 輸入NodeID，取得對應BrowseName
	NodeidMap = make(map[string]string)
)

const (
	OpcTitle            = "opc.tcp://"
	DefaultTmpName      = "IoMCP 2.0"
	DefaultOPCPort      = 8886
	DefaultEGPort       = 8887
	DefaultEGPortStr    = ":8887"
	connectRecordTable  = "connect_record"
	devAuthTable        = "device_auth"
	tempRelTable        = "node_template_rel"
	tempTable           = "node_template"
	nodeStructTable     = "node_structure"
	converterTable      = "data_converter"
	edgeAuthTable       = "edge_auth"
	devInfoTable        = "device_info"
	authPermissionTable = "auth_permission"
	systemSettingTable  = "system_setting"
	stageStr            = "Stage"
	LevelStr            = "LEVEL"
	CycleTimeStr        = "Cycle_Time"
	MoldCountStr        = "Mold_Count"
	MoldSumUp           = "Machine_molded_pieoces_sum_up"
	BrandStr            = "Machine_Brand"
	TypeStr             = "Machine_Type"
	MotorStr            = "Machine_Motor"
	StatusStr           = "Machine_Status"
	AlarmStr            = "Alarm_Status"
	HeaterStr           = "Machine_Heater"
	MopTimeStr          = "MoldOpen_Time"
	MclTimeStr          = "MoldClose_Time"
	SerialNumStr        = "Machine_Serial_Number"
	IoMVersion          = "ACMT_IoMCP_Version"
	ManuDateStr         = "Machine_Manufacture_Date"
	ProductionMoldName  = "Production_Mold_name"
	PartsNumber         = "Parts_number"
	ResinType           = "Resin_type(Product_Series)"
	Level0Str           = "LEVEL0"
	Level1Str           = "LEVEL1"
	Level2Str           = "LEVEL2"
	Level3Str           = "LEVEL3"
	Level4Str           = "LEVEL4"
	Level5Str           = "LEVEL5"
	HourStr             = "_Hour"
	MinStr              = "_Min"
	DayStr              = "_Day"
	MonthStr            = "_Month"
	YearStr             = "_Year"
	EstimatedTime       = "Estimated_time_to_complete_product"
	MonthDate           = "Every_Month/500Hr_Date_reset"
	Month3Date          = "Every_3_Month/1500Hr_Date_reset"
	Month6Date          = "Every_6_Month/500Hr_Date_reset"
	PowerOnTime         = "Power_on_time_sum_up"
)

// SystemSetting 系統設定
type SystemSetting struct {
	ID   int64  `json:"id" orm:"column(id)"`
	IP   string `json:"ip" orm:"column(ip)"`
	Port int    `json:"port" orm:"column(port)"`
}

// AuthPermissopn 帳戶權限
type AuthPermissopn struct {
	ID         int64  `json:"id" orm:"column(id)"`
	AccountID  int64  `json:"accountId" orm:"column(account_id)"`
	BrowseName string `json:"browseName" orm:"column(browse_name)"`
	Action     int    `json:"action" orm:"column(action)"`
}

// EditPermission 編輯帳戶權限
type EditPermission struct {
	ID          int64            `json:"id" orm:"column(id)"`
	EditContent []AuthPermissopn `json:"editContent"`
}

// EdgeAuth 帳戶列表
type EdgeAuth struct {
	ID        int64  `json:"id" orm:"column(id)"`
	Name      string `json:"name" orm:"column(name)"`
	Account   string `json:"account" orm:"column(account);unique"`
	Password  string `json:"password" orm:"column(password)"`
	Timestamp int64  `json:"timestamp" orm:"column(timestamp)"`
}

// EditEdgeAuth 編輯帳戶資料
type EditEdgeAuth struct {
	ID          int64    `json:"id" orm:"column(id)"`
	EditContent EdgeAuth `json:"editContent"`
}

// NodeDef Node屬性
type NodeDef struct {
	NodeID      *ua.NodeID
	NodeClass   ua.NodeClass
	BrowseName  string
	Description string
	AccessLevel ua.AccessLevelType
	Path        string
	DataType    string
	Writable    bool
	Unit        string
	Scale       string
	Min         string
	Max         string
}

// NodeStruct 預設節點欄位
type NodeStruct struct {
	ID               int64  `orm:"column(id)"`
	Namespace        string `orm:"column(namespace)"`
	Level            string `orm:"-"`
	Group            string `orm:"-"`
	BrowseName       string `orm:"column(browse_name)"`
	ParentBrowseName string `orm:"column(parent_browse_name)"`
	Nodeid           string `orm:"-"`
	Description      string `orm:"column(description)"`
	DataType         string `orm:"column(data_type)"`
}

// TemplateRel 範本關聯資料
type TemplateRel struct {
	ID             int64  `json:"id" orm:"column(id)"`
	TempID         int64  `json:"templateId" orm:"column(template_id)"`
	SrcNamespace   int    `json:"srcNamespace" orm:"column(src_namespace)"`
	SrcBrowse      string `json:"srcBrowseName" orm:"column(src_browse_name)"`
	SrcNodeid      string `json:"srcNodeId" orm:"column(src_node_id)"`
	SrcUnit        string `json:"srcUnit" orm:"column(src_unit)"`
	DstBrowse      string `json:"dstBrowseName" orm:"column(dst_browse_name)"`
	DstNodeid      string `json:"dstNodeId" orm:"column(dst_node_id)"`
	ConvFunc       string `json:"convertFunc" orm:"column(convert_func)"`
	RefBrowseName1 string `json:"referenceBrowseName1" orm:"column(reference_browse_name_1)"`
	RefBrowseName2 string `json:"referenceBrowseName2" orm:"column(reference_browse_name_2)"`
	RefBrowseName3 string `json:"referenceBrowseName3" orm:"column(reference_browse_name_3)"`
}

// GetStage 回傳空字串表示無stage object
func (c *TemplateRel) GetStage() string {
	strArr := strings.Split(c.DstNodeid, ".")
	for _, v := range strArr {
		if strings.Contains(v, stageStr) {
			return v
		}
	}
	return ""
}

// EditTemplateRel 範本關聯資料
type EditTemplateRel struct {
	Name        string        `json:"name"`
	TempID      int64         `json:"templateId"`
	EditContent []TemplateRel `json:"editContent"`
}

// Converter 轉換表
type Converter struct {
	ID             int64  `json:"id" json:"templateId" orm:"column(id)"`
	SrcNamespace   int    `json:"srcNamespace" orm:"column(src_namespace)"`
	ScrDevName     string `json:"srcDevName" orm:"column(src_dev_name)"`
	SrcBrowse      string `json:"srcBrowseName" orm:"column(src_browse_name)"`
	SrcNodeid      string `json:"srcNodeId" orm:"column(src_node_id)"`
	SrcUnit        string `json:"srcUnit" orm:"column(src_unit)"`
	DstBrowse      string `json:"dstBrowseName" orm:"column(dst_browse_name)"`
	DstNodeid      string `json:"dstNodeId" orm:"column(dst_node_id)"`
	ConvFunc       string `json:"convertFunc" orm:"column(convert_func)"`
	RefBrowseName1 string `json:"referenceBrowseName1" orm:"column(reference_browse_name_1)"`
	RefBrowseName2 string `json:"referenceBrowseName2" orm:"column(reference_browse_name_2)"`
	RefBrowseName3 string `json:"referenceBrowseName3" orm:"column(reference_browse_name_3)"`
	Value          string `json:"value" orm:"column(value)"`
	Modify         int    `json:"modify" orm:"column(modify)"`
}

// GetLevel 回傳空字串表示無level object
func (c *Converter) GetLevel() string {
	strArr := strings.Split(c.DstNodeid, ".")
	for _, v := range strArr {
		if strings.Contains(v, "LEVEL") {
			return v
		}
	}
	return ""
}

// EditConverter 編輯轉換表資料
type EditConverter struct {
	Name        string      `json:"name"`
	EditContent []Converter `json:"editContent"`
}

// Template 預設節點欄位
type Template struct {
	ID           int64  `json:"id" orm:"column(id)"`
	Name         string `json:"name" orm:"column(name);unique"`
	Description  string `json:"description" orm:"column(description)"`
	CreateTime   int64  `json:"createTime" orm:"column(create_time)"`
	ModifyTime   int64  `json:"modifyTime" orm:"column(modify_time)"`
	SystemCreate bool   `json:"-" orm:"column(sys_create)"`
	Model        string `json:"model" orm:"column(model)"`
}

// EditTemplate 編輯節點欄位
type EditTemplate struct {
	ID          int64    `json:"id"`
	EditContent Template `json:"editContent"`
}

// ErrMsg 錯誤訊息
type ErrMsg struct {
	Msg string `json:"msg"`
}

// DevInfo 設備資料
type DevInfo struct {
	ID       int64  `json:"id" orm:"column(id)"`
	Name     string `json:"name"`
	ConID    int64  `json:"conId" orm:"column(connect_id)"`
	ConName  string `json:"conName" orm:"column(connect_name)"`
	TempID   int64  `json:"tempID" orm:"column(template_id)"`
	TempName string `json:"tempName" orm:"column(template_name)"`
	Brand    string `json:"brand" orm:"column(brand)"`
	Protocol int    `json:"protocol" orm:"column(protocol)"`
	Status   int    `json:"status" orm:"column(status)"`
	Mac      string `json:"-" orm:"column(mac);unique"`
	Auth     string `json:"-" orm:"column(basic_auth)"`
}

// EditDeviceInfo 編輯Edge連線資料
type EditDeviceInfo struct {
	ID          int64   `json:"id"`
	EditContent DevInfo `json:"editContent"`
}

// ConInfo 連線資料
type ConInfo struct {
	ID            int64  `json:"id" orm:"column(id)"`
	IP            string `json:"ip" orm:"column(ip)"`
	Port          string `json:"port" orm:"column(port)"`
	Name          string `json:"name" orm:"column(name);unique"`
	Protocol      int    `json:"protocol" orm:"column(protocol)"`
	Status        int    `json:"status" orm:"column(status)"`
	Account       string `json:"account" orm:"column(account)"`
	Password      string `json:"password" orm:"column(password)"`
	Certification string `json:"certification" orm:"column(certification)"`
	Timestamp     int64  `json:"timestamp" orm:"column(timestamp)"`
	Exist         int    `json:"exist" orm:"column(exist)"`
}

// EditConInfo Edge連線資料
type EditConInfo struct {
	ID          int64   `json:"id"`
	EditContent ConInfo `json:"editContent"`
}

// ConnectRecord 系統記錄資料
type ConnectRecord struct {
	ID        int64  `json:"id" orm:"column(id)"`
	IP        string `json:"ip" orm:"column(ip)"`
	Type      int    `json:"type" orm:"column(type)"`
	URL       string `json:"url" orm:"column(url)"`
	Account   string `json:"account" orm:"column(account)"`
	Timestamp int64  `json:"timestamp" orm:"column(timestamp)"`
}

// TableName 系統記錄表名
func (c *ConnectRecord) TableName() string {
	return connectRecordTable
}

// TableName 連線資料表表名
func (c *ConInfo) TableName() string {
	return devAuthTable
}

// TableName 節點資料表名
func (c *NodeStruct) TableName() string {
	return nodeStructTable
}

// TableName 範本表名
func (c *Template) TableName() string {
	return tempTable
}

// TableName 關聯表表名
func (c *TemplateRel) TableName() string {
	return tempRelTable
}

// TableName 轉換表表名
func (c *Converter) TableName() string {
	return converterTable
}

// TableName 設備列表表名
func (c *DevInfo) TableName() string {
	return devInfoTable
}

// TableName 帳戶列表表名
func (c *EdgeAuth) TableName() string {
	return edgeAuthTable
}

// TableName 帳戶權限表名
func (c *AuthPermissopn) TableName() string {
	return authPermissionTable
}

// TableName 系統設定表名
func (c *SystemSetting) TableName() string {
	return systemSettingTable
}

// HistoryData 歷史資料
type HistoryData struct {
	ACMTIoMCPVersion                    string  `json:"ACMT_IoMCP_Version" influx:"ACMT_IoMCP_Version"`
	MachineBrand                        string  `json:"Machine_Brand" influx:"Machine_Brand"`
	MachineType                         string  `json:"Machine_Type" influx:"Machine_Type"`
	MachineSerialNumber                 string  `json:"Machine_Serial_Number" influx:"Machine_Serial_Number"`
	MachineManufactureDate              string  `json:"Machine_Manufacture_Date" influx:"Machine_Manufacture_Date"`
	MachinePowerSource                  string  `json:"Machine_Power_Source" influx:"Machine_Power_Source"`
	MachinePowerCapacity                string  `json:"Machine_Power_Capacity" influx:"Machine_Power_Capacity"`
	MachineOperateOilCapacity           string  `json:"Machine_Operate_Oil_Capacity" influx:"Machine_Operate_Oil_Capacity"`
	MachineLength                       string  `json:"Machine_Length" influx:"Machine_Length"`
	MachineWidth                        string  `json:"Machine_Width" influx:"Machine_Width"`
	MachineHeight                       string  `json:"Machine_Height" influx:"Machine_Height"`
	MachineNetWeight                    string  `json:"Machine_Net_Weight" influx:"Machine_Net_Weight"`
	MoldThicknessMin                    string  `json:"Mold_Thickness_Min" influx:"Mold_Thickness_Min"`
	MoldThicknessMax                    string  `json:"Mold_Thickness_Max" influx:"Mold_Thickness_Max"`
	TieBarWidth                         string  `json:"Tie_Bar_Width" influx:"Tie_Bar_Width"`
	TieBarHeight                        string  `json:"Tie_Bar_Height" influx:"Tie_Bar_Height"`
	PlatenHeight                        string  `json:"Platen_Height" influx:"Platen_Height"`
	PlatenDepth                         string  `json:"Platen_Depth" influx:"Platen_Depth"`
	MaxClampStroke                      string  `json:"Max_Clamp_Stroke" influx:"Max_Clamp_Stroke"`
	ClampForce                          string  `json:"Clamp_Force" influx:"Clamp_Force"`
	EjectorStroke                       string  `json:"Ejector_Stroke" influx:"Ejector_Stroke"`
	EjectorForce                        string  `json:"Ejector_Force" influx:"Ejector_Force"`
	BarrelCount                         string  `json:"Barrel_Count" influx:"Barrel_Count"`
	Barrel1ScrewStroke                  string  `json:"Barrel-1_Screw_Stroke" influx:"Barrel-1_Screw_Stroke"`
	Barrel1InjectSpeedMax               string  `json:"Barrel-1_Inject_Speed_Max" influx:"Barrel-1_Inject_Speed_Max"`
	Barrel1PlasticatingRate             string  `json:"Barrel-1_Plasticating_Rate" influx:"Barrel-1_Plasticating_Rate"`
	Barrel1HeatCapacity                 string  `json:"Barrel-1_Heat_Capacity" influx:"Barrel-1_Heat_Capacity"`
	Barrel1ScrewDiameter                string  `json:"Barrel-1_Screw_Diameter" influx:"Barrel-1_Screw_Diameter"`
	Barrel1MaxInjectionPressure         string  `json:"Barrel-1_Max_Injection_Pressure" influx:"Barrel-1_Max_Injection_Pressure"`
	Barrel1MaxHoldPressure              string  `json:"Barrel-1_Max_Hold_Pressure " influx:"Barrel-1_Max_Hold_Pressure "`
	Barrel1ScrewSpeed                   string  `json:"Barrel-1_Screw_Speed" influx:"Barrel-1_Screw_Speed"`
	Barrel1NozzleSealingForce           string  `json:"Barrel-1_Nozzle_Sealing_Force" influx:"Barrel-1_Nozzle_Sealing_Force"`
	Barrel2ScrewStroke                  string  `json:"Barrel-2_Screw_Stroke" influx:"Barrel-2_Screw_Stroke"`
	Barrel2InjectSpeedMax               string  `json:"Barrel-2_Inject_Speed_Max" influx:"Barrel-2_Inject_Speed_Max"`
	Barrel2PlasticatingRate             string  `json:"Barrel-2_Plasticating_Rate" influx:"Barrel-2_Plasticating_Rate"`
	Barrel2HeatCapacity                 string  `json:"Barrel-2_Heat_Capacity" influx:"Barrel-2_Heat_Capacity"`
	Barrel2ScrewDiameter                string  `json:"Barrel-2_Screw_Diameter" influx:"Barrel-2_Screw_Diameter"`
	Barrel2MaxInjectionPressure         string  `json:"Barrel-2_Max_Injection_Pressure" influx:"Barrel-2_Max_Injection_Pressure"`
	Barrel2MaxHoldPressure              string  `json:"Barrel-2_Max_Hold_Pressure" influx:"Barrel-2_Max_Hold_Pressure"`
	Barrel2ScrewSpeed                   string  `json:"Barrel-2_Screw_Speed" influx:"Barrel-2_Screw_Speed"`
	Barrel2NozzleSealingForce           string  `json:"Barrel-2_Nozzle_Sealing_Force" influx:"Barrel-2_Nozzle_Sealing_Force"`
	Barrel3ScrewStroke                  string  `json:"Barrel-3_Screw_Stroke" influx:"Barrel-3_Screw_Stroke"`
	Barrel3InjectSpeedMax               string  `json:"Barrel-3_Inject_Speed_Max" influx:"Barrel-3_Inject_Speed_Max"`
	Barrel3PlasticatingRate             string  `json:"Barrel-3_Plasticating_Rate" influx:"Barrel-3_Plasticating_Rate"`
	Barrel3HeatCapacity                 string  `json:"Barrel-3_Heat_Capacity" influx:"Barrel-3_Heat_Capacity"`
	Barrel3ScrewDiameter                string  `json:"Barrel-3_Screw_Diameter" influx:"Barrel-3_Screw_Diameter"`
	Barrel3MaxInjectionPressure         string  `json:"Barrel-3_Max_Injection_Pressure" influx:"Barrel-3_Max_Injection_Pressure"`
	Barrel3MaxHoldPressure              string  `json:"Barrel-3_Max_Hold_Pressure" influx:"Barrel-3_Max_Hold_Pressure"`
	Barrel3ScrewSpeed                   string  `json:"Barrel-3_Screw_Speed" influx:"Barrel-3_Screw_Speed"`
	Barrel3NozzleSealingForce           string  `json:"Barrel-3_Nozzle_Sealing_Force" influx:"Barrel-3_Nozzle_Sealing_Force"`
	UnitSelectSpeed                     string  `json:"Unit_Select_Speed" influx:"Unit_Select_Speed"`
	UnitSelectPressure                  string  `json:"Unit_Select_Pressure" influx:"Unit_Select_Pressure"`
	UnitSelectPosition                  string  `json:"Unit_Select_Position" influx:"Unit_Select_Position"`
	UnitSelectTemperature               string  `json:"Unit_Select_Temperature" influx:"Unit_Select_Temperature"`
	UnitSelectClampForce                string  `json:"Unit_Select_Clamp_Force" influx:"Unit_Select_Clamp_Force"`
	UnitSelectWeight                    string  `json:"Unit_Select_Weight" influx:"Unit_Select_Weight"`
	UnitSelectTime                      string  `json:"Unit_Select_Time" influx:"Unit_Select_Time"`
	MachineStatus                       int     `json:"Machine_Status" influx:"Machine_Status"`
	AlarmStatus                         string  `json:"Alarm_Status" influx:"Alarm_Status"`
	MachineMotor                        string  `json:"Machine_Motor" influx:"Machine_Motor"`
	MachineHeater                       string  `json:"Machine_Heater" influx:"Machine_Heater"`
	SaftDoorOperationSide               string  `json:"Saft_Door_Operation_Side" influx:"Saft_Door_Operation_Side"`
	SaftDoorNonOperationSide            string  `json:"Saft_Door_NonOperation_Side" influx:"Saft_Door_NonOperation_Side"`
	MoldCount                           int     `json:"Mold_Count" influx:"Mold_Count"`
	CycleTime                           float64 `json:"Cycle_Time" influx:"Cycle_Time"`
	MoldOpenTime                        string  `json:"MoldOpen_Time" influx:"MoldOpen_Time"`
	MoldCloseTime                       string  `json:"MoldClose_Time" influx:"MoldClose_Time"`
	MoldNo                              string  `json:"Mold_No" influx:"Mold_No"`
	Machinemoldedpieocessumup           string  `json:"Machine_molded_pieoces_sum_up" influx:"Machine_molded_pieoces_sum_up"`
	ProductionMoldname                  string  `json:"Production_Mold_name" influx:"Production_Mold_name"`
	Partsnumber                         string  `json:"Parts_number" influx:"Parts_number"`
	Resintype                           string  `json:"Resin_type(Product_Series)" influx:"Resin_type(Product_Series)"`
	Automoldingshotssumup               string  `json:"Auto_molding_shots_sum_up" influx:"Auto_molding_shots_sum_up"`
	Totallymoldedpieces                 string  `json:"Totally_molded_pieces" influx:"Totally_molded_pieces"`
	Powerontimesumup                    string  `json:"Power_on_time_sum_up" influx:"Power_on_time_sum_up"`
	Automoldingtimesumup                string  `json:"Auto_molding_time_sum_up" influx:"Auto_molding_time_sum_up"`
	TargetamountrequestedSet            string  `json:"Target_amount_requested_Set" influx:"Target_amount_requested_Set"`
	TargetamountrequestedActual         string  `json:"Target_amount_requested_Actual" influx:"Target_amount_requested_Actual"`
	Finishingratio                      string  `json:"Finishing_ratio" influx:"Finishing_ratio"`
	TotalgoodcyclesproducedSet          string  `json:"Total_good_cycles_produced_Set" influx:"Total_good_cycles_produced_Set"`
	TotalgoodcyclesproducedActual       string  `json:"Total_good_cycles_produced_Actual" influx:"Total_good_cycles_produced_Actual"`
	Goodratio                           string  `json:"Good_ratio" influx:"Good_ratio"`
	DefectivetotalproduceSet            string  `json:"Defective_total_produce_Set" influx:"Defective_total_produce_Set"`
	DefectivetotalproduceActual         string  `json:"Defective_total_produce_Actual" influx:"Defective_total_produce_Actual"`
	Badratio                            string  `json:"Bad_ratio" influx:"Bad_ratio"`
	Estimatedtimetocompleteproduct      string  `json:"Estimated_time_to_complete_product" influx:"Estimated_time_to_complete_product"`
	MachineDate                         string  `json:"Machine_Date" influx:"Machine_Date"`
	MachineTime                         string  `json:"Machine_Time" influx:"Machine_Time"`
	MachineWeek                         string  `json:"Machine_Week" influx:"Machine_Week"`
	EveryMonth500HrDatereset            string  `json:"Every_Month/500Hr_Date_reset" influx:"Every_Month/500Hr_Date_reset"`
	Every3Month1500HrDatereset          string  `json:"Every_3_Month/1500Hr_Date_reset" influx:"Every_3_Month/1500Hr_Date_reset"`
	Every6Month500HrDatereset           string  `json:"Every_6_Month/500Hr_Date_reset" influx:"Every_6_Month/500Hr_Date_reset"`
	Every1YearDatereset                 string  `json:"Every_1_Year_Date_reset" influx:"Every_1_Year_Date_reset"`
	Barrel1InjectStageSet               string  `json:"Barrel-1_Inject_Stage_Set" influx:"Barrel-1_Inject_Stage_Set"`
	Barrel1Inject1SpeedSet              string  `json:"Barrel-1_Inject_1_Speed_Set" influx:"Barrel-1_Inject_1_Speed_Set"`
	Barrel1Inject2SpeedSet              string  `json:"Barrel-1_Inject_2_Speed_Set" influx:"Barrel-1_Inject_2_Speed_Set"`
	Barrel1Inject3SpeedSet              string  `json:"Barrel-1_Inject_3_Speed_Set" influx:"Barrel-1_Inject_3_Speed_Set"`
	Barrel1Inject4SpeedSet              string  `json:"Barrel-1_Inject_4_Speed_Set" influx:"Barrel-1_Inject_4_Speed_Set"`
	Barrel1Inject5SpeedSet              string  `json:"Barrel-1_Inject_5_Speed_Set" influx:"Barrel-1_Inject_5_Speed_Set"`
	Barrel1Inject6SpeedSet              string  `json:"Barrel-1_Inject_6_Speed_Set" influx:"Barrel-1_Inject_6_Speed_Set"`
	Barrel1Inject7SpeedSet              string  `json:"Barrel-1_Inject_7_Speed_Set" influx:"Barrel-1_Inject_7_Speed_Set"`
	Barrel1Inject8SpeedSet              string  `json:"Barrel-1_Inject_8_Speed_Set" influx:"Barrel-1_Inject_8_Speed_Set"`
	Barrel1Inject9SpeedSet              string  `json:"Barrel-1_Inject_9_Speed_Set" influx:"Barrel-1_Inject_9_Speed_Set"`
	Barrel1Inject10SpeedSet             string  `json:"Barrel-1_Inject_10_Speed_Set" influx:"Barrel-1_Inject_10_Speed_Set"`
	Barrel1Inject1PositionSet           string  `json:"Barrel-1_Inject_1_Position_Set" influx:"Barrel-1_Inject_1_Position_Set"`
	Barrel1Inject2PositionSet           string  `json:"Barrel-1_Inject_2_Position_Set" influx:"Barrel-1_Inject_2_Position_Set"`
	Barrel1Inject3PositionSet           string  `json:"Barrel-1_Inject_3_Position_Set" influx:"Barrel-1_Inject_3_Position_Set"`
	Barrel1Inject4PositionSet           string  `json:"Barrel-1_Inject_4_Position_Set" influx:"Barrel-1_Inject_4_Position_Set"`
	Barrel1Inject5PositionSet           string  `json:"Barrel-1_Inject_5_Position_Set" influx:"Barrel-1_Inject_5_Position_Set"`
	Barrel1Inject6PositionSet           string  `json:"Barrel-1_Inject_6_Position_Set" influx:"Barrel-1_Inject_6_Position_Set"`
	Barrel1Inject7PositionSet           string  `json:"Barrel-1_Inject_7_Position_Set" influx:"Barrel-1_Inject_7_Position_Set"`
	Barrel1Inject8PositionSet           string  `json:"Barrel-1_Inject_8_Position_Set" influx:"Barrel-1_Inject_8_Position_Set"`
	Barrel1Inject9PositionSet           string  `json:"Barrel-1_Inject_9_Position_Set" influx:"Barrel-1_Inject_9_Position_Set"`
	Barrel1Inject10PositionSet          string  `json:"Barrel-1_Inject_10_Position_Set" influx:"Barrel-1_Inject_10_Position_Set"`
	Barrel1Inject1PressureSet           string  `json:"Barrel-1_Inject_1_Pressure_Set" influx:"Barrel-1_Inject_1_Pressure_Set"`
	Barrel1Inject2PressureSet           string  `json:"Barrel-1_Inject_2_Pressure_Set" influx:"Barrel-1_Inject_2_Pressure_Set"`
	Barrel1Inject3PressureSet           string  `json:"Barrel-1_Inject_3_Pressure_Set" influx:"Barrel-1_Inject_3_Pressure_Set"`
	Barrel1Inject4PressureSet           string  `json:"Barrel-1_Inject_4_Pressure_Set" influx:"Barrel-1_Inject_4_Pressure_Set"`
	Barrel1Inject5PressureSet           string  `json:"Barrel-1_Inject_5_Pressure_Set" influx:"Barrel-1_Inject_5_Pressure_Set"`
	Barrel1Inject6PressureSet           string  `json:"Barrel-1_Inject_6_Pressure_Set" influx:"Barrel-1_Inject_6_Pressure_Set"`
	Barrel1Inject7PressureSet           string  `json:"Barrel-1_Inject_7_Pressure_Set" influx:"Barrel-1_Inject_7_Pressure_Set"`
	Barrel1Inject8PressureSet           string  `json:"Barrel-1_Inject_8_Pressure_Set" influx:"Barrel-1_Inject_8_Pressure_Set"`
	Barrel1Inject9PressureSet           string  `json:"Barrel-1_Inject_9_Pressure_Set" influx:"Barrel-1_Inject_9_Pressure_Set"`
	Barrel1Inject10PressureSet          string  `json:"Barrel-1_Inject_10_Pressure_Set" influx:"Barrel-1_Inject_10_Pressure_Set"`
	Barrel1VPInjectTransferHoldSet      string  `json:"Barrel-1_V-P_Inject_Transfer_Hold_Set" influx:"Barrel-1_V-P_Inject_Transfer_Hold_Set"`
	Barrel1InjectTimeSet                string  `json:"Barrel-1_Inject_Time_Set" influx:"Barrel-1_Inject_Time_Set"`
	Barrel1VPPositionSet                string  `json:"Barrel-1_V-P_Position_Set" influx:"Barrel-1_V-P_Position_Set"`
	Barrel1VPTransferPressureSet        string  `json:"Barrel-1_V-P_Transfer_Pressure_Set" influx:"Barrel-1_V-P_Transfer_Pressure_Set"`
	Barrel1HoldStageSet                 string  `json:"Barrel-1_Hold_Stage_Set" influx:"Barrel-1_Hold_Stage_Set"`
	Barrel1Hold1TimeSet                 string  `json:"Barrel-1_Hold_1_Time_Set" influx:"Barrel-1_Hold_1_Time_Set"`
	Barrel1Hold2TimeSet                 string  `json:"Barrel-1_Hold_2_Time_Set" influx:"Barrel-1_Hold_2_Time_Set"`
	Barrel1Hold3TimeSet                 string  `json:"Barrel-1_Hold_3_Time_Set" influx:"Barrel-1_Hold_3_Time_Set"`
	Barrel1Hold4TimeSet                 string  `json:"Barrel-1_Hold_4_Time_Set" influx:"Barrel-1_Hold_4_Time_Set"`
	Barrel1Hold5TimeSet                 string  `json:"Barrel-1_Hold_5_Time_Set" influx:"Barrel-1_Hold_5_Time_Set"`
	Barrel1Hold6TimeSet                 string  `json:"Barrel-1_Hold_6_Time_Set" influx:"Barrel-1_Hold_6_Time_Set"`
	Barrel1Hold7TimeSet                 string  `json:"Barrel-1_Hold_7_Time_Set" influx:"Barrel-1_Hold_7_Time_Set"`
	Barrel1Hold8TimeSet                 string  `json:"Barrel-1_Hold_8_Time_Set" influx:"Barrel-1_Hold_8_Time_Set"`
	Barrel1Hold9TimeSet                 string  `json:"Barrel-1_Hold_9_Time_Set" influx:"Barrel-1_Hold_9_Time_Set"`
	Barrel1Hold10TimeSet                string  `json:"Barrel-1_Hold_10_Time_Set" influx:"Barrel-1_Hold_10_Time_Set"`
	Barrel1Hold1SpeedSet                string  `json:"Barrel-1_Hold_1_Speed_Set" influx:"Barrel-1_Hold_1_Speed_Set"`
	Barrel1Hold2SpeedSet                string  `json:"Barrel-1_Hold_2_Speed_Set" influx:"Barrel-1_Hold_2_Speed_Set"`
	Barrel1Hold3SpeedSet                string  `json:"Barrel-1_Hold_3_Speed_Set" influx:"Barrel-1_Hold_3_Speed_Set"`
	Barrel1Hold4SpeedSet                string  `json:"Barrel-1_Hold_4_Speed_Set" influx:"Barrel-1_Hold_4_Speed_Set"`
	Barrel1Hold5SpeedSet                string  `json:"Barrel-1_Hold_5_Speed_Set" influx:"Barrel-1_Hold_5_Speed_Set"`
	Barrel1Hold6SpeedSet                string  `json:"Barrel-1_Hold_6_Speed_Set" influx:"Barrel-1_Hold_6_Speed_Set"`
	Barrel1Hold7SpeedSet                string  `json:"Barrel-1_Hold_7_Speed_Set" influx:"Barrel-1_Hold_7_Speed_Set"`
	Barrel1Hold8SpeedSet                string  `json:"Barrel-1_Hold_8_Speed_Set" influx:"Barrel-1_Hold_8_Speed_Set"`
	Barrel1Hold9SpeedSet                string  `json:"Barrel-1_Hold_9_Speed_Set" influx:"Barrel-1_Hold_9_Speed_Set"`
	Barrel1Hold10SpeedSet               string  `json:"Barrel-1_Hold_10_Speed_Set" influx:"Barrel-1_Hold_10_Speed_Set"`
	Barrel1Hold1PressureSet             string  `json:"Barrel-1_Hold_1_Pressure_Set" influx:"Barrel-1_Hold_1_Pressure_Set"`
	Barrel1Hold2PressureSet             string  `json:"Barrel-1_Hold_2_Pressure_Set" influx:"Barrel-1_Hold_2_Pressure_Set"`
	Barrel1Hold3PressureSet             string  `json:"Barrel-1_Hold_3_Pressure_Set" influx:"Barrel-1_Hold_3_Pressure_Set"`
	Barrel1Hold4PressureSet             string  `json:"Barrel-1_Hold_4_Pressure_Set" influx:"Barrel-1_Hold_4_Pressure_Set"`
	Barrel1Hold5PressureSet             string  `json:"Barrel-1_Hold_5_Pressure_Set" influx:"Barrel-1_Hold_5_Pressure_Set"`
	Barrel1Hold6PressureSet             string  `json:"Barrel-1_Hold_6_Pressure_Set" influx:"Barrel-1_Hold_6_Pressure_Set"`
	Barrel1Hold7PressureSet             string  `json:"Barrel-1_Hold_7_Pressure_Set" influx:"Barrel-1_Hold_7_Pressure_Set"`
	Barrel1Hold8PressureSet             string  `json:"Barrel-1_Hold_8_Pressure_Set" influx:"Barrel-1_Hold_8_Pressure_Set"`
	Barrel1Hold9PressureSet             string  `json:"Barrel-1_Hold_9_Pressure_Set" influx:"Barrel-1_Hold_9_Pressure_Set"`
	Barrel1Hold10PressureSet            string  `json:"Barrel-1_Hold_10_Pressure_Set" influx:"Barrel-1_Hold_10_Pressure_Set"`
	Barrel1TempStageSet                 string  `json:"Barrel-1_Temp_Stage_Set" influx:"Barrel-1_Temp_Stage_Set"`
	Barrel1NozzTemp1Set                 string  `json:"Barrel-1_NozzTemp_1_Set" influx:"Barrel-1_NozzTemp_1_Set"`
	Barrel1NozzTemp2Set                 string  `json:"Barrel-1_NozzTemp_2_Set" influx:"Barrel-1_NozzTemp_2_Set"`
	Barrel1Temp1stSet                   string  `json:"Barrel-1_Temp_1st_Set" influx:"Barrel-1_Temp_1st_Set"`
	Barrel1Temp2ndSet                   string  `json:"Barrel-1_Temp_2nd_Set" influx:"Barrel-1_Temp_2nd_Set"`
	Barrel1Temp3rdSet                   string  `json:"Barrel-1_Temp_3rd_Set" influx:"Barrel-1_Temp_3rd_Set"`
	Barrel1Temp4thSet                   string  `json:"Barrel-1_Temp_4th_Set" influx:"Barrel-1_Temp_4th_Set"`
	Barrel1Temp5thSet                   string  `json:"Barrel-1_Temp_5th_Set" influx:"Barrel-1_Temp_5th_Set"`
	Barrel1Temp6thSet                   string  `json:"Barrel-1_Temp_6th_Set" influx:"Barrel-1_Temp_6th_Set"`
	Barrel1Temp7thSet                   string  `json:"Barrel-1_Temp_7th_Set" influx:"Barrel-1_Temp_7th_Set"`
	Barrel1Temp8thSet                   string  `json:"Barrel-1_Temp_8th_Set" influx:"Barrel-1_Temp_8th_Set"`
	Barrel1TempfeedSet                  string  `json:"Barrel-1_Temp_feed_Set" influx:"Barrel-1_Temp_feed_Set"`
	Barrel1ChargeStageSet               string  `json:"Barrel-1_Charge_Stage_Set" influx:"Barrel-1_Charge_Stage_Set"`
	Barrel1Charge1SpeedSet              string  `json:"Barrel-1_Charge_1_Speed_Set" influx:"Barrel-1_Charge_1_Speed_Set"`
	Barrel1Charge2SpeedSet              string  `json:"Barrel-1_Charge_2_Speed_Set" influx:"Barrel-1_Charge_2_Speed_Set"`
	Barrel1Charge3SpeedSet              string  `json:"Barrel-1_Charge_3_Speed_Set" influx:"Barrel-1_Charge_3_Speed_Set"`
	Barrel1Charge4SpeedSet              string  `json:"Barrel-1_Charge_4_Speed_Set" influx:"Barrel-1_Charge_4_Speed_Set"`
	Barrel1Charge5SpeedSet              string  `json:"Barrel-1_Charge_5_Speed_Set" influx:"Barrel-1_Charge_5_Speed_Set"`
	Barrel1Charge6SpeedSet              string  `json:"Barrel-1_Charge_6_Speed_Set" influx:"Barrel-1_Charge_6_Speed_Set"`
	Barrel1Charge7SpeedSet              string  `json:"Barrel-1_Charge_7_Speed_Set" influx:"Barrel-1_Charge_7_Speed_Set"`
	Barrel1Charge8SpeedSet              string  `json:"Barrel-1_Charge_8_Speed_Set" influx:"Barrel-1_Charge_8_Speed_Set"`
	Barrel1Charge9SpeedSet              string  `json:"Barrel-1_Charge_9_Speed_Set" influx:"Barrel-1_Charge_9_Speed_Set"`
	Barrel1Charge10SpeedSet             string  `json:"Barrel-1_Charge_10_Speed_Set" influx:"Barrel-1_Charge_10_Speed_Set"`
	Barrel1Charge1PositionSet           string  `json:"Barrel-1_Charge_1_Position_Set" influx:"Barrel-1_Charge_1_Position_Set"`
	Barrel1Charge2PositionSet           string  `json:"Barrel-1_Charge_2_Position_Set" influx:"Barrel-1_Charge_2_Position_Set"`
	Barrel1Charge3PositionSet           string  `json:"Barrel-1_Charge_3_Position_Set" influx:"Barrel-1_Charge_3_Position_Set"`
	Barrel1Charge4PositionSet           string  `json:"Barrel-1_Charge_4_Position_Set" influx:"Barrel-1_Charge_4_Position_Set"`
	Barrel1Charge5PositionSet           string  `json:"Barrel-1_Charge_5_Position_Set" influx:"Barrel-1_Charge_5_Position_Set"`
	Barrel1Charge6PositionSet           string  `json:"Barrel-1_Charge_6_Position_Set" influx:"Barrel-1_Charge_6_Position_Set"`
	Barrel1Charge7PositionSet           string  `json:"Barrel-1_Charge_7_Position_Set" influx:"Barrel-1_Charge_7_Position_Set"`
	Barrel1Charge8PositionSet           string  `json:"Barrel-1_Charge_8_Position_Set" influx:"Barrel-1_Charge_8_Position_Set"`
	Barrel1Charge9PositionSet           string  `json:"Barrel-1_Charge_9_Position_Set" influx:"Barrel-1_Charge_9_Position_Set"`
	Barrel1Charge10PositionSet          string  `json:"Barrel-1_Charge_10_Position_Set" influx:"Barrel-1_Charge_10_Position_Set"`
	Barrel1Charge1BackPressureSet       string  `json:"Barrel-1_Charge_1_BackPressure_Set" influx:"Barrel-1_Charge_1_BackPressure_Set"`
	Barrel1Charge2BackPressureSet       string  `json:"Barrel-1_Charge_2_BackPressure_Set" influx:"Barrel-1_Charge_2_BackPressure_Set"`
	Barrel1Charge3BackPressureSet       string  `json:"Barrel-1_Charge_3_BackPressure_Set" influx:"Barrel-1_Charge_3_BackPressure_Set"`
	Barrel1Charge4BackPressureSet       string  `json:"Barrel-1_Charge_4_BackPressure_Set" influx:"Barrel-1_Charge_4_BackPressure_Set"`
	Barrel1Charge5BackPressureSet       string  `json:"Barrel-1_Charge_5_BackPressure_Set" influx:"Barrel-1_Charge_5_BackPressure_Set"`
	Barrel1Charge6BackPressureSet       string  `json:"Barrel-1_Charge_6_BackPressure_Set" influx:"Barrel-1_Charge_6_BackPressure_Set"`
	Barrel1Charge7BackPressureSet       string  `json:"Barrel-1_Charge_7_BackPressure_Set" influx:"Barrel-1_Charge_7_BackPressure_Set"`
	Barrel1Charge8BackPressureSet       string  `json:"Barrel-1_Charge_8_BackPressure_Set" influx:"Barrel-1_Charge_8_BackPressure_Set"`
	Barrel1Charge9BackPressureSet       string  `json:"Barrel-1_Charge_9_BackPressure_Set" influx:"Barrel-1_Charge_9_BackPressure_Set"`
	Barrel1Charge10BackPressureSet      string  `json:"Barrel-1_Charge_10_BackPressure_Set" influx:"Barrel-1_Charge_10_BackPressure_Set"`
	Barrel1ChargeProtectTimeSet         string  `json:"Barrel-1_Charge_Protect_Time_Set" influx:"Barrel-1_Charge_Protect_Time_Set"`
	Barrel1MaxProtectTimeSet            string  `json:"Barrel-1_Max_Protect_Time_Set" influx:"Barrel-1_Max_Protect_Time_Set"`
	Barrel1SuckModeSet                  string  `json:"Barrel-1_Suck_Mode_Set" influx:"Barrel-1_Suck_Mode_Set"`
	Barrel1PreSuckSpeedSet              string  `json:"Barrel-1_Pre_Suck_Speed_Set" influx:"Barrel-1_Pre_Suck_Speed_Set"`
	Barrel1PreSuckPositionSet           string  `json:"Barrel-1_Pre_Suck_Position_Set" influx:"Barrel-1_Pre_Suck_Position_Set"`
	Barrel1PreSuckPressureSet           string  `json:"Barrel-1_Pre_Suck_Pressure_Set" influx:"Barrel-1_Pre_Suck_Pressure_Set"`
	Barrel1BackSuckModeSet              string  `json:"Barrel-1_Back_Suck_Mode_Set" influx:"Barrel-1_Back_Suck_Mode_Set"`
	Barrel1BackSuckSpeedSet             string  `json:"Barrel-1_Back_Suck_Speed_Set" influx:"Barrel-1_Back_Suck_Speed_Set"`
	Barrel1BackSuckPositionSet          string  `json:"Barrel-1_Back_Suck_Position_Set" influx:"Barrel-1_Back_Suck_Position_Set"`
	Barrel1BackSuckPressureSet          string  `json:"Barrel-1_Back_Suck_Pressure_Set" influx:"Barrel-1_Back_Suck_Pressure_Set"`
	Barrel2InjectStageSet               string  `json:"Barrel-2_Inject_Stage_Set" influx:"Barrel-2_Inject_Stage_Set"`
	Barrel2Inject1SpeedSet              string  `json:"Barrel-2_Inject_1_Speed_Set" influx:"Barrel-2_Inject_1_Speed_Set"`
	Barrel2Inject2SpeedSet              string  `json:"Barrel-2_Inject_2_Speed_Set" influx:"Barrel-2_Inject_2_Speed_Set"`
	Barrel2Inject3SpeedSet              string  `json:"Barrel-2_Inject_3_Speed_Set" influx:"Barrel-2_Inject_3_Speed_Set"`
	Barrel2Inject4SpeedSet              string  `json:"Barrel-2_Inject_4_Speed_Set" influx:"Barrel-2_Inject_4_Speed_Set"`
	Barrel2Inject5SpeedSet              string  `json:"Barrel-2_Inject_5_Speed_Set" influx:"Barrel-2_Inject_5_Speed_Set"`
	Barrel2Inject6SpeedSet              string  `json:"Barrel-2_Inject_6_Speed_Set" influx:"Barrel-2_Inject_6_Speed_Set"`
	Barrel2Inject7SpeedSet              string  `json:"Barrel-2_Inject_7_Speed_Set" influx:"Barrel-2_Inject_7_Speed_Set"`
	Barrel2Inject8SpeedSet              string  `json:"Barrel-2_Inject_8_Speed_Set" influx:"Barrel-2_Inject_8_Speed_Set"`
	Barrel2Inject9SpeedSet              string  `json:"Barrel-2_Inject_9_Speed_Set" influx:"Barrel-2_Inject_9_Speed_Set"`
	Barrel2Inject10SpeedSet             string  `json:"Barrel-2_Inject_10_Speed_Set" influx:"Barrel-2_Inject_10_Speed_Set"`
	Barrel2Inject1PositionSet           string  `json:"Barrel-2_Inject_1_Position_Set" influx:"Barrel-2_Inject_1_Position_Set"`
	Barrel2Inject2PositionSet           string  `json:"Barrel-2_Inject_2_Position_Set" influx:"Barrel-2_Inject_2_Position_Set"`
	Barrel2Inject3PositionSet           string  `json:"Barrel-2_Inject_3_Position_Set" influx:"Barrel-2_Inject_3_Position_Set"`
	Barrel2Inject4PositionSet           string  `json:"Barrel-2_Inject_4_Position_Set" influx:"Barrel-2_Inject_4_Position_Set"`
	Barrel2Inject5PositionSet           string  `json:"Barrel-2_Inject_5_Position_Set" influx:"Barrel-2_Inject_5_Position_Set"`
	Barrel2Inject6PositionSet           string  `json:"Barrel-2_Inject_6_Position_Set" influx:"Barrel-2_Inject_6_Position_Set"`
	Barrel2Inject7PositionSet           string  `json:"Barrel-2_Inject_7_Position_Set" influx:"Barrel-2_Inject_7_Position_Set"`
	Barrel2Inject8PositionSet           string  `json:"Barrel-2_Inject_8_Position_Set" influx:"Barrel-2_Inject_8_Position_Set"`
	Barrel2Inject9PositionSet           string  `json:"Barrel-2_Inject_9_Position_Set" influx:"Barrel-2_Inject_9_Position_Set"`
	Barrel2Inject10PositionSet          string  `json:"Barrel-2_Inject_10_Position_Set" influx:"Barrel-2_Inject_10_Position_Set"`
	Barrel2Inject1PressureSet           string  `json:"Barrel-2_Inject_1_Pressure_Set" influx:"Barrel-2_Inject_1_Pressure_Set"`
	Barrel2Inject2PressureSet           string  `json:"Barrel-2_Inject_2_Pressure_Set" influx:"Barrel-2_Inject_2_Pressure_Set"`
	Barrel2Inject3PressureSet           string  `json:"Barrel-2_Inject_3_Pressure_Set" influx:"Barrel-2_Inject_3_Pressure_Set"`
	Barrel2Inject4PressureSet           string  `json:"Barrel-2_Inject_4_Pressure_Set" influx:"Barrel-2_Inject_4_Pressure_Set"`
	Barrel2Inject5PressureSet           string  `json:"Barrel-2_Inject_5_Pressure_Set" influx:"Barrel-2_Inject_5_Pressure_Set"`
	Barrel2Inject6PressureSet           string  `json:"Barrel-2_Inject_6_Pressure_Set" influx:"Barrel-2_Inject_6_Pressure_Set"`
	Barrel2Inject7PressureSet           string  `json:"Barrel-2_Inject_7_Pressure_Set" influx:"Barrel-2_Inject_7_Pressure_Set"`
	Barrel2Inject8PressureSet           string  `json:"Barrel-2_Inject_8_Pressure_Set" influx:"Barrel-2_Inject_8_Pressure_Set"`
	Barrel2Inject9PressureSet           string  `json:"Barrel-2_Inject_9_Pressure_Set" influx:"Barrel-2_Inject_9_Pressure_Set"`
	Barrel2Inject10PressureSet          string  `json:"Barrel-2_Inject_10_Pressure_Set" influx:"Barrel-2_Inject_10_Pressure_Set"`
	Barrel2VPInjectTransferHoldSet      string  `json:"Barrel-2_V-P_Inject_Transfer_Hold_Set" influx:"Barrel-2_V-P_Inject_Transfer_Hold_Set"`
	Barrel2InjectTimeSet                string  `json:"Barrel-2_Inject_Time_Set" influx:"Barrel-2_Inject_Time_Set"`
	Barrel2VPPositionSet                string  `json:"Barrel-2_V-P_Position_Set" influx:"Barrel-2_V-P_Position_Set"`
	Barrel2VPTransferPressureSet        string  `json:"Barrel-2_V-P_Transfer_Pressure_Set" influx:"Barrel-2_V-P_Transfer_Pressure_Set"`
	Barrel2HoldStageSet                 string  `json:"Barrel-2_Hold_Stage_Set" influx:"Barrel-2_Hold_Stage_Set"`
	Barrel2Hold1TimeSet                 string  `json:"Barrel-2_Hold_1_Time_Set" influx:"Barrel-2_Hold_1_Time_Set"`
	Barrel2Hold2TimeSet                 string  `json:"Barrel-2_Hold_2_Time_Set" influx:"Barrel-2_Hold_2_Time_Set"`
	Barrel2Hold3TimeSet                 string  `json:"Barrel-2_Hold_3_Time_Set" influx:"Barrel-2_Hold_3_Time_Set"`
	Barrel2Hold4TimeSet                 string  `json:"Barrel-2_Hold_4_Time_Set" influx:"Barrel-2_Hold_4_Time_Set"`
	Barrel2Hold5TimeSet                 string  `json:"Barrel-2_Hold_5_Time_Set" influx:"Barrel-2_Hold_5_Time_Set"`
	Barrel2Hold6TimeSet                 string  `json:"Barrel-2_Hold_6_Time_Set" influx:"Barrel-2_Hold_6_Time_Set"`
	Barrel2Hold7TimeSet                 string  `json:"Barrel-2_Hold_7_Time_Set" influx:"Barrel-2_Hold_7_Time_Set"`
	Barrel2Hold8TimeSet                 string  `json:"Barrel-2_Hold_8_Time_Set" influx:"Barrel-2_Hold_8_Time_Set"`
	Barrel2Hold9TimeSet                 string  `json:"Barrel-2_Hold_9_Time_Set" influx:"Barrel-2_Hold_9_Time_Set"`
	Barrel2Hold10TimeSet                string  `json:"Barrel-2_Hold_10_Time_Set" influx:"Barrel-2_Hold_10_Time_Set"`
	Barrel2Hold1SpeedSet                string  `json:"Barrel-2_Hold_1_Speed_Set" influx:"Barrel-2_Hold_1_Speed_Set"`
	Barrel2Hold2SpeedSet                string  `json:"Barrel-2_Hold_2_Speed_Set" influx:"Barrel-2_Hold_2_Speed_Set"`
	Barrel2Hold3SpeedSet                string  `json:"Barrel-2_Hold_3_Speed_Set" influx:"Barrel-2_Hold_3_Speed_Set"`
	Barrel2Hold4SpeedSet                string  `json:"Barrel-2_Hold_4_Speed_Set" influx:"Barrel-2_Hold_4_Speed_Set"`
	Barrel2Hold5SpeedSet                string  `json:"Barrel-2_Hold_5_Speed_Set" influx:"Barrel-2_Hold_5_Speed_Set"`
	Barrel2Hold6SpeedSet                string  `json:"Barrel-2_Hold_6_Speed_Set" influx:"Barrel-2_Hold_6_Speed_Set"`
	Barrel2Hold7SpeedSet                string  `json:"Barrel-2_Hold_7_Speed_Set" influx:"Barrel-2_Hold_7_Speed_Set"`
	Barrel2Hold8SpeedSet                string  `json:"Barrel-2_Hold_8_Speed_Set" influx:"Barrel-2_Hold_8_Speed_Set"`
	Barrel2Hold9SpeedSet                string  `json:"Barrel-2_Hold_9_Speed_Set" influx:"Barrel-2_Hold_9_Speed_Set"`
	Barrel2Hold10SpeedSet               string  `json:"Barrel-2_Hold_10_Speed_Set" influx:"Barrel-2_Hold_10_Speed_Set"`
	Barrel2Hold1PressureSet             string  `json:"Barrel-2_Hold_1_Pressure_Set" influx:"Barrel-2_Hold_1_Pressure_Set"`
	Barrel2Hold2PressureSet             string  `json:"Barrel-2_Hold_2_Pressure_Set" influx:"Barrel-2_Hold_2_Pressure_Set"`
	Barrel2Hold3PressureSet             string  `json:"Barrel-2_Hold_3_Pressure_Set" influx:"Barrel-2_Hold_3_Pressure_Set"`
	Barrel2Hold4PressureSet             string  `json:"Barrel-2_Hold_4_Pressure_Set" influx:"Barrel-2_Hold_4_Pressure_Set"`
	Barrel2Hold5PressureSet             string  `json:"Barrel-2_Hold_5_Pressure_Set" influx:"Barrel-2_Hold_5_Pressure_Set"`
	Barrel2Hold6PressureSet             string  `json:"Barrel-2_Hold_6_Pressure_Set" influx:"Barrel-2_Hold_6_Pressure_Set"`
	Barrel2Hold7PressureSet             string  `json:"Barrel-2_Hold_7_Pressure_Set" influx:"Barrel-2_Hold_7_Pressure_Set"`
	Barrel2Hold8PressureSet             string  `json:"Barrel-2_Hold_8_Pressure_Set" influx:"Barrel-2_Hold_8_Pressure_Set"`
	Barrel2Hold9PressureSet             string  `json:"Barrel-2_Hold_9_Pressure_Set" influx:"Barrel-2_Hold_9_Pressure_Set"`
	Barrel2Hold10PressureSet            string  `json:"Barrel-2_Hold_10_Pressure_Set" influx:"Barrel-2_Hold_10_Pressure_Set"`
	Barrel2TempStageSet                 string  `json:"Barrel-2_Temp_Stage_Set" influx:"Barrel-2_Temp_Stage_Set"`
	Barrel2NozzTemp1Set                 string  `json:"Barrel-2_NozzTemp_1_Set" influx:"Barrel-2_NozzTemp_1_Set"`
	Barrel2NozzTemp2Set                 string  `json:"Barrel-2_NozzTemp_2_Set" influx:"Barrel-2_NozzTemp_2_Set"`
	Barrel2Temp1stSet                   string  `json:"Barrel-2_Temp_1st_Set" influx:"Barrel-2_Temp_1st_Set"`
	Barrel2Temp2ndSet                   string  `json:"Barrel-2_Temp_2nd_Set" influx:"Barrel-2_Temp_2nd_Set"`
	Barrel2Temp3rdSet                   string  `json:"Barrel-2_Temp_3rd_Set" influx:"Barrel-2_Temp_3rd_Set"`
	Barrel2Temp4thSet                   string  `json:"Barrel-2_Temp_4th_Set" influx:"Barrel-2_Temp_4th_Set"`
	Barrel2Temp5thSet                   string  `json:"Barrel-2_Temp_5th_Set" influx:"Barrel-2_Temp_5th_Set"`
	Barrel2Temp6thSet                   string  `json:"Barrel-2_Temp_6th_Set" influx:"Barrel-2_Temp_6th_Set"`
	Barrel2Temp7thSet                   string  `json:"Barrel-2_Temp_7th_Set" influx:"Barrel-2_Temp_7th_Set"`
	Barrel2Temp8thSet                   string  `json:"Barrel-2_Temp_8th_Set" influx:"Barrel-2_Temp_8th_Set"`
	Barrel2TempfeedSet                  string  `json:"Barrel-2_Temp_feed_Set" influx:"Barrel-2_Temp_feed_Set"`
	Barrel2ChargeStageSet               string  `json:"Barrel-2_Charge_Stage_Set" influx:"Barrel-2_Charge_Stage_Set"`
	Barrel2Charge1SpeedSet              string  `json:"Barrel-2_Charge_1_Speed_Set" influx:"Barrel-2_Charge_1_Speed_Set"`
	Barrel2Charge2SpeedSet              string  `json:"Barrel-2_Charge_2_Speed_Set" influx:"Barrel-2_Charge_2_Speed_Set"`
	Barrel2Charge3SpeedSet              string  `json:"Barrel-2_Charge_3_Speed_Set" influx:"Barrel-2_Charge_3_Speed_Set"`
	Barrel2Charge4SpeedSet              string  `json:"Barrel-2_Charge_4_Speed_Set" influx:"Barrel-2_Charge_4_Speed_Set"`
	Barrel2Charge5SpeedSet              string  `json:"Barrel-2_Charge_5_Speed_Set" influx:"Barrel-2_Charge_5_Speed_Set"`
	Barrel2Charge6SpeedSet              string  `json:"Barrel-2_Charge_6_Speed_Set" influx:"Barrel-2_Charge_6_Speed_Set"`
	Barrel2Charge7SpeedSet              string  `json:"Barrel-2_Charge_7_Speed_Set" influx:"Barrel-2_Charge_7_Speed_Set"`
	Barrel2Charge8SpeedSet              string  `json:"Barrel-2_Charge_8_Speed_Set" influx:"Barrel-2_Charge_8_Speed_Set"`
	Barrel2Charge9SpeedSet              string  `json:"Barrel-2_Charge_9_Speed_Set" influx:"Barrel-2_Charge_9_Speed_Set"`
	Barrel2Charge10SpeedSet             string  `json:"Barrel-2_Charge_10_Speed_Set" influx:"Barrel-2_Charge_10_Speed_Set"`
	Barrel2Charge1PositionSet           string  `json:"Barrel-2_Charge_1_Position_Set" influx:"Barrel-2_Charge_1_Position_Set"`
	Barrel2Charge2PositionSet           string  `json:"Barrel-2_Charge_2_Position_Set" influx:"Barrel-2_Charge_2_Position_Set"`
	Barrel2Charge3PositionSet           string  `json:"Barrel-2_Charge_3_Position_Set" influx:"Barrel-2_Charge_3_Position_Set"`
	Barrel2Charge4PositionSet           string  `json:"Barrel-2_Charge_4_Position_Set" influx:"Barrel-2_Charge_4_Position_Set"`
	Barrel2Charge5PositionSet           string  `json:"Barrel-2_Charge_5_Position_Set" influx:"Barrel-2_Charge_5_Position_Set"`
	Barrel2Charge6PositionSet           string  `json:"Barrel-2_Charge_6_Position_Set" influx:"Barrel-2_Charge_6_Position_Set"`
	Barrel2Charge7PositionSet           string  `json:"Barrel-2_Charge_7_Position_Set" influx:"Barrel-2_Charge_7_Position_Set"`
	Barrel2Charge8PositionSet           string  `json:"Barrel-2_Charge_8_Position_Set" influx:"Barrel-2_Charge_8_Position_Set"`
	Barrel2Charge9PositionSet           string  `json:"Barrel-2_Charge_9_Position_Set" influx:"Barrel-2_Charge_9_Position_Set"`
	Barrel2Charge10PositionSet          string  `json:"Barrel-2_Charge_10_Position_Set" influx:"Barrel-2_Charge_10_Position_Set"`
	Barrel2Charge1BackPressureSet       string  `json:"Barrel-2_Charge_1_BackPressure_Set" influx:"Barrel-2_Charge_1_BackPressure_Set"`
	Barrel2Charge2BackPressureSet       string  `json:"Barrel-2_Charge_2_BackPressure_Set" influx:"Barrel-2_Charge_2_BackPressure_Set"`
	Barrel2Charge3BackPressureSet       string  `json:"Barrel-2_Charge_3_BackPressure_Set" influx:"Barrel-2_Charge_3_BackPressure_Set"`
	Barrel2Charge4BackPressureSet       string  `json:"Barrel-2_Charge_4_BackPressure_Set" influx:"Barrel-2_Charge_4_BackPressure_Set"`
	Barrel2Charge5BackPressureSet       string  `json:"Barrel-2_Charge_5_BackPressure_Set" influx:"Barrel-2_Charge_5_BackPressure_Set"`
	Barrel2Charge6BackPressureSet       string  `json:"Barrel-2_Charge_6_BackPressure_Set" influx:"Barrel-2_Charge_6_BackPressure_Set"`
	Barrel2Charge7BackPressureSet       string  `json:"Barrel-2_Charge_7_BackPressure_Set" influx:"Barrel-2_Charge_7_BackPressure_Set"`
	Barrel2Charge8BackPressureSet       string  `json:"Barrel-2_Charge_8_BackPressure_Set" influx:"Barrel-2_Charge_8_BackPressure_Set"`
	Barrel2Charge9BackPressureSet       string  `json:"Barrel-2_Charge_9_BackPressure_Set" influx:"Barrel-2_Charge_9_BackPressure_Set"`
	Barrel2Charge10BackPressureSet      string  `json:"Barrel-2_Charge_10_BackPressure_Set" influx:"Barrel-2_Charge_10_BackPressure_Set"`
	Barrel2ChargeProtectTimeSet         string  `json:"Barrel-2_Charge_Protect_Time_Set" influx:"Barrel-2_Charge_Protect_Time_Set"`
	Barrel2MaxProtectTimeSet            string  `json:"Barrel-2_Max_Protect_Time_Set" influx:"Barrel-2_Max_Protect_Time_Set"`
	Barrel2SuckModeSet                  string  `json:"Barrel-2_Suck_Mode_Set" influx:"Barrel-2_Suck_Mode_Set"`
	Barrel2PreSuckSpeedSet              string  `json:"Barrel-2_Pre_Suck_Speed_Set" influx:"Barrel-2_Pre_Suck_Speed_Set"`
	Barrel2PreSuckPositionSet           string  `json:"Barrel-2_Pre_Suck_Position_Set" influx:"Barrel-2_Pre_Suck_Position_Set"`
	Barrel2PreSuckPressureSet           string  `json:"Barrel-2_Pre_Suck_Pressure_Set" influx:"Barrel-2_Pre_Suck_Pressure_Set"`
	Barrel2BackSuckModeSet              string  `json:"Barrel-2_Back_Suck_Mode_Set" influx:"Barrel-2_Back_Suck_Mode_Set"`
	Barrel2BackSuckSpeedSet             string  `json:"Barrel-2_Back_Suck_Speed_Set" influx:"Barrel-2_Back_Suck_Speed_Set"`
	Barrel2BackSuckPositionSet          string  `json:"Barrel-2_Back_Suck_Position_Set" influx:"Barrel-2_Back_Suck_Position_Set"`
	Barrel2BackSuckPressureSet          string  `json:"Barrel-2_Back_Suck_Pressure_Set" influx:"Barrel-2_Back_Suck_Pressure_Set"`
	Barrel3InjectStageSet               string  `json:"Barrel-3_Inject_Stage_Set" influx:"Barrel-3_Inject_Stage_Set"`
	Barrel3Inject1SpeedSet              string  `json:"Barrel-3_Inject_1_Speed_Set" influx:"Barrel-3_Inject_1_Speed_Set"`
	Barrel3Inject2SpeedSet              string  `json:"Barrel-3_Inject_2_Speed_Set" influx:"Barrel-3_Inject_2_Speed_Set"`
	Barrel3Inject3SpeedSet              string  `json:"Barrel-3_Inject_3_Speed_Set" influx:"Barrel-3_Inject_3_Speed_Set"`
	Barrel3Inject4SpeedSet              string  `json:"Barrel-3_Inject_4_Speed_Set" influx:"Barrel-3_Inject_4_Speed_Set"`
	Barrel3Inject5SpeedSet              string  `json:"Barrel-3_Inject_5_Speed_Set" influx:"Barrel-3_Inject_5_Speed_Set"`
	Barrel3Inject6SpeedSet              string  `json:"Barrel-3_Inject_6_Speed_Set" influx:"Barrel-3_Inject_6_Speed_Set"`
	Barrel3Inject7SpeedSet              string  `json:"Barrel-3_Inject_7_Speed_Set" influx:"Barrel-3_Inject_7_Speed_Set"`
	Barrel3Inject8SpeedSet              string  `json:"Barrel-3_Inject_8_Speed_Set" influx:"Barrel-3_Inject_8_Speed_Set"`
	Barrel3Inject9SpeedSet              string  `json:"Barrel-3_Inject_9_Speed_Set" influx:"Barrel-3_Inject_9_Speed_Set"`
	Barrel3Inject10SpeedSet             string  `json:"Barrel-3_Inject_10_Speed_Set" influx:"Barrel-3_Inject_10_Speed_Set"`
	Barrel3Inject1PositionSet           string  `json:"Barrel-3_Inject_1_Position_Set" influx:"Barrel-3_Inject_1_Position_Set"`
	Barrel3Inject2PositionSet           string  `json:"Barrel-3_Inject_2_Position_Set" influx:"Barrel-3_Inject_2_Position_Set"`
	Barrel3Inject3PositionSet           string  `json:"Barrel-3_Inject_3_Position_Set" influx:"Barrel-3_Inject_3_Position_Set"`
	Barrel3Inject4PositionSet           string  `json:"Barrel-3_Inject_4_Position_Set" influx:"Barrel-3_Inject_4_Position_Set"`
	Barrel3Inject5PositionSet           string  `json:"Barrel-3_Inject_5_Position_Set" influx:"Barrel-3_Inject_5_Position_Set"`
	Barrel3Inject6PositionSet           string  `json:"Barrel-3_Inject_6_Position_Set" influx:"Barrel-3_Inject_6_Position_Set"`
	Barrel3Inject7PositionSet           string  `json:"Barrel-3_Inject_7_Position_Set" influx:"Barrel-3_Inject_7_Position_Set"`
	Barrel3Inject8PositionSet           string  `json:"Barrel-3_Inject_8_Position_Set" influx:"Barrel-3_Inject_8_Position_Set"`
	Barrel3Inject9PositionSet           string  `json:"Barrel-3_Inject_9_Position_Set" influx:"Barrel-3_Inject_9_Position_Set"`
	Barrel3Inject10PositionSet          string  `json:"Barrel-3_Inject_10_Position_Set" influx:"Barrel-3_Inject_10_Position_Set"`
	Barrel3Inject1PressureSet           string  `json:"Barrel-3_Inject_1_Pressure_Set" influx:"Barrel-3_Inject_1_Pressure_Set"`
	Barrel3Inject2PressureSet           string  `json:"Barrel-3_Inject_2_Pressure_Set" influx:"Barrel-3_Inject_2_Pressure_Set"`
	Barrel3Inject3PressureSet           string  `json:"Barrel-3_Inject_3_Pressure_Set" influx:"Barrel-3_Inject_3_Pressure_Set"`
	Barrel3Inject4PressureSet           string  `json:"Barrel-3_Inject_4_Pressure_Set" influx:"Barrel-3_Inject_4_Pressure_Set"`
	Barrel3Inject5PressureSet           string  `json:"Barrel-3_Inject_5_Pressure_Set" influx:"Barrel-3_Inject_5_Pressure_Set"`
	Barrel3Inject6PressureSet           string  `json:"Barrel-3_Inject_6_Pressure_Set" influx:"Barrel-3_Inject_6_Pressure_Set"`
	Barrel3Inject7PressureSet           string  `json:"Barrel-3_Inject_7_Pressure_Set" influx:"Barrel-3_Inject_7_Pressure_Set"`
	Barrel3Inject8PressureSet           string  `json:"Barrel-3_Inject_8_Pressure_Set" influx:"Barrel-3_Inject_8_Pressure_Set"`
	Barrel3Inject9PressureSet           string  `json:"Barrel-3_Inject_9_Pressure_Set" influx:"Barrel-3_Inject_9_Pressure_Set"`
	Barrel3Inject10PressureSet          string  `json:"Barrel-3_Inject_10_Pressure_Set" influx:"Barrel-3_Inject_10_Pressure_Set"`
	Barrel3VPInjectTransferHoldSet      string  `json:"Barrel-3_V-P_Inject_Transfer_Hold_Set" influx:"Barrel-3_V-P_Inject_Transfer_Hold_Set"`
	Barrel3InjectTimeSet                string  `json:"Barrel-3_Inject_Time_Set" influx:"Barrel-3_Inject_Time_Set"`
	Barrel3VPPositionSet                string  `json:"Barrel-3_V-P_Position_Set" influx:"Barrel-3_V-P_Position_Set"`
	Barrel3VPTransferPressureSet        string  `json:"Barrel-3_V-P_Transfer_Pressure_Set" influx:"Barrel-3_V-P_Transfer_Pressure_Set"`
	Barrel3HoldStageSet                 string  `json:"Barrel-3_Hold_Stage_Set" influx:"Barrel-3_Hold_Stage_Set"`
	Barrel3Hold1TimeSet                 string  `json:"Barrel-3_Hold_1_Time_Set" influx:"Barrel-3_Hold_1_Time_Set"`
	Barrel3Hold2TimeSet                 string  `json:"Barrel-3_Hold_2_Time_Set" influx:"Barrel-3_Hold_2_Time_Set"`
	Barrel3Hold3TimeSet                 string  `json:"Barrel-3_Hold_3_Time_Set" influx:"Barrel-3_Hold_3_Time_Set"`
	Barrel3Hold4TimeSet                 string  `json:"Barrel-3_Hold_4_Time_Set" influx:"Barrel-3_Hold_4_Time_Set"`
	Barrel3Hold5TimeSet                 string  `json:"Barrel-3_Hold_5_Time_Set" influx:"Barrel-3_Hold_5_Time_Set"`
	Barrel3Hold6TimeSet                 string  `json:"Barrel-3_Hold_6_Time_Set" influx:"Barrel-3_Hold_6_Time_Set"`
	Barrel3Hold7TimeSet                 string  `json:"Barrel-3_Hold_7_Time_Set" influx:"Barrel-3_Hold_7_Time_Set"`
	Barrel3Hold8TimeSet                 string  `json:"Barrel-3_Hold_8_Time_Set" influx:"Barrel-3_Hold_8_Time_Set"`
	Barrel3Hold9TimeSet                 string  `json:"Barrel-3_Hold_9_Time_Set" influx:"Barrel-3_Hold_9_Time_Set"`
	Barrel3Hold10TimeSet                string  `json:"Barrel-3_Hold_10_Time_Set" influx:"Barrel-3_Hold_10_Time_Set"`
	Barrel3Hold1SpeedSet                string  `json:"Barrel-3_Hold_1_Speed_Set" influx:"Barrel-3_Hold_1_Speed_Set"`
	Barrel3Hold2SpeedSet                string  `json:"Barrel-3_Hold_2_Speed_Set" influx:"Barrel-3_Hold_2_Speed_Set"`
	Barrel3Hold3SpeedSet                string  `json:"Barrel-3_Hold_3_Speed_Set" influx:"Barrel-3_Hold_3_Speed_Set"`
	Barrel3Hold4SpeedSet                string  `json:"Barrel-3_Hold_4_Speed_Set" influx:"Barrel-3_Hold_4_Speed_Set"`
	Barrel3Hold5SpeedSet                string  `json:"Barrel-3_Hold_5_Speed_Set" influx:"Barrel-3_Hold_5_Speed_Set"`
	Barrel3Hold6SpeedSet                string  `json:"Barrel-3_Hold_6_Speed_Set" influx:"Barrel-3_Hold_6_Speed_Set"`
	Barrel3Hold7SpeedSet                string  `json:"Barrel-3_Hold_7_Speed_Set" influx:"Barrel-3_Hold_7_Speed_Set"`
	Barrel3Hold8SpeedSet                string  `json:"Barrel-3_Hold_8_Speed_Set" influx:"Barrel-3_Hold_8_Speed_Set"`
	Barrel3Hold9SpeedSet                string  `json:"Barrel-3_Hold_9_Speed_Set" influx:"Barrel-3_Hold_9_Speed_Set"`
	Barrel3Hold10SpeedSet               string  `json:"Barrel-3_Hold_10_Speed_Set" influx:"Barrel-3_Hold_10_Speed_Set"`
	Barrel3Hold1PressureSet             string  `json:"Barrel-3_Hold_1_Pressure_Set" influx:"Barrel-3_Hold_1_Pressure_Set"`
	Barrel3Hold2PressureSet             string  `json:"Barrel-3_Hold_2_Pressure_Set" influx:"Barrel-3_Hold_2_Pressure_Set"`
	Barrel3Hold3PressureSet             string  `json:"Barrel-3_Hold_3_Pressure_Set" influx:"Barrel-3_Hold_3_Pressure_Set"`
	Barrel3Hold4PressureSet             string  `json:"Barrel-3_Hold_4_Pressure_Set" influx:"Barrel-3_Hold_4_Pressure_Set"`
	Barrel3Hold5PressureSet             string  `json:"Barrel-3_Hold_5_Pressure_Set" influx:"Barrel-3_Hold_5_Pressure_Set"`
	Barrel3Hold6PressureSet             string  `json:"Barrel-3_Hold_6_Pressure_Set" influx:"Barrel-3_Hold_6_Pressure_Set"`
	Barrel3Hold7PressureSet             string  `json:"Barrel-3_Hold_7_Pressure_Set" influx:"Barrel-3_Hold_7_Pressure_Set"`
	Barrel3Hold8PressureSet             string  `json:"Barrel-3_Hold_8_Pressure_Set" influx:"Barrel-3_Hold_8_Pressure_Set"`
	Barrel3Hold9PressureSet             string  `json:"Barrel-3_Hold_9_Pressure_Set" influx:"Barrel-3_Hold_9_Pressure_Set"`
	Barrel3Hold10PressureSet            string  `json:"Barrel-3_Hold_10_Pressure_Set" influx:"Barrel-3_Hold_10_Pressure_Set"`
	Barrel3TempStageSet                 string  `json:"Barrel-3_Temp_Stage_Set" influx:"Barrel-3_Temp_Stage_Set"`
	Barrel3NozzTemp1Set                 string  `json:"Barrel-3_NozzTemp_1_Set" influx:"Barrel-3_NozzTemp_1_Set"`
	Barrel3NozzTemp2Set                 string  `json:"Barrel-3_NozzTemp_2_Set" influx:"Barrel-3_NozzTemp_2_Set"`
	Barrel3Temp1stSet                   string  `json:"Barrel-3_Temp_1st_Set" influx:"Barrel-3_Temp_1st_Set"`
	Barrel3Temp2ndSet                   string  `json:"Barrel-3_Temp_2nd_Set" influx:"Barrel-3_Temp_2nd_Set"`
	Barrel3Temp3rdSet                   string  `json:"Barrel-3_Temp_3rd_Set" influx:"Barrel-3_Temp_3rd_Set"`
	Barrel3Temp4thSet                   string  `json:"Barrel-3_Temp_4th_Set" influx:"Barrel-3_Temp_4th_Set"`
	Barrel3Temp5thSet                   string  `json:"Barrel-3_Temp_5th_Set" influx:"Barrel-3_Temp_5th_Set"`
	Barrel3Temp6thSet                   string  `json:"Barrel-3_Temp_6th_Set" influx:"Barrel-3_Temp_6th_Set"`
	Barrel3Temp7thSet                   string  `json:"Barrel-3_Temp_7th_Set" influx:"Barrel-3_Temp_7th_Set"`
	Barrel3Temp8thSet                   string  `json:"Barrel-3_Temp_8th_Set" influx:"Barrel-3_Temp_8th_Set"`
	Barrel3TempfeedSet                  string  `json:"Barrel-3_Temp_feed_Set" influx:"Barrel-3_Temp_feed_Set"`
	Barrel3ChargeStageSet               string  `json:"Barrel-3_Charge_Stage_Set" influx:"Barrel-3_Charge_Stage_Set"`
	Barrel3Charge1SpeedSet              string  `json:"Barrel-3_Charge_1_Speed_Set" influx:"Barrel-3_Charge_1_Speed_Set"`
	Barrel3Charge2SpeedSet              string  `json:"Barrel-3_Charge_2_Speed_Set" influx:"Barrel-3_Charge_2_Speed_Set"`
	Barrel3Charge3SpeedSet              string  `json:"Barrel-3_Charge_3_Speed_Set" influx:"Barrel-3_Charge_3_Speed_Set"`
	Barrel3Charge4SpeedSet              string  `json:"Barrel-3_Charge_4_Speed_Set" influx:"Barrel-3_Charge_4_Speed_Set"`
	Barrel3Charge5SpeedSet              string  `json:"Barrel-3_Charge_5_Speed_Set" influx:"Barrel-3_Charge_5_Speed_Set"`
	Barrel3Charge6SpeedSet              string  `json:"Barrel-3_Charge_6_Speed_Set" influx:"Barrel-3_Charge_6_Speed_Set"`
	Barrel3Charge7SpeedSet              string  `json:"Barrel-3_Charge_7_Speed_Set" influx:"Barrel-3_Charge_7_Speed_Set"`
	Barrel3Charge8SpeedSet              string  `json:"Barrel-3_Charge_8_Speed_Set" influx:"Barrel-3_Charge_8_Speed_Set"`
	Barrel3Charge9SpeedSet              string  `json:"Barrel-3_Charge_9_Speed_Set" influx:"Barrel-3_Charge_9_Speed_Set"`
	Barrel3Charge10SpeedSet             string  `json:"Barrel-3_Charge_10_Speed_Set" influx:"Barrel-3_Charge_10_Speed_Set"`
	Barrel3Charge1PositionSet           string  `json:"Barrel-3_Charge_1_Position_Set" influx:"Barrel-3_Charge_1_Position_Set"`
	Barrel3Charge2PositionSet           string  `json:"Barrel-3_Charge_2_Position_Set" influx:"Barrel-3_Charge_2_Position_Set"`
	Barrel3Charge3PositionSet           string  `json:"Barrel-3_Charge_3_Position_Set" influx:"Barrel-3_Charge_3_Position_Set"`
	Barrel3Charge4PositionSet           string  `json:"Barrel-3_Charge_4_Position_Set" influx:"Barrel-3_Charge_4_Position_Set"`
	Barrel3Charge5PositionSet           string  `json:"Barrel-3_Charge_5_Position_Set" influx:"Barrel-3_Charge_5_Position_Set"`
	Barrel3Charge6PositionSet           string  `json:"Barrel-3_Charge_6_Position_Set" influx:"Barrel-3_Charge_6_Position_Set"`
	Barrel3Charge7PositionSet           string  `json:"Barrel-3_Charge_7_Position_Set" influx:"Barrel-3_Charge_7_Position_Set"`
	Barrel3Charge8PositionSet           string  `json:"Barrel-3_Charge_8_Position_Set" influx:"Barrel-3_Charge_8_Position_Set"`
	Barrel3Charge9PositionSet           string  `json:"Barrel-3_Charge_9_Position_Set" influx:"Barrel-3_Charge_9_Position_Set"`
	Barrel3Charge10PositionSet          string  `json:"Barrel-3_Charge_10_Position_Set" influx:"Barrel-3_Charge_10_Position_Set"`
	Barrel3Charge1BackPressureSet       string  `json:"Barrel-3_Charge_1_BackPressure_Set" influx:"Barrel-3_Charge_1_BackPressure_Set"`
	Barrel3Charge2BackPressureSet       string  `json:"Barrel-3_Charge_2_BackPressure_Set" influx:"Barrel-3_Charge_2_BackPressure_Set"`
	Barrel3Charge3BackPressureSet       string  `json:"Barrel-3_Charge_3_BackPressure_Set" influx:"Barrel-3_Charge_3_BackPressure_Set"`
	Barrel3Charge4BackPressureSet       string  `json:"Barrel-3_Charge_4_BackPressure_Set" influx:"Barrel-3_Charge_4_BackPressure_Set"`
	Barrel3Charge5BackPressureSet       string  `json:"Barrel-3_Charge_5_BackPressure_Set" influx:"Barrel-3_Charge_5_BackPressure_Set"`
	Barrel3Charge6BackPressureSet       string  `json:"Barrel-3_Charge_6_BackPressure_Set" influx:"Barrel-3_Charge_6_BackPressure_Set"`
	Barrel3Charge7BackPressureSet       string  `json:"Barrel-3_Charge_7_BackPressure_Set" influx:"Barrel-3_Charge_7_BackPressure_Set"`
	Barrel3Charge8BackPressureSet       string  `json:"Barrel-3_Charge_8_BackPressure_Set" influx:"Barrel-3_Charge_8_BackPressure_Set"`
	Barrel3Charge9BackPressureSet       string  `json:"Barrel-3_Charge_9_BackPressure_Set" influx:"Barrel-3_Charge_9_BackPressure_Set"`
	Barrel3Charge10BackPressureSet      string  `json:"Barrel-3_Charge_10_BackPressure_Set" influx:"Barrel-3_Charge_10_BackPressure_Set"`
	Barrel3ChargeProtectTimeSet         string  `json:"Barrel-3_Charge_Protect_Time_Set" influx:"Barrel-3_Charge_Protect_Time_Set"`
	Barrel3MaxProtectTimeSet            string  `json:"Barrel-3_Max_Protect_Time_Set" influx:"Barrel-3_Max_Protect_Time_Set"`
	Barrel3SuckModeSet                  string  `json:"Barrel-3_Suck_Mode_Set" influx:"Barrel-3_Suck_Mode_Set"`
	Barrel3PreSuckSpeedSet              string  `json:"Barrel-3_Pre_Suck_Speed_Set" influx:"Barrel-3_Pre_Suck_Speed_Set"`
	Barrel3PreSuckPositionSet           string  `json:"Barrel-3_Pre_Suck_Position_Set" influx:"Barrel-3_Pre_Suck_Position_Set"`
	Barrel3PreSuckPressureSet           string  `json:"Barrel-3_Pre_Suck_Pressure_Set" influx:"Barrel-3_Pre_Suck_Pressure_Set"`
	Barrel3BackSuckModeSet              string  `json:"Barrel-3_Back_Suck_Mode_Set" influx:"Barrel-3_Back_Suck_Mode_Set"`
	Barrel3BackSuckSpeedSet             string  `json:"Barrel-3_Back_Suck_Speed_Set" influx:"Barrel-3_Back_Suck_Speed_Set"`
	Barrel3BackSuckPositionSet          string  `json:"Barrel-3_Back_Suck_Position_Set" influx:"Barrel-3_Back_Suck_Position_Set"`
	Barrel3BackSuckPressureSet          string  `json:"Barrel-3_Back_Suck_Pressure_Set" influx:"Barrel-3_Back_Suck_Pressure_Set"`
	CoolTimeSet                         string  `json:"Cool_Time_Set" influx:"Cool_Time_Set"`
	OilTankTempLimitSet                 string  `json:"Oil_TankTemp_Limit_Set" influx:"Oil_TankTemp_Limit_Set"`
	MoldTempStageSet                    string  `json:"Mold_Temp_Stage_Set" influx:"Mold_Temp_Stage_Set"`
	MoldTemp1Set                        string  `json:"Mold_Temp_1_Set" influx:"Mold_Temp_1_Set"`
	MoldTemp2Set                        string  `json:"Mold_Temp_2_Set" influx:"Mold_Temp_2_Set"`
	MoldTemp3Set                        string  `json:"Mold_Temp_3_Set" influx:"Mold_Temp_3_Set"`
	MoldTemp4Set                        string  `json:"Mold_Temp_4_Set" influx:"Mold_Temp_4_Set"`
	MoldTemp5Set                        string  `json:"Mold_Temp_5_Set" influx:"Mold_Temp_5_Set"`
	MoldTemp6Set                        string  `json:"Mold_Temp_6_Set" influx:"Mold_Temp_6_Set"`
	MoldTemp7Set                        string  `json:"Mold_Temp_7_Set" influx:"Mold_Temp_7_Set"`
	MoldTemp8Set                        string  `json:"Mold_Temp_8_Set" influx:"Mold_Temp_8_Set"`
	MoldTemp9Set                        string  `json:"Mold_Temp_9_Set" influx:"Mold_Temp_9_Set"`
	MoldTemp10Set                       string  `json:"Mold_Temp_10_Set" influx:"Mold_Temp_10_Set"`
	MoldTemp11Set                       string  `json:"Mold_Temp_11_Set" influx:"Mold_Temp_11_Set"`
	MoldTemp12Set                       string  `json:"Mold_Temp_12_Set" influx:"Mold_Temp_12_Set"`
	ClampForceSet                       string  `json:"Clamp_Force_Set" influx:"Clamp_Force_Set"`
	MoldOpenStageSet                    string  `json:"MoldOpen_Stage_Set" influx:"MoldOpen_Stage_Set"`
	MoldOpen1SpeedSet                   string  `json:"MoldOpen_1_Speed_Set" influx:"MoldOpen_1_Speed_Set"`
	MoldOpen2SpeedSet                   string  `json:"MoldOpen_2_Speed_Set" influx:"MoldOpen_2_Speed_Set"`
	MoldOpen3SpeedSet                   string  `json:"MoldOpen_3_Speed_Set" influx:"MoldOpen_3_Speed_Set"`
	MoldOpen4SpeedSet                   string  `json:"MoldOpen_4_Speed_Set" influx:"MoldOpen_4_Speed_Set"`
	MoldOpen5SpeedSet                   string  `json:"MoldOpen_5_Speed_Set" influx:"MoldOpen_5_Speed_Set"`
	MoldOpen1PositionSet                string  `json:"MoldOpen_1_Position_Set" influx:"MoldOpen_1_Position_Set"`
	MoldOpen2PositionSet                string  `json:"MoldOpen_2_Position_Set" influx:"MoldOpen_2_Position_Set"`
	MoldOpen3PositionSet                string  `json:"MoldOpen_3_Position_Set" influx:"MoldOpen_3_Position_Set"`
	MoldOpen4PositionSet                string  `json:"MoldOpen_4_Position_Set" influx:"MoldOpen_4_Position_Set"`
	MoldOpen5PositionSet                string  `json:"MoldOpen_5_Position_Set" influx:"MoldOpen_5_Position_Set"`
	MoldOpen1PressureSet                string  `json:"MoldOpen_1_Pressure_Set" influx:"MoldOpen_1_Pressure_Set"`
	MoldOpen2PressureSet                string  `json:"MoldOpen_2_Pressure_Set" influx:"MoldOpen_2_Pressure_Set"`
	MoldOpen3PressureSet                string  `json:"MoldOpen_3_Pressure_Set" influx:"MoldOpen_3_Pressure_Set"`
	MoldOpen4PressureSet                string  `json:"MoldOpen_4_Pressure_Set" influx:"MoldOpen_4_Pressure_Set"`
	MoldOpen5PressureSet                string  `json:"MoldOpen_5_Pressure_Set" influx:"MoldOpen_5_Pressure_Set"`
	MoldCloseStageSet                   string  `json:"MoldClose_Stage_Set" influx:"MoldClose_Stage_Set"`
	MoldClose1SpeedSet                  string  `json:"MoldClose_1_Speed_Set" influx:"MoldClose_1_Speed_Set"`
	MoldClose2SpeedSet                  string  `json:"MoldClose_2_Speed_Set" influx:"MoldClose_2_Speed_Set"`
	MoldClose3SpeedSet                  string  `json:"MoldClose_3_Speed_Set" influx:"MoldClose_3_Speed_Set"`
	MoldClose4SpeedSet                  string  `json:"MoldClose_4_Speed_Set" influx:"MoldClose_4_Speed_Set"`
	MoldClose5SpeedSet                  string  `json:"MoldClose_5_Speed_Set" influx:"MoldClose_5_Speed_Set"`
	MoldClose1PositionSet               string  `json:"MoldClose_1_Position_Set" influx:"MoldClose_1_Position_Set"`
	MoldClose2PositionSet               string  `json:"MoldClose_2_Position_Set" influx:"MoldClose_2_Position_Set"`
	MoldClose3PositionSet               string  `json:"MoldClose_3_Position_Set" influx:"MoldClose_3_Position_Set"`
	MoldClose4PositionSet               string  `json:"MoldClose_4_Position_Set" influx:"MoldClose_4_Position_Set"`
	MoldClose5PositionSet               string  `json:"MoldClose_5_Position_Set" influx:"MoldClose_5_Position_Set"`
	MoldClose1PressureSet               string  `json:"MoldClose_1_Pressure_Set" influx:"MoldClose_1_Pressure_Set"`
	MoldClose2PressureSet               string  `json:"MoldClose_2_Pressure_Set" influx:"MoldClose_2_Pressure_Set"`
	MoldClose3PressureSet               string  `json:"MoldClose_3_Pressure_Set" influx:"MoldClose_3_Pressure_Set"`
	MoldClose4PressureSet               string  `json:"MoldClose_4_Pressure_Set" influx:"MoldClose_4_Pressure_Set"`
	MoldClose5PressureSet               string  `json:"MoldClose_5_Pressure_Set" influx:"MoldClose_5_Pressure_Set"`
	Ejector1AdvanceStageSet             string  `json:"1-EjectorAdvance_Stage_Set" influx:"1-EjectorAdvance_Stage_Set"`
	Ejector1Advance1SpeedSet            string  `json:"1-EjectorAdvance_1_Speed_Set" influx:"1-EjectorAdvance_1_Speed_Set"`
	Ejector1Advance2SpeedSet            string  `json:"1-EjectorAdvance_2_Speed_Set" influx:"1-EjectorAdvance_2_Speed_Set"`
	Ejector1Advance3SpeedSet            string  `json:"1-EjectorAdvance_3_Speed_Set" influx:"1-EjectorAdvance_3_Speed_Set"`
	Ejector1Advance1PositionSet         string  `json:"1-EjectorAdvance_1_Position_Set" influx:"1-EjectorAdvance_1_Position_Set"`
	Ejector1Advance2PositionSet         string  `json:"1-EjectorAdvance_2_Position_Set" influx:"1-EjectorAdvance_2_Position_Set"`
	Ejector1Advance3PositionSet         string  `json:"1-EjectorAdvance_3_Position_Set" influx:"1-EjectorAdvance_3_Position_Set"`
	Ejector1Advance1PressureSet         string  `json:"1-EjectorAdvance_1_Pressure_Set" influx:"1-EjectorAdvance_1_Pressure_Set"`
	Ejector1Advance2PressureSet         string  `json:"1-EjectorAdvance_2_Pressure_Set" influx:"1-EjectorAdvance_2_Pressure_Set"`
	Ejector1Advance3PressureSet         string  `json:"1-EjectorAdvance_3_Pressure_Set" influx:"1-EjectorAdvance_3_Pressure_Set"`
	Ejector2AdvanceStageSet             string  `json:"2-EjectorAdvance_Stage_Set" influx:"2-EjectorAdvance_Stage_Set"`
	Ejector2Advance1SpeedSet            string  `json:"2-EjectorAdvance_1_Speed_Set" influx:"2-EjectorAdvance_1_Speed_Set"`
	Ejector2Advance2SpeedSet            string  `json:"2-EjectorAdvance_2_Speed_Set" influx:"2-EjectorAdvance_2_Speed_Set"`
	Ejector2Advance3SpeedSet            string  `json:"2-EjectorAdvance_3_Speed_Set" influx:"2-EjectorAdvance_3_Speed_Set"`
	Ejector2Advance1PositionSet         string  `json:"2-EjectorAdvance_1_Position_Set" influx:"2-EjectorAdvance_1_Position_Set"`
	Ejector2Advance2PositionSet         string  `json:"2-EjectorAdvance_2_Position_Set" influx:"2-EjectorAdvance_2_Position_Set"`
	Ejector2Advance3PositionSet         string  `json:"2-EjectorAdvance_3_Position_Set" influx:"2-EjectorAdvance_3_Position_Set"`
	Ejector2Advance1PressureSet         string  `json:"2-EjectorAdvance_1_Pressure_Set" influx:"2-EjectorAdvance_1_Pressure_Set"`
	Ejector2Advance2PressureSet         string  `json:"2-EjectorAdvance_2_Pressure_Set" influx:"2-EjectorAdvance_2_Pressure_Set"`
	Ejector2Advance3PressureSet         string  `json:"2-EjectorAdvance_3_Pressure_Set" influx:"2-EjectorAdvance_3_Pressure_Set"`
	Ejector1RetractStageSet             string  `json:"1-EjectorRetract_Stage_Set" influx:"1-EjectorRetract_Stage_Set"`
	Ejector1Retract1SpeedSet            string  `json:"1-EjectorRetract_1_Speed_Set" influx:"1-EjectorRetract_1_Speed_Set"`
	Ejector1Retract2SpeedSet            string  `json:"1-EjectorRetract_2_Speed_Set" influx:"1-EjectorRetract_2_Speed_Set"`
	Ejector1Retract3SpeedSet            string  `json:"1-EjectorRetract_3_Speed_Set" influx:"1-EjectorRetract_3_Speed_Set"`
	Ejector1Retract1PositionSet         string  `json:"1-EjectorRetract_1_Position_Set" influx:"1-EjectorRetract_1_Position_Set"`
	Ejector1Retract2PositionSet         string  `json:"1-EjectorRetract_2_Position_Set" influx:"1-EjectorRetract_2_Position_Set"`
	Ejector1Retract3PositionSet         string  `json:"1-EjectorRetract_3_Position_Set" influx:"1-EjectorRetract_3_Position_Set"`
	Ejector1Retract1PressureSet         string  `json:"1-EjectorRetract_1_Pressure_Set" influx:"1-EjectorRetract_1_Pressure_Set"`
	Ejector1Retract2PressureSet         string  `json:"1-EjectorRetract_2_Pressure_Set" influx:"1-EjectorRetract_2_Pressure_Set"`
	Ejector1Retract3PressureSet         string  `json:"1-EjectorRetract_3_Pressure_Set" influx:"1-EjectorRetract_3_Pressure_Set"`
	Ejector2RetractStageSet             string  `json:"2-EjectorRetract_Stage_Set" influx:"2-EjectorRetract_Stage_Set"`
	Ejector2Retract1SpeedSet            string  `json:"2-EjectorRetract_1_Speed_Set" influx:"2-EjectorRetract_1_Speed_Set"`
	Ejector2Retract2SpeedSet            string  `json:"2-EjectorRetract_2_Speed_Set" influx:"2-EjectorRetract_2_Speed_Set"`
	Ejector2Retract3SpeedSet            string  `json:"2-EjectorRetract_3_Speed_Set" influx:"2-EjectorRetract_3_Speed_Set"`
	Ejector2Retract1PositionSet         string  `json:"2-EjectorRetract_1_Position_Set" influx:"2-EjectorRetract_1_Position_Set"`
	Ejector2Retract2PositionSet         string  `json:"2-EjectorRetract_2_Position_Set" influx:"2-EjectorRetract_2_Position_Set"`
	Ejector2Retract3PositionSet         string  `json:"2-EjectorRetract_3_Position_Set" influx:"2-EjectorRetract_3_Position_Set"`
	Ejector2Retract1PressureSet         string  `json:"2-EjectorRetract_1_Pressure_Set" influx:"2-EjectorRetract_1_Pressure_Set"`
	Ejector2Retract2PressureSet         string  `json:"2-EjectorRetract_2_Pressure_Set" influx:"2-EjectorRetract_2_Pressure_Set"`
	Ejector2Retract3PressureSet         string  `json:"2-EjectorRetract_3_Pressure_Set" influx:"2-EjectorRetract_3_Pressure_Set"`
	MoldProtectTimeSet                  string  `json:"Mold_Protect_Time_Set" influx:"Mold_Protect_Time_Set"`
	Eject1ModeSet                       string  `json:"1-Eject_Mode_Set" influx:"1-Eject_Mode_Set"`
	Eject2ModeSet                       string  `json:"2-Eject_Mode_Set" influx:"2-Eject_Mode_Set"`
	Eject1CounterSet                    string  `json:"1-Eject_Counter_Set" influx:"1-Eject_Counter_Set"`
	Eject2CounterSet                    string  `json:"2-Eject_Counter_Set" influx:"2-Eject_Counter_Set"`
	OilTankTempActual                   string  `json:"Oil_TankTemp_Actual" influx:"Oil_TankTemp_Actual"`
	PeakClampForceActual                string  `json:"PeakClampForce_Actual" influx:"PeakClampForce_Actual"`
	Ejection1TimeSumActual              string  `json:"1-Ejection_Time_Sum_Actual" influx:"1-Ejection_Time_Sum_Actual"`
	Ejection2TimeSumActual              string  `json:"2-Ejection_Time_Sum_Actual" influx:"2-Ejection_Time_Sum_Actual"`
	Barrel1InjectFillTimeActual         string  `json:"Barrel-1_InjectFillTime_Actual" influx:"Barrel-1_InjectFillTime_Actual"`
	Barrel1PeakInjectSpeedActual        string  `json:"Barrel-1_PeakInjectSpeed_Actual" influx:"Barrel-1_PeakInjectSpeed_Actual"`
	Barrel1CushionPositionActual        string  `json:"Barrel-1_CushionPosition_Actual" influx:"Barrel-1_CushionPosition_Actual"`
	Barrel1PeakInjectPressureActual     string  `json:"Barrel-1_PeakInjectPressure_Actual" influx:"Barrel-1_PeakInjectPressure_Actual"`
	Barrel1VPPositionActual             string  `json:"Barrel-1_V-P_Position_Actual" influx:"Barrel-1_V-P_Position_Actual"`
	Barrel1VPPressureActual             string  `json:"Barrel-1_V-P_Pressure_Actual" influx:"Barrel-1_V-P_Pressure_Actual"`
	Barrel1HoldPressurePostionActual    string  `json:"Barrel-1_HoldPressure_Postion_Actual" influx:"Barrel-1_HoldPressure_Postion_Actual"`
	Barrel1PeakHoldPressureActual       string  `json:"Barrel-1_PeakHoldPressure_Actual" influx:"Barrel-1_PeakHoldPressure_Actual"`
	Barrel1NozzTemp1Actual              string  `json:"Barrel-1_NozzTemp_1_Actual" influx:"Barrel-1_NozzTemp_1_Actual"`
	Barrel1NozzTemp2Actual              string  `json:"Barrel-1_NozzTemp_2_Actual" influx:"Barrel-1_NozzTemp_2_Actual"`
	Barrel1Temp1Actual                  string  `json:"Barrel-1_Temp_1_Actual" influx:"Barrel-1_Temp_1_Actual"`
	Barrel1Temp2Actual                  string  `json:"Barrel-1_Temp_2_Actual" influx:"Barrel-1_Temp_2_Actual"`
	Barrel1Temp3Actual                  string  `json:"Barrel-1_Temp_3_Actual" influx:"Barrel-1_Temp_3_Actual"`
	Barrel1Temp4Actual                  string  `json:"Barrel-1_Temp_4_Actual" influx:"Barrel-1_Temp_4_Actual"`
	Barrel1Temp5Actual                  string  `json:"Barrel-1_Temp_5_Actual" influx:"Barrel-1_Temp_5_Actual"`
	Barrel1Temp6Actual                  string  `json:"Barrel-1_Temp_6_Actual" influx:"Barrel-1_Temp_6_Actual"`
	Barrel1Temp7Actual                  string  `json:"Barrel-1_Temp_7_Actual" influx:"Barrel-1_Temp_7_Actual"`
	Barrel1Temp8Actual                  string  `json:"Barrel-1_Temp_8_Actual" influx:"Barrel-1_Temp_8_Actual"`
	Barrel1TempfeedActual               string  `json:"Barrel-1_Temp_feed_Actual" influx:"Barrel-1_Temp_feed_Actual"`
	Barrel1ChargeTimeActual             string  `json:"Barrel-1_ChargeTime_Actual" influx:"Barrel-1_ChargeTime_Actual"`
	Barrel1ChargeStartPositionActual    string  `json:"Barrel-1_ChargeStartPosition_Actual" influx:"Barrel-1_ChargeStartPosition_Actual"`
	Barrel1WholeInjectionTimeActual     string  `json:"Barrel-1_Whole_Injection_Time_Actual" influx:"Barrel-1_Whole_Injection_Time_Actual"`
	Barrel1ScrewRPMActual               string  `json:"Barrel-1_Screw_RPM_Actual" influx:"Barrel-1_Screw_RPM_Actual"`
	Barrel1MoldopenendpositionActual    string  `json:"Barrel-1_Mold_open_end_position_Actual" influx:"Barrel-1_Mold_open_end_position_Actual"`
	Barrel1InjectionstartpositionActual string  `json:"Barrel-1_Injection_start_position_Actual" influx:"Barrel-1_Injection_start_position_Actual"`
	Barrel1ChargeendpositionActual      string  `json:"Barrel-1_Charge_end_position_Actual" influx:"Barrel-1_Charge_end_position_Actual"`
	Barrel1ShortresinpositionActual     string  `json:"Barrel-1_Short_resin_position_Actual" influx:"Barrel-1_Short_resin_position_Actual"`
	Barrel2InjectFillTimeActual         string  `json:"Barrel-2_InjectFillTime_Actual" influx:"Barrel-2_InjectFillTime_Actual"`
	Barrel2PeakInjectSpeedActual        string  `json:"Barrel-2_PeakInjectSpeed_Actual" influx:"Barrel-2_PeakInjectSpeed_Actual"`
	Barrel2CushionPositionActual        string  `json:"Barrel-2_CushionPosition_Actual" influx:"Barrel-2_CushionPosition_Actual"`
	Barrel2PeakInjectPressureActual     string  `json:"Barrel-2_PeakInjectPressure_Actual" influx:"Barrel-2_PeakInjectPressure_Actual"`
	Barrel2VPPositionActual             string  `json:"Barrel-2_V-P_Position_Actual" influx:"Barrel-2_V-P_Position_Actual"`
	Barrel2VPPressureActual             string  `json:"Barrel-2_V-P_Pressure_Actual" influx:"Barrel-2_V-P_Pressure_Actual"`
	Barrel2HoldPressurePostionActual    string  `json:"Barrel-2_HoldPressure_Postion_Actual" influx:"Barrel-2_HoldPressure_Postion_Actual"`
	Barrel2PeakHoldPressureActual       string  `json:"Barrel-2_PeakHoldPressure_Actual" influx:"Barrel-2_PeakHoldPressure_Actual"`
	Barrel2NozzTemp1Actual              string  `json:"Barrel-2_NozzTemp_1_Actual" influx:"Barrel-2_NozzTemp_1_Actual"`
	Barrel2NozzTemp2Actual              string  `json:"Barrel-2_NozzTemp_2_Actual" influx:"Barrel-2_NozzTemp_2_Actual"`
	Barrel2Temp1Actual                  string  `json:"Barrel-2_Temp_1_Actual" influx:"Barrel-2_Temp_1_Actual"`
	Barrel2Temp2Actual                  string  `json:"Barrel-2_Temp_2_Actual" influx:"Barrel-2_Temp_2_Actual"`
	Barrel2Temp3Actual                  string  `json:"Barrel-2_Temp_3_Actual" influx:"Barrel-2_Temp_3_Actual"`
	Barrel2Temp4Actual                  string  `json:"Barrel-2_Temp_4_Actual" influx:"Barrel-2_Temp_4_Actual"`
	Barrel2Temp5Actual                  string  `json:"Barrel-2_Temp_5_Actual" influx:"Barrel-2_Temp_5_Actual"`
	Barrel2Temp6Actual                  string  `json:"Barrel-2_Temp_6_Actual" influx:"Barrel-2_Temp_6_Actual"`
	Barrel2Temp7Actual                  string  `json:"Barrel-2_Temp_7_Actual" influx:"Barrel-2_Temp_7_Actual"`
	Barrel2Temp8Actual                  string  `json:"Barrel-2_Temp_8_Actual" influx:"Barrel-2_Temp_8_Actual"`
	Barrel2TempfeedActual               string  `json:"Barrel-2_Temp_feed_Actual" influx:"Barrel-2_Temp_feed_Actual"`
	Barrel2ChargeTimeActual             string  `json:"Barrel-2_ChargeTime_Actual" influx:"Barrel-2_ChargeTime_Actual"`
	Barrel2ChargeStartPositionActual    string  `json:"Barrel-2_ChargeStartPosition_Actual" influx:"Barrel-2_ChargeStartPosition_Actual"`
	Barrel2WholeInjectionTimeActual     string  `json:"Barrel-2_Whole_Injection_Time_Actual" influx:"Barrel-2_Whole_Injection_Time_Actual"`
	Barrel2ScrewRPMActual               string  `json:"Barrel-2_Screw_RPM_Actual" influx:"Barrel-2_Screw_RPM_Actual"`
	Barrel2MoldopenendpositionActual    string  `json:"Barrel-2_Mold_open_end_position_Actual" influx:"Barrel-2_Mold_open_end_position_Actual"`
	Barrel2InjectionstartpositionActual string  `json:"Barrel-2_Injection_start_position_Actual" influx:"Barrel-2_Injection_start_position_Actual"`
	Barrel2ChargeendpositionActual      string  `json:"Barrel-2_Charge_end_position_Actual" influx:"Barrel-2_Charge_end_position_Actual"`
	Barrel2ShortresinpositionActual     string  `json:"Barrel-2_Short_resin_position_Actual" influx:"Barrel-2_Short_resin_position_Actual"`
	Barrel3InjectFillTimeActual         string  `json:"Barrel-3_InjectFillTime_Actual" influx:"Barrel-3_InjectFillTime_Actual"`
	Barrel3PeakInjectSpeedActual        string  `json:"Barrel-3_PeakInjectSpeed_Actual" influx:"Barrel-3_PeakInjectSpeed_Actual"`
	Barrel3CushionPositionActual        string  `json:"Barrel-3_CushionPosition_Actual" influx:"Barrel-3_CushionPosition_Actual"`
	Barrel3PeakInjectPressureActual     string  `json:"Barrel-3_PeakInjectPressure_Actual" influx:"Barrel-3_PeakInjectPressure_Actual"`
	Barrel3VPPositionActual             string  `json:"Barrel-3_V-P_Position_Actual" influx:"Barrel-3_V-P_Position_Actual"`
	Barrel3VPPressureActual             string  `json:"Barrel-3_V-P_Pressure_Actual" influx:"Barrel-3_V-P_Pressure_Actual"`
	Barrel3HoldPressurePostionActual    string  `json:"Barrel-3_HoldPressure_Postion_Actual" influx:"Barrel-3_HoldPressure_Postion_Actual"`
	Barrel3PeakHoldPressureActual       string  `json:"Barrel-3_PeakHoldPressure_Actual" influx:"Barrel-3_PeakHoldPressure_Actual"`
	Barrel3NozzTemp1Actual              string  `json:"Barrel-3_NozzTemp_1_Actual" influx:"Barrel-3_NozzTemp_1_Actual"`
	Barrel3NozzTemp2Actual              string  `json:"Barrel-3_NozzTemp_2_Actual" influx:"Barrel-3_NozzTemp_2_Actual"`
	Barrel3Temp1Actual                  string  `json:"Barrel-3_Temp_1_Actual" influx:"Barrel-3_Temp_1_Actual"`
	Barrel3Temp2Actual                  string  `json:"Barrel-3_Temp_2_Actual" influx:"Barrel-3_Temp_2_Actual"`
	Barrel3Temp3Actual                  string  `json:"Barrel-3_Temp_3_Actual" influx:"Barrel-3_Temp_3_Actual"`
	Barrel3Temp4Actual                  string  `json:"Barrel-3_Temp_4_Actual" influx:"Barrel-3_Temp_4_Actual"`
	Barrel3Temp5Actual                  string  `json:"Barrel-3_Temp_5_Actual" influx:"Barrel-3_Temp_5_Actual"`
	Barrel3Temp6Actual                  string  `json:"Barrel-3_Temp_6_Actual" influx:"Barrel-3_Temp_6_Actual"`
	Barrel3Temp7Actual                  string  `json:"Barrel-3_Temp_7_Actual" influx:"Barrel-3_Temp_7_Actual"`
	Barrel3Temp8Actual                  string  `json:"Barrel-3_Temp_8_Actual" influx:"Barrel-3_Temp_8_Actual"`
	Barrel3TempfeedActual               string  `json:"Barrel-3_Temp_feed_Actual" influx:"Barrel-3_Temp_feed_Actual"`
	Barrel3ChargeTimeActual             string  `json:"Barrel-3_ChargeTime_Actual" influx:"Barrel-3_ChargeTime_Actual"`
	Barrel3ChargeStartPositionActual    string  `json:"Barrel-3_ChargeStartPosition_Actual" influx:"Barrel-3_ChargeStartPosition_Actual"`
	Barrel3WholeInjectionTimeActual     string  `json:"Barrel-3_Whole_Injection_Time_Actual" influx:"Barrel-3_Whole_Injection_Time_Actual"`
	Barrel3ScrewRPMActual               string  `json:"Barrel-3_Screw_RPM_Actual" influx:"Barrel-3_Screw_RPM_Actual"`
	Barrel3MoldopenendpositionActual    string  `json:"Barrel-3_Mold_open_end_position_Actual" influx:"Barrel-3_Mold_open_end_position_Actual"`
	Barrel3InjectionstartpositionActual string  `json:"Barrel-3_Injection_start_position_Actual" influx:"Barrel-3_Injection_start_position_Actual"`
	Barrel3ChargeendpositionActual      string  `json:"Barrel-3_Charge_end_position_Actual" influx:"Barrel-3_Charge_end_position_Actual"`
	Barrel3ShortresinpositionActual     string  `json:"Barrel-3_Short_resin_position_Actual" influx:"Barrel-3_Short_resin_position_Actual"`
	CycleTimeCCL                        string  `json:"Cycle_Time_CCL" influx:"Cycle_Time_CCL"`
	CycleTimeTOL                        string  `json:"Cycle_Time_TOL" influx:"Cycle_Time_TOL"`
	MoldOpenTimeCCL                     string  `json:"MoldOpen_Time_CCL" influx:"MoldOpen_Time_CCL"`
	MoldOpenTimeTOL                     string  `json:"MoldOpen_Time_TOL" influx:"MoldOpen_Time_TOL"`
	MoldopenendpositionCCL              string  `json:"Mold_open_end_position_CCL" influx:"Mold_open_end_position_CCL"`
	MoldopenendpositionTOL              string  `json:"Mold_open_end_position_TOL" influx:"Mold_open_end_position_TOL"`
	MoldCloseTimeCCL                    string  `json:"MoldClose_Time_CCL" influx:"MoldClose_Time_CCL"`
	MoldCloseTimeTOL                    string  `json:"MoldClose_Time_TOL" influx:"MoldClose_Time_TOL"`
	Ejection1TimeSumCCL                 string  `json:"1-Ejection_Time_Sum_CCL" influx:"1-Ejection_Time_Sum_CCL"`
	Ejection1TimeSumTOL                 string  `json:"1-Ejection_Time_Sum_TOL" influx:"1-Ejection_Time_Sum_TOL"`
	Ejection2TimeSumCCL                 string  `json:"2-Ejection_Time_Sum_CCL" influx:"2-Ejection_Time_Sum_CCL"`
	Ejection2TimeSumTOL                 string  `json:"2-Ejection_Time_Sum_TOL" influx:"2-Ejection_Time_Sum_TOL"`
	PeakClampForceCCL                   string  `json:"PeakClampForce_CCL" influx:"PeakClampForce_CCL"`
	PeakClampForceTOL                   string  `json:"PeakClampForce_TOL" influx:"PeakClampForce_TOL"`
	Barrel1InjectFillTimeCCL            string  `json:"Barrel-1_InjectFillTime_CCL" influx:"Barrel-1_InjectFillTime_CCL"`
	Barrel1InjectFillTimeTOL            string  `json:"Barrel-1_InjectFillTime_TOL" influx:"Barrel-1_InjectFillTime_TOL"`
	Barrel1CushionPositionCCL           string  `json:"Barrel-1_CushionPosition_CCL" influx:"Barrel-1_CushionPosition_CCL"`
	Barrel1CushionPositionTOL           string  `json:"Barrel-1_CushionPosition_TOL" influx:"Barrel-1_CushionPosition_TOL"`
	Barrel1PeakInjectPressureCCL        string  `json:"Barrel-1_PeakInjectPressure_CCL" influx:"Barrel-1_PeakInjectPressure_CCL"`
	Barrel1PeakInjectPressureTOL        string  `json:"Barrel-1_PeakInjectPressure_TOL" influx:"Barrel-1_PeakInjectPressure_TOL"`
	Barrel1VPPositionCCL                string  `json:"Barrel-1_V-P_Position_CCL" influx:"Barrel-1_V-P_Position_CCL"`
	Barrel1VPPositionTOL                string  `json:"Barrel-1_V-P_Position_TOL" influx:"Barrel-1_V-P_Position_TOL"`
	Barrel1VPPressureCCL                string  `json:"Barrel-1_V-P_Pressure_CCL" influx:"Barrel-1_V-P_Pressure_CCL"`
	Barrel1VPPressureTOL                string  `json:"Barrel-1_V-P_Pressure_TOL" influx:"Barrel-1_V-P_Pressure_TOL"`
	Barrel1HoldPressurePostionCCL       string  `json:"Barrel-1_HoldPressure_Postion_CCL" influx:"Barrel-1_HoldPressure_Postion_CCL"`
	Barrel1HoldPressurePostionTOL       string  `json:"Barrel-1_HoldPressure_Postion_TOL" influx:"Barrel-1_HoldPressure_Postion_TOL"`
	Barrel1PeakHoldPressureCCL          string  `json:"Barrel-1_PeakHoldPressure_CCL" influx:"Barrel-1_PeakHoldPressure_CCL"`
	Barrel1PeakHoldPressureTOL          string  `json:"Barrel-1_PeakHoldPressure_TOL" influx:"Barrel-1_PeakHoldPressure_TOL"`
	Barrel1ChargeTimeCCL                string  `json:"Barrel-1_ChargeTime_CCL" influx:"Barrel-1_ChargeTime_CCL"`
	Barrel1ChargeTimeTOL                string  `json:"Barrel-1_ChargeTime_TOL" influx:"Barrel-1_ChargeTime_TOL"`
	Barrel1ChargeStartPositionCCL       string  `json:"Barrel-1_ChargeStartPosition_CCL" influx:"Barrel-1_ChargeStartPosition_CCL"`
	Barrel1ChargeStartPositionTOL       string  `json:"Barrel-1_ChargeStartPosition_TOL" influx:"Barrel-1_ChargeStartPosition_TOL"`
	Barrel1WholeInjectionTimeCCL        string  `json:"Barrel-1_Whole_Injection_Time_CCL" influx:"Barrel-1_Whole_Injection_Time_CCL"`
	Barrel1WholeInjectionTimeTOL        string  `json:"Barrel-1_Whole_Injection_Time_TOL" influx:"Barrel-1_Whole_Injection_Time_TOL"`
	Barrel1ScrewRPMCCL                  string  `json:"Barrel-1_Screw_RPM_CCL" influx:"Barrel-1_Screw_RPM_CCL"`
	Barrel1ScrewRPMTOL                  string  `json:"Barrel-1_Screw_RPM_TOL" influx:"Barrel-1_Screw_RPM_TOL"`
	Barrel2InjectFillTimeCCL            string  `json:"Barrel-2_InjectFillTime_CCL" influx:"Barrel-2_InjectFillTime_CCL"`
	Barrel2InjectFillTimeTOL            string  `json:"Barrel-2_InjectFillTime_TOL" influx:"Barrel-2_InjectFillTime_TOL"`
	Barrel2CushionPositionCCL           string  `json:"Barrel-2_CushionPosition_CCL" influx:"Barrel-2_CushionPosition_CCL"`
	Barrel2CushionPositionTOL           string  `json:"Barrel-2_CushionPosition_TOL" influx:"Barrel-2_CushionPosition_TOL"`
	Barrel2PeakInjectPressureCCL        string  `json:"Barrel-2_PeakInjectPressure_CCL" influx:"Barrel-2_PeakInjectPressure_CCL"`
	Barrel2PeakInjectPressureTOL        string  `json:"Barrel-2_PeakInjectPressure_TOL" influx:"Barrel-2_PeakInjectPressure_TOL"`
	Barrel2VPPositionCCL                string  `json:"Barrel-2_V-P_Position_CCL" influx:"Barrel-2_V-P_Position_CCL"`
	Barrel2VPPositionTOL                string  `json:"Barrel-2_V-P_Position_TOL" influx:"Barrel-2_V-P_Position_TOL"`
	Barrel2VPPressureCCL                string  `json:"Barrel-2_V-P_Pressure_CCL" influx:"Barrel-2_V-P_Pressure_CCL"`
	Barrel2VPPressureTOL                string  `json:"Barrel-2_V-P_Pressure_TOL" influx:"Barrel-2_V-P_Pressure_TOL"`
	Barrel2HoldPressurePostionCCL       string  `json:"Barrel-2_HoldPressure_Postion_CCL" influx:"Barrel-2_HoldPressure_Postion_CCL"`
	Barrel2HoldPressurePostionTOL       string  `json:"Barrel-2_HoldPressure_Postion_TOL" influx:"Barrel-2_HoldPressure_Postion_TOL"`
	Barrel2PeakHoldPressureCCL          string  `json:"Barrel-2_PeakHoldPressure_CCL" influx:"Barrel-2_PeakHoldPressure_CCL"`
	Barrel2PeakHoldPressureTOL          string  `json:"Barrel-2_PeakHoldPressure_TOL" influx:"Barrel-2_PeakHoldPressure_TOL"`
	Barrel2ChargeTimeCCL                string  `json:"Barrel-2_ChargeTime_CCL" influx:"Barrel-2_ChargeTime_CCL"`
	Barrel2ChargeTimeTOL                string  `json:"Barrel-2_ChargeTime_TOL" influx:"Barrel-2_ChargeTime_TOL"`
	Barrel2ChargeStartPositionCCL       string  `json:"Barrel-2_ChargeStartPosition_CCL" influx:"Barrel-2_ChargeStartPosition_CCL"`
	Barrel2ChargeStartPositionTOL       string  `json:"Barrel-2_ChargeStartPosition_TOL" influx:"Barrel-2_ChargeStartPosition_TOL"`
	Barrel2WholeInjectionTimeCCL        string  `json:"Barrel-2_Whole_Injection_Time_CCL" influx:"Barrel-2_Whole_Injection_Time_CCL"`
	Barrel2WholeInjectionTimeTOL        string  `json:"Barrel-2_Whole_Injection_Time_TOL" influx:"Barrel-2_Whole_Injection_Time_TOL"`
	Barrel2ScrewRPMCCL                  string  `json:"Barrel-2_Screw_RPM_CCL" influx:"Barrel-2_Screw_RPM_CCL"`
	Barrel2ScrewRPMTOL                  string  `json:"Barrel-2_Screw_RPM_TOL" influx:"Barrel-2_Screw_RPM_TOL"`
	Barrel3InjectFillTimeCCL            string  `json:"Barrel-3_InjectFillTime_CCL" influx:"Barrel-3_InjectFillTime_CCL"`
	Barrel3InjectFillTimeTOL            string  `json:"Barrel-3_InjectFillTime_TOL" influx:"Barrel-3_InjectFillTime_TOL"`
	Barrel3CushionPositionCCL           string  `json:"Barrel-3_CushionPosition_CCL" influx:"Barrel-3_CushionPosition_CCL"`
	Barrel3CushionPositionTOL           string  `json:"Barrel-3_CushionPosition_TOL" influx:"Barrel-3_CushionPosition_TOL"`
	Barrel3PeakInjectPressureCCL        string  `json:"Barrel-3_PeakInjectPressure_CCL" influx:"Barrel-3_PeakInjectPressure_CCL"`
	Barrel3PeakInjectPressureTOL        string  `json:"Barrel-3_PeakInjectPressure_TOL" influx:"Barrel-3_PeakInjectPressure_TOL"`
	Barrel3VPPositionCCL                string  `json:"Barrel-3_V-P_Position_CCL" influx:"Barrel-3_V-P_Position_CCL"`
	Barrel3VPPositionTOL                string  `json:"Barrel-3_V-P_Position_TOL" influx:"Barrel-3_V-P_Position_TOL"`
	Barrel3VPPressureCCL                string  `json:"Barrel-3_V-P_Pressure_CCL" influx:"Barrel-3_V-P_Pressure_CCL"`
	Barrel3VPPressureTOL                string  `json:"Barrel-3_V-P_Pressure_TOL" influx:"Barrel-3_V-P_Pressure_TOL"`
	Barrel3HoldPressurePostionCCL       string  `json:"Barrel-3_HoldPressure_Postion_CCL" influx:"Barrel-3_HoldPressure_Postion_CCL"`
	Barrel3HoldPressurePostionTOL       string  `json:"Barrel-3_HoldPressure_Postion_TOL" influx:"Barrel-3_HoldPressure_Postion_TOL"`
	Barrel3PeakHoldPressureCCL          string  `json:"Barrel-3_PeakHoldPressure_CCL" influx:"Barrel-3_PeakHoldPressure_CCL"`
	Barrel3PeakHoldPressureTOL          string  `json:"Barrel-3_PeakHoldPressure_TOL" influx:"Barrel-3_PeakHoldPressure_TOL"`
	Barrel3ChargeTimeCCL                string  `json:"Barrel-3_ChargeTime_CCL" influx:"Barrel-3_ChargeTime_CCL"`
	Barrel3ChargeTimeTOL                string  `json:"Barrel-3_ChargeTime_TOL" influx:"Barrel-3_ChargeTime_TOL"`
	Barrel3ChargeStartPositionCCL       string  `json:"Barrel-3_ChargeStartPosition_CCL" influx:"Barrel-3_ChargeStartPosition_CCL"`
	Barrel3ChargeStartPositionTOL       string  `json:"Barrel-3_ChargeStartPosition_TOL" influx:"Barrel-3_ChargeStartPosition_TOL"`
	Barrel3WholeInjectionTimeCCL        string  `json:"Barrel-3_Whole_Injection_Time_CCL" influx:"Barrel-3_Whole_Injection_Time_CCL"`
	Barrel3WholeInjectionTimeTOL        string  `json:"Barrel-3_Whole_Injection_Time_TOL" influx:"Barrel-3_Whole_Injection_Time_TOL"`
	Barrel3ScrewRPMCCL                  string  `json:"Barrel-3_Screw_RPM_CCL" influx:"Barrel-3_Screw_RPM_CCL"`
	Barrel3ScrewRPMTOL                  string  `json:"Barrel-3_Screw_RPM_TOL" influx:"Barrel-3_Screw_RPM_TOL"`
	Timestamp                           int64   `json:"time" influx:"timestamp"`
	InfluxMeasurement                   influxdbhelper.Measurement
	Time                                time.Time `json:"-" influx:"time"`
}
