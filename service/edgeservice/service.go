package edgeservice

import (
	"FSRV_Edge/global"
	"FSRV_Edge/init/initlog"
	"FSRV_Edge/nodeattr"
	"FSRV_Edge/service/opcuaclient"
	"FSRV_Edge/service/opcuaservice"
	"FSRV_Edge/service/routerservice"
	"FSRV_Edge/service/wiseservice"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/gopcua/opcua"
	"github.com/mostlygeek/arp"
)

var (
	nodeStageMap   = make(map[string]string)
	hasRecordLevel = make(map[string]bool)
	// AllNodeData 存放所有節點即時資料
	AllDcData             dcData
	childBrowseArr        []string
	nodeValInfo           valueMap
	defaultPermission     nodeattr.AuthPermissopn
	defaultAccountNameArr = []string{rootAccount, adminAccount}
	globalLogger          = initlog.GetLogger()
	DefaultNode           DefaultNodeData
	NameInfo              NameData
	hasReadNode           readRecord
)

type DefaultNodeData struct {
	DataMap map[string]interface{}
	Lock    sync.Mutex
}
type readRecord struct {
	DataMap map[string]bool
	Lock    sync.Mutex
}
type dcData struct {
	DataMap map[string][]NodeStruct
	Lock    sync.Mutex
}

const (
	dcMacStr        = "00D0C9"
	abnormalStr     = "error validating input"
	edgeAuthTable   = "edge_auth"
	permissionTable = "auth_permission"
	rootAccount     = "root"
	adminAccount    = "admin"
	identErrStr     = "user identity token is valid"
	deviceStr       = "Device"
	runMsg          = 1
	stopMsg         = 2
	conMsg          = 1
	disConMsg       = 2
	unknowMsg       = 3
	abnormalMsg     = 0
	defaultAct      = 2
	uniqueErr       = "UNIQUE constraint failed"
	nodeParentExcel = "template/node_parent.xlsx"
)

type NameData struct {
	DataMap map[string]string
	Lock    sync.Mutex
}

func writeExcel(xlsx *excelize.File, idx string) {
	t := time.Now()
	index := xlsx.NewSheet("Sheet1")
	if idx == "1" {
		xlsx.SetCellValue("Sheet1", "A"+idx, "CPU Used")
		xlsx.SetCellValue("Sheet1", "B"+idx, "Mem Used")
		xlsx.SetCellValue("Sheet1", "C"+idx, "Disk Used")
		xlsx.SetCellValue("Sheet1", "D"+idx, "Time")
		xlsx.SetActiveSheet(index)
		// Save xlsx file by the given path.
		xlsx.SaveAs("./egc_usedpercent" + strconv.Itoa(int(t.Month())) + strconv.Itoa(t.Day()) + ".xlsx")
		return
	}
	// Create a new sheet.
	ram := PrintMemUsage()
	v := strconv.Itoa(t.Hour()) + ":" + strconv.Itoa(t.Minute()) + ":" + strconv.Itoa(t.Second())
	// Set value of a cell.
	// xlsx.SetCellValue("Sheet1", "A"+idx, cpuTime)
	xlsx.SetCellValue("Sheet1", "B"+idx, ram)
	// xlsx.SetCellValue("Sheet1", "C"+idx, diskTime)
	xlsx.SetCellValue("Sheet1", "D"+idx, v)
	// Set active sheet of the workbook.
	xlsx.SetActiveSheet(index)
	// Save xlsx file by the given path.
	xlsx.SaveAs("./egc_usedpercent" + strconv.Itoa(int(t.Month())) + strconv.Itoa(t.Day()) + ".xlsx")
}
func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}
func PrintMemUsage() uint64 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	// For info on each, see: https://golang.org/pkg/runtime/#MemStats
	return bToMb(m.Alloc)
}
func init() {
	DefaultNode.Lock.Lock()
	DefaultNode.DataMap = make(map[string]interface{})
	DefaultNode.Lock.Unlock()
	hasReadNode.Lock.Lock()
	hasReadNode.DataMap = make(map[string]bool)
	hasReadNode.Lock.Unlock()
	global.Devs, _ = routerservice.GetAllDevInfo()
	NameInfo.Lock.Lock()
	NameInfo.DataMap = make(map[string]string)
	NameInfo.DataMap["admin"] = "管理者"
	NameInfo.DataMap["root"] = "root"
	NameInfo.Lock.Unlock()
	AllDcData.Lock.Lock()
	AllDcData.DataMap = make(map[string][]NodeStruct)
	AllDcData.Lock.Unlock()
	defer func() {
		if r := recover(); r != nil {
			globalLogger.Criticalf("[Init edge service error]: %v", r)
		}
	}()
	exist := make(map[string]bool)
	opcuaclient.GlobalNodeParent.Init()
	if f, err := excelize.OpenFile(nodeParentExcel); err != nil {
		panic("load node parent error: " + err.Error())
	} else {
		rows := f.GetRows("工作表1")
		for i := 0; i < len(rows); i++ {
			row := rows[i]
			if len(row) >= 3 {
				opcuaclient.GlobalNodeParent.SetMap(row[1], row[2])
				childBrowseArr = append(childBrowseArr, row[1])
				DefaultNode.DataMap[row[1]] = ""
				if !exist[row[2]] {
					exist[row[2]] = true
					opcuaclient.ParentBrowseArr = append(opcuaclient.ParentBrowseArr, row[2])
				}
			}
		}
	}

	if err := insertDefaultAccount(); err != nil {
		panic("Insert default account error " + err.Error())
	} else {
		if err := insertDefaultPermission(); err != nil {
			panic("Insert default permission error " + err.Error())
		}
	}
	conv, _, _ := GetDefaultConv("")
	setValMap(conv)
}

func insertDefaultAccount() error {
	if size, err := routerservice.GetTableSize(edgeAuthTable); err == nil {
		if size == 0 {
			var defaultAccountArr []nodeattr.EdgeAuth
			NameInfo.Lock.Lock()
			nameInfo := NameInfo.DataMap
			NameInfo.Lock.Unlock()
			for _, name := range defaultAccountNameArr {
				var tmp nodeattr.EdgeAuth
				tmp.Account = name
				tmp.Password = name
				tmp.Name = nameInfo[name]
				defaultAccountArr = append(defaultAccountArr, tmp)
			}
			if err := routerservice.InsertEdgeAuth(defaultAccountArr); err != nil {
				return err
			}

		}
	}
	return nil
}
func insertDefaultPermission() error {
	if size, err := routerservice.GetTableSize(permissionTable); err == nil {
		if size == 0 {
			var tmpArr []nodeattr.AuthPermissopn
			for _, name := range defaultAccountNameArr {
				if account, err := routerservice.GetEdgeAuthByName(name); err == nil {
					for _, node := range childBrowseArr {
						var tmp nodeattr.AuthPermissopn
						tmp.AccountID = account.ID
						tmp.BrowseName = node
						tmp.Action = defaultAct
						tmpArr = append(tmpArr, tmp)
					}
				} else {
					return err
				}
			}
			if err := routerservice.InsertPermission(tmpArr); err != nil {
				return err
			}
		}
	}
	return nil
}

// UpdateConStatus 定時更新連線、設備狀態
func UpdateConStatus() {
	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		if r := recover(); r != nil {
			cancel()
			globalLogger.Criticalf("[UpdateConStatus error]:%v", r)
		}
	}()
	i := 1
	// xlsx := excelize.NewFile()
	for {
		// idx := strconv.Itoa(i)
		// writeExcel(xlsx, idx)
		global.Devs, _ = routerservice.GetAllDevInfo()
		checkConStatus(ctx)
		if global.PingWiseSetting == 1 {
			wiseservice.UpdateDcStatus()
		}
		wiseservice.CheckNowDcStatus()
		i++
		time.Sleep(30 * time.Second)
	}
}

func checkConStatus(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			globalLogger.Criticalf("[Check connect error] %v", r)
		}
	}()
	var dev nodeattr.DevInfo
	getConv := make(map[string]bool)
	conArr, err := routerservice.GetAllConInfo()
	if err != nil {
		panic("GetAllConInfo" + err.Error())
	}
	for _, con := range conArr {
		if con.Protocol != global.OpcStr {
			continue
		}
		var endpoint string
		if con.Port == "80" {
			endpoint = nodeattr.OpcTitle + con.IP
		} else {
			endpoint = nodeattr.OpcTitle + con.IP + ":" + con.Port
		}
		c, conErr := opcuaclient.ConnectServer(ctx, con.Account, con.Password, endpoint)
		if conErr != nil { // 無法找到連線
			dev.ConID = con.ID
			if strings.Contains(conErr.Error(), abnormalStr) || strings.Contains(conErr.Error(), identErrStr) {
				// 帳密錯誤
				con.Status = abnormalMsg
				dev.Status = abnormalMsg
			} else {
				con.Status = disConMsg
				dev.Status = disConMsg
			}
			routerservice.UpdateDevInfoByConID(dev)
		} else { // 可連到連線
			dev.Protocol = global.OpcStr
			dev.ConName = con.Name
			dev.ConID = con.ID
			con.Status = conMsg
			con.Timestamp = time.Now().UnixNano() / int64(time.Millisecond)
			isExist := false
			var idStr string
			for _, v := range global.Devs {
				if v.Protocol == global.OpcStr {
					idStr = strconv.FormatInt(v.ID, 10)
					opcuaclient.HasBrowse.Lock.Lock()
					hasBrowse := opcuaclient.HasBrowse.DataMap[idStr]
					opcuaclient.HasBrowse.Lock.Unlock()
					if !hasBrowse { // 檢查是否browse過，沒有則先browse
						opcuaclient.BrowseAll(c, idStr)
					}
					if v.ConID == con.ID { //連線已存在DB
						dev.TempID = v.TempID
						isExist = true
						break
					}
				}
			} // for
			if isExist {
				if devs, err := routerservice.GetDevInfoByConID(con.ID); err == nil {
					for _, v := range devs {
						dev.ID = v.ID
						devName := "Device" + idStr
						// TODO 檢查build server是否會造成錯誤
						if global.BuildNodeSetting == 1 {
							opcuaservice.BuildAllDevNode(devName)
						}
					}
				}

				if dev.TempID == 0 { // 新增的設備，帶入預設範本
					dev.TempID = routerservice.DefaultTemp.ID
					dev.TempName = routerservice.DefaultTemp.Name
				} else {
					devID := strconv.FormatInt(dev.ID, 10)
					brand, _ := opcuaclient.ReadNode(c, nodeattr.BrandStr, devID)
					status, _ := getStatus(c, devID)
					dev.Brand = brand
					dev.Status = status
					if !getConv[devID] {
						getConv[devID] = true
						UpdateConv(devID)
					}
					hasReadNode.Lock.Lock()
					hasRead := hasReadNode.DataMap[devID]
					hasReadNode.Lock.Unlock()
					if !hasRead {
						opcuaclient.ReadDevAllNode(c, devID)
						hasReadNode.Lock.Lock()
						hasReadNode.DataMap[devID] = true
						hasReadNode.Lock.Unlock()
					}
					opcuaclient.GetMonitorItem(con, devID)
				}
			} else {
				dev.TempID = routerservice.DefaultTemp.ID
				dev.TempName = routerservice.DefaultTemp.Name
				dev.Mac = con.IP
				routerservice.InsertDevInfo(dev)
			}
			if len(global.Devs) == 0 {
				dev.TempID = routerservice.DefaultTemp.ID
				dev.TempName = routerservice.DefaultTemp.Name
				routerservice.InsertDevInfo(dev)
			}
			routerservice.UpdateDevInfo(dev)
			c.Close()
		} // else
		routerservice.UpdateConnectInfo(con)
		global.Devs, _ = routerservice.GetAllDevInfo()
	}
}

// GetDefaultConv 取得預設轉換表
func GetDefaultConv(devName string) ([]nodeattr.Converter, int64, error) {
	var convArr []nodeattr.Converter
	var conv nodeattr.Converter
	var id int64
	if tmp, err := routerservice.GetTemplateByName(nodeattr.DefaultTmpName); err != nil {
		return convArr, id, err
	} else {
		routerservice.DefaultTemp.ID = tmp.ID
		id = tmp.ID
		if tmpRel, err := routerservice.GetTemplateRel(tmp.ID); err != nil {
			return convArr, id, err
		} else {
			for _, v := range tmpRel {
				conv.ConvFunc = v.ConvFunc
				conv.DstBrowse = v.DstBrowse
				conv.DstNodeid = v.DstNodeid
				conv.RefBrowseName1 = v.RefBrowseName1
				conv.RefBrowseName2 = v.RefBrowseName2
				conv.RefBrowseName3 = v.RefBrowseName3
				conv.ScrDevName = devName
				conv.SrcBrowse = v.SrcBrowse
				conv.SrcNamespace = v.SrcNamespace
				conv.SrcNodeid = v.SrcNodeid
				conv.SrcUnit = v.SrcUnit
				convArr = append(convArr, conv)
			}
		}
	}
	return convArr, id, nil
}

func getStatus(c *opcua.Client, devID string) (int, error) {
	c1, err := opcuaclient.ReadNode(c, nodeattr.CycleTimeStr, devID)
	if err != nil {
		return unknowMsg, err
	}
	time.Sleep(500 * time.Millisecond)
	c2, err := opcuaclient.ReadNode(c, nodeattr.CycleTimeStr, devID)
	if err != nil {
		return unknowMsg, err
	}
	if c1 == c2 {
		return stopMsg, err
	}
	return runMsg, err
}

// NodeStruct 節點資點
type NodeStruct struct {
	Name  string      `json:"name"`
	Value interface{} `json:"value"`
}
type valueMap struct {
	dataMap  map[string]interface{}
	levelMap map[string]string
	lock     sync.Mutex
}

// 設定預設LEVEL欄位顯示順序Map
func setValMap(convInfos []nodeattr.Converter) {
	nodeValInfo.lock.Lock()
	nodeValInfo.dataMap = make(map[string]interface{})
	nodeValInfo.levelMap = make(map[string]string)
	nodeValInfo.lock.Unlock()
	tmpMap := opcuaclient.GlobalNodeParent.GetMap()
	nodeValInfo.lock.Lock()
	for _, conv := range convInfos {
		level := conv.GetLevel()
		key := tmpMap[conv.DstBrowse] // key是 dstBrowse 的 parent
		count := opcuaclient.GlobalNodeParent.GetCount(key)
		valStr := ""
		nodeValInfo.levelMap[key] = level
		if count == 1 {
			nodeValInfo.dataMap[key] = valStr
		} else {
			if nodeValInfo.dataMap[key] == nil {
				nodeValInfo.dataMap[key] = make([]string, 0)
			}
			nodeValInfo.dataMap[key] = append(nodeValInfo.dataMap[key].([]string), valStr)
		}
	}
	nodeValInfo.lock.Unlock()
}

func UpdateConv(devID string) {
	defer func() {
		if r := recover(); r != nil {
			globalLogger.Criticalf("[updateConv error]: %v", r)
		}
	}()
	convInfos, err := routerservice.GetOriginConv("Device" + devID)
	if err != nil {
		panic("GetDataConv error" + err.Error())
	}
	var dev nodeattr.DevInfo
	for _, v := range global.Devs {
		idStr := strconv.Itoa(int(v.ID))
		if idStr == devID {
			dev = v
			break
		}
	}
	if len(convInfos) == 0 {
		var rels []nodeattr.TemplateRel
		tmpRels, _ := routerservice.GetTempRel(dev.TempID)
		if dev.TempID == routerservice.DefaultTemp.ID {
			// 設備當前是套用預設範本
			rels = opcuaservice.DefaultRels
			for _, v := range rels {
				if len(v.DstBrowse) == 0 {
					continue
				}
				var con nodeattr.Converter
				con.ID = v.ID
				con.ConvFunc = v.ConvFunc
				con.DstBrowse = v.DstBrowse
				con.SrcNamespace = 1
				con.SrcNodeid = v.DstBrowse
				con.DstNodeid = "Deivce" + devID + "." + v.DstNodeid
				con.RefBrowseName1 = v.RefBrowseName1
				con.RefBrowseName2 = v.RefBrowseName2
				con.RefBrowseName3 = v.RefBrowseName3
				convInfos = append(convInfos, con)
			}
		} else {
			// 設備當前不是套用預設範本
			rels = tmpRels
			for _, v := range rels {
				if len(v.DstBrowse) == 0 {
					continue
				}
				var con nodeattr.Converter
				con.ID = v.ID
				con.ConvFunc = v.ConvFunc
				con.DstBrowse = v.DstBrowse
				// if v.SrcNamespace == 0 {
				// 	con.SrcNamespace = 1
				// } else {
				// 	con.SrcNamespace = v.SrcNamespace
				// }
				// if len(v.SrcNodeid) == 0 {
				// 	con.SrcNodeid = v.DstBrowse
				// } else {
				// 	con.SrcNodeid = v.SrcNodeid
				// }
				con.SrcNamespace = v.SrcNamespace
				con.SrcNodeid = v.SrcNodeid
				con.SrcBrowse = v.SrcBrowse
				con.DstNodeid = "Deivce" + devID + "." + v.DstNodeid
				con.RefBrowseName1 = v.RefBrowseName1
				con.RefBrowseName2 = v.RefBrowseName2
				con.RefBrowseName3 = v.RefBrowseName3
				convInfos = append(convInfos, con)
			}
		}

	} else {
		convInfos, err = routerservice.GetDataConv("Device"+devID, 0)
		if err != nil {
			panic("GetDataConv error" + err.Error())
		}
		// for i := 0; i < len(convInfos); i++ {
		// 	if len(convInfos[i].SrcNodeid) == 0 {
		// 		convInfos[i].SrcNodeid = convInfos[i].DstBrowse
		// 	}
		// 	if convInfos[i].SrcNamespace == 0 {
		// 		convInfos[i].SrcNamespace = 1
		// 	}
		// }
	}
	opcuaclient.AllNodeDataConv.Lock.Lock()
	opcuaclient.AllNodeDataConv.DataMap[devID] = convInfos
	opcuaclient.AllNodeDataConv.Lock.Unlock()
	// boolMap := make(map[interface{}]int)
	// boolMap["true"] = 1
	// boolMap["false"] = 0
	// if _, ok := hasRecordLevel[devID]; !ok {
	// 	AllNodeData.Lock.Lock()
	// 	hasRecordLevel[devID] = true
	// 	AllNodeData.DataMap = make(map[string]map[string][]NodeStruct)
	// 	AllNodeData.DataMap[devID] = make(map[string][]NodeStruct)
	// 	AllNodeData.Lock.Unlock()
	// }
	// var l0NodeArr []NodeStruct
	// var l1NodeArr []NodeStruct
	// var l2NodeArr []NodeStruct
	// var l3NodeArr []NodeStruct
	// var l4NodeArr []NodeStruct
	// var l5NodeArr []NodeStruct
	// var tmp valueMap
	// tmp.dataMap = make(map[string]interface{})
	// tmp.levelMap = make(map[string]string)
	// tmpMap := GlobalNodeParent.GetMap()
	// for _, conv := range convInfos {
	// 	level := conv.GetLevel()
	// 	valStr, _ := opcuaclient.ReadNode(c, conv.DstBrowse, devID)
	// 	key := tmpMap[conv.DstBrowse] // key是 dstBrowse 的 parent
	// 	count := GlobalNodeParent.GetCount(key)
	// 	tmp.levelMap[key] = level
	// 	if count == 1 {
	// 		tmp.dataMap[key] = valStr
	// 	} else {
	// 		if tmp.dataMap[key] == nil {
	// 			tmp.dataMap[key] = make([]string, 0)
	// 		}
	// 		tmp.dataMap[key] = append(tmp.dataMap[key].([]string), valStr)
	// 	}
	// }
	// for _, key := range parentBrowseArr {
	// 	var tmpT NodeStruct
	// 	tmpT.Name = key
	// 	tmpT.Value = tmp.dataMap[key]
	// 	if tmpT.Name == nodeattr.MotorStr {
	// 		tmpT.Value = boolMap[tmpT.Value]
	// 	} else if tmpT.Name == nodeattr.StatusStr {
	// 		if tmpT.Value != nil {
	// 			if m, err := strconv.Atoi(tmpT.Value.(string)); err != nil {
	// 				tmpT.Value = ""
	// 			} else {
	// 				tmpT.Value = m
	// 			}
	// 		}
	// 	} else if tmpT.Name == nodeattr.MoldCountStr {
	// 		if tmpT.Value != nil {
	// 			if m, err := strconv.Atoi(tmpT.Value.(string)); err != nil {
	// 				tmpT.Value = 0
	// 			} else {
	// 				tmpT.Value = m
	// 			}
	// 		}
	// 	}

	// 	if tmp.levelMap[key] == nodeattr.Level2Str {
	// 		l2NodeArr = append(l2NodeArr, tmpT)
	// 	} else if tmp.levelMap[key] == nodeattr.Level3Str {
	// 		l3NodeArr = append(l3NodeArr, tmpT)
	// 	} else if tmp.levelMap[key] == nodeattr.Level0Str {
	// 		l0NodeArr = append(l0NodeArr, tmpT)
	// 	} else if tmp.levelMap[key] == nodeattr.Level1Str {
	// 		l1NodeArr = append(l1NodeArr, tmpT)
	// 	} else if tmp.levelMap[key] == nodeattr.Level4Str {
	// 		l4NodeArr = append(l4NodeArr, tmpT)
	// 	} else if tmp.levelMap[key] == nodeattr.Level5Str {
	// 		l5NodeArr = append(l5NodeArr, tmpT)
	// 	}
	// }
	// nodeValInfo.lock.Lock()
	// nodeValInfo.dataMap = tmp.dataMap
	// nodeValInfo.lock.Unlock()
	// AllNodeData.Lock.Lock()
	// AllNodeData.DataMap[devID][nodeattr.Level0Str] = l0NodeArr
	// AllNodeData.DataMap[devID][nodeattr.Level1Str] = l1NodeArr
	// AllNodeData.DataMap[devID][nodeattr.Level2Str] = l2NodeArr
	// AllNodeData.DataMap[devID][nodeattr.Level3Str] = l3NodeArr
	// AllNodeData.DataMap[devID][nodeattr.Level4Str] = l4NodeArr
	// AllNodeData.DataMap[devID][nodeattr.Level5Str] = l5NodeArr
	// AllNodeData.Lock.Unlock()
}

// GetLevelHDA 取得對應level歷史資料
func GetLevelHDA(level string, hdaData []map[string]interface{}) (tmpArr []NodeStruct) {
	defer func() {
		if r := recover(); r != nil {
			globalLogger.Criticalf("[Get level data error]:%v", r)
		}
	}()
	var nodeValTmpInfo valueMap
	tmpMap := opcuaclient.GlobalNodeParent.GetMap()
	nodeValInfo.lock.Lock()
	nodeValTmpInfo.levelMap = nodeValInfo.levelMap
	nodeValInfo.lock.Unlock()
	for _, val := range hdaData {
		nodeValTmpInfo.dataMap = make(map[string]interface{})
		for _, browse := range childBrowseArr {
			key := tmpMap[browse] // key是 dstBrowse 的 parent
			count := opcuaclient.GlobalNodeParent.GetCount(key)
			if nodeValTmpInfo.levelMap[key] == level {
				// 屬於對應level的資料
				if count == 1 {
					nodeValTmpInfo.dataMap[key] = val[browse]
				} else {
					if nodeValTmpInfo.dataMap[key] == nil {
						nodeValTmpInfo.dataMap[key] = make([]string, 0)
					}
					if val[browse] == nil {
						nodeValTmpInfo.dataMap[key] = append(nodeValTmpInfo.dataMap[key].([]string), "")
					} else {
						nodeValTmpInfo.dataMap[key] = append(nodeValTmpInfo.dataMap[key].([]string), val[browse].(string))
					}
				}
			}
		}
		for _, key := range opcuaclient.ParentBrowseArr {
			if nodeValTmpInfo.levelMap[key] != level {
				continue
			}
			var tmp NodeStruct
			tmp.Name = key
			tmp.Value = nodeValTmpInfo.dataMap[key]

			if tmp.Name == nodeattr.MotorStr {
				if tmp.Value == nil {
					tmp.Value = ""
				} else {
					switch tmp.Value.(type) {
					case nil:
						tmp.Value = ""
					case string:
						if m, err := strconv.Atoi(tmp.Value.(string)); err == nil {
							tmp.Value = m
						}
					case int:
						tmp.Value = tmp.Value.(int)
					case json.Number:
						if m, err := strconv.Atoi(tmp.Value.(json.Number).String()); err == nil {
							tmp.Value = m
						}
					default:
						tmp.Value = ""
					}
				}
			} else if tmp.Name == nodeattr.StatusStr {
				if tmp.Value == nil {
					tmp.Value = ""
				} else {
					switch tmp.Value.(type) {
					case nil:
						tmp.Value = ""
					case string:
						if m, err := strconv.Atoi(tmp.Value.(string)); err == nil {
							tmp.Value = m
						}
					case int:
						tmp.Value = tmp.Value.(int)
					default:
						tmp.Value = ""
					}
				}
			}
			tmpArr = append(tmpArr, tmp)
		}
	}
	return
}
func CheckIPAddressType(ip string) error {
	ipRange := strings.Split(ip, "-")
	if len(ipRange) == 2 {
		if _, err := strconv.Atoi(ipRange[1]); err != nil {
			return err
		}
		ip = ipRange[0]
		if net.ParseIP(ip) == nil {
			err := errors.New("Invalid IP Address")
			return err
		}
		for i := 0; i < len(ip); i++ {
			switch ip[i] {
			case '.':
				return nil
			}
		}
	} else {
		if net.ParseIP(ip) == nil {
			err := errors.New("Invalid IP Address")
			return err
		}
		for i := 0; i < len(ip); i++ {
			switch ip[i] {
			case '.':
				return nil
			}
		}
	}
	return nil
}

// ScanIP 掃描網域下IP
func ScanIP(ipRange string) (ipArr []nodeattr.ConInfo) {
	p1 := "80"
	p2 := "4840"
	timeout := time.Duration(50 * time.Millisecond)
	ips := getAllIP(ipRange)
	m := make(map[string]bool)
	for _, v := range global.Cons {
		if v.Exist != 3 {
			ipArr = append(ipArr, v)
		}
		m[v.IP] = true
	}
	for _, v := range ips {
		if m[v] {
			continue
		}
		var tmp nodeattr.ConInfo
		_, err := net.DialTimeout("tcp", v+":"+p1, timeout)
		if err == nil {
			tmp.Exist = 0
			tmp.IP = v
			tmp.Port = p1
			ipArr = append(ipArr, tmp)
		}
		_, err = net.DialTimeout("tcp", v+":"+p2, timeout)
		if err == nil {
			tmp.Exist = 0
			tmp.IP = v
			tmp.Port = p2
			ipArr = append(ipArr, tmp)
		}
	}
	return
}
func getIp(con string) (string, int, int) {
	var start int
	var end int
	var midVal string
	ipRange := strings.Split(con, "-")
	if len(ipRange) == 2 {
		if v1, err := strconv.Atoi(ipRange[1]); err == nil {
			ipIdx := strings.Split(ipRange[0], ".")
			if len(ipIdx) == 4 {
				if v2, err := strconv.Atoi(ipIdx[3]); err == nil {
					if v1 > v2 {
						end = v1
						start = v2
						midVal = ipIdx[0] + "." + ipIdx[1] + "." + ipIdx[2]
					} else if v1 < v2 {
						end = v2
						start = v1
						midVal = ipIdx[0] + "." + ipIdx[1] + "." + ipIdx[2]
					} else {
						midVal = ipRange[0]
					}
					return midVal, start, end
				}
			}
		}
	}
	return con, start, end
}

// getAllIP 取得指定網域IP
func getAllIP(ipRange string) (ipVal []string) {
	midVal, st, end := getIp(ipRange)
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("arp error ", err)
		}
	}()
	if err := exec.Command("arp", "-da").Run(); err != nil {
		fmt.Println("arp:" + err.Error())
	}
	arp.CacheUpdate()

	var target string
	if st == end {
		target = midVal
	}
	for ip := range arp.Table() {
		mac := strings.Replace(arp.Search(ip), ":", "", -1)
		mac = strings.ToUpper(mac)
		ipIdx := strings.Split(ip, ".")
		startsWith := strings.HasPrefix(mac, dcMacStr)
		if !startsWith {
			tmp := ipIdx[0] + "." + ipIdx[1] + "." + ipIdx[2]
			if ip == target {
				ipVal = append(ipVal, ip)
				return
			} else if tmp == midVal {
				if v, err := strconv.Atoi(ipIdx[3]); err == nil {
					if v >= st && v <= end {
						ipVal = append(ipVal, ip)
					}
				}
			}
		}
	}
	return
}
