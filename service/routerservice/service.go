package routerservice

import (
	"FSRV_Edge/dao/opcdao"
	"FSRV_Edge/dao/settingdao"
	"FSRV_Edge/global"
	"FSRV_Edge/init/initlog"
	"FSRV_Edge/nodeattr"
	"FSRV_Edge/service/opcuaservice"
	"crypto/sha512"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/Knetic/govaluate"
	"github.com/astaxie/beego/orm"
	"github.com/jackpal/gateway"
)

var (
	globalLogger = initlog.GetLogger()
	usedMap      = make(map[string]bool)
	// DefaultRels 預設關聯表節點資料
	DefaultRels []nodeattr.TemplateRel
	// DefaultTemp 預設範本
	DefaultTemp nodeattr.Template
	levelArr    = []string{"LEVEL0", "LEVEL1", "LEVEL2", "LEVEL3", "LEVEL4", "LEVEL5"}
)

const (
	startRows  = 2
	noTableStr = "no such table"
	stageStr   = "_Stage"
)

type nodeMap struct {
	dataMap map[string][]nodeattr.TemplateRel
	lock    sync.Mutex
}

// Machine 設備資料
type Machine struct {
	Name int
}

func init() {
	size, err := GetTableSize("node_template_rel")
	if err != nil {
		fmt.Println("[GetTableSize error]", err)
	}
	opcuaservice.NodeDataMap.Lock.Lock()
	nodeMap := opcuaservice.NodeDataMap.DataMap
	opcuaservice.NodeDataMap.Lock.Unlock()
	DefaultTemp.Name = nodeattr.DefaultTmpName
	DefaultTemp.CreateTime = time.Now().UnixNano() / int64(time.Millisecond)
	DefaultTemp.ModifyTime = DefaultTemp.CreateTime
	DefaultTemp.SystemCreate = true
	InsertTemplate(DefaultTemp)
	tmpArr, err := GetAllTemplate()
	if err != nil {
		fmt.Println("[GetAllTemplate error]", err)
	} else {
		for _, v := range tmpArr {
			if v.SystemCreate {
				DefaultTemp.ID = v.ID
			}
		}
	}
	if size == 0 {
		for _, v := range levelArr {
			for _, rel := range nodeMap[v] {
				DefaultRels = append(DefaultRels, rel)
			}
			InsertTemplateRel(nodeMap[v], DefaultTemp.ID)
		}
	} else {
		if len(DefaultRels) == 0 {
			for _, v := range levelArr {
				for _, rel := range nodeMap[v] {
					DefaultRels = append(DefaultRels, rel)
				}
			}
		}
	}
	if v, err := gateway.DiscoverGateway(); err == nil {
		var con nodeattr.ConInfo
		con.IP = v.String() + "-254"
		con.Name = "LAN"
		con.Status = 1
		con.Exist = 3
		con.Timestamp = time.Now().UnixNano() / int64(time.Millisecond)
		con.Protocol = global.HttpStr
		InsertConInfo(con)
		UpdateConnectInfo(con)
	}
}

// GetTableSize 取得table資料數
func GetTableSize(name string) (size int64, err error) {
	o := orm.NewOrm()
	o.Begin()
	defer func() {
		if r := recover(); r != nil {
			o.Rollback()
			switch x := r.(type) {
			case string:
				err = errors.New(x)
			case error:
				err = x
			default:
				err = errors.New("Unknown panic")
			}
		} else {
			o.Commit()
		}
	}()
	size, err = opcdao.GetTableSize(name, o)
	return
}

// GetTemplateByName 取得範本
func GetTemplateByName(name string) (tmp nodeattr.Template, err error) {
	o := orm.NewOrm()
	o.Begin()
	defer func() {
		if r := recover(); r != nil {
			o.Rollback()
			switch x := r.(type) {
			case string:
				err = errors.New(x)
			case error:
				err = x
			default:
				err = errors.New("Unknown panic")
			}
		} else {
			o.Commit()
		}
	}()
	tmp, err = opcdao.GetTemplateByName(name, o)
	return
}

// GetTemplateRel 取得範本關聯
func GetTemplateRel(tempID int64) (tmpRels []nodeattr.TemplateRel, err error) {
	o := orm.NewOrm()
	o.Begin()
	defer func() {
		if r := recover(); r != nil {
			o.Rollback()
			switch x := r.(type) {
			case string:
				err = errors.New(x)
			case error:
				err = x
			default:
				err = errors.New("Unknown panic")
			}
		} else {
			o.Commit()
		}
	}()
	tmpRels, err = opcdao.GetTempRel(tempID, o)
	return
}

// InsertTemplate 新增範本
func InsertTemplate(tmp nodeattr.Template) (id int64, err error) {
	o := orm.NewOrm()
	o.Begin()
	defer func() {
		if r := recover(); r != nil {
			o.Rollback()
			switch x := r.(type) {
			case string:
				err = errors.New(x)
			case error:
				err = x
			default:
				err = errors.New("Unknown panic")
			}
		} else {
			o.Commit()
		}
	}()
	id, err = opcdao.InsertTemplate(tmp, o)
	return
}

// UpdateTemplate 新增範本
func UpdateTemplate(temp nodeattr.EditTemplate) (err error) {
	o := orm.NewOrm()
	o.Begin()
	defer func() {
		if r := recover(); r != nil {
			o.Rollback()
			switch x := r.(type) {
			case string:
				err = errors.New(x)
			case error:
				err = x
			default:
				err = errors.New("Unknown panic")
			}
		} else {
			o.Commit()
		}
	}()
	err = opcdao.UpdateTemplate(temp, o)
	return
}

// DelTemplate 刪除範本
func DelTemplate(id int) (err error) {
	o := orm.NewOrm()
	o.Begin()
	defer func() {
		if r := recover(); r != nil {
			o.Rollback()
			switch x := r.(type) {
			case string:
				err = errors.New(x)
			case error:
				err = x
			default:
				err = errors.New("Unknown panic")
			}
		} else {
			o.Commit()
		}
	}()
	err = opcdao.DelTemplate(id, o)
	return
}

// DelTempRel 刪除範本關聯
func DelTempRel(tempRel nodeattr.EditTemplateRel) (err error) {
	o := orm.NewOrm()
	o.Begin()
	defer func() {
		if r := recover(); r != nil {
			o.Rollback()
			switch x := r.(type) {
			case string:
				err = errors.New(x)
			case error:
				err = x
			default:
				err = errors.New("Unknown panic")
			}
		} else {
			o.Commit()
		}
	}()
	err = opcdao.DelTempRel(tempRel, o)
	return
}

// GetDataConvDev 取得轉換表設備名稱
func GetDataConvDev() (devArr []string, err error) {
	o := orm.NewOrm()
	defer func() {
		if r := recover(); r != nil {
			switch x := r.(type) {
			case string:
				err = errors.New(x)
			case error:
				err = x
			default:
				err = errors.New("Unknown panic")
			}
		}
	}()
	if devs, err := opcdao.GetDataConvDev(o); err != nil {
		panic(err)
	} else {
		devArr = devs
	}
	return
}

// DelDataConvDev 刪除設備轉換表
func DelDataConvDev(devName string) (err error) {
	o := orm.NewOrm()
	defer func() {
		if r := recover(); r != nil {
			switch x := r.(type) {
			case string:
				err = errors.New(x)
			case error:
				err = x
			default:
				err = errors.New("Unknown panic")
			}
		}
	}()
	err = opcdao.DelAllDataConv(devName, o)
	return
}

// GetActConv 取得轉換表實際資料
func GetActConv(devName string) (convArr []nodeattr.Converter, err error) {
	o := orm.NewOrm()
	defer func() {
		if r := recover(); r != nil {
			switch x := r.(type) {
			case string:
				err = errors.New(x)
			case error:
				err = x
			default:
				err = errors.New("Unknown panic")
			}
		}
	}()
	if convs, err := opcdao.GetDataConv(devName, o); err != nil {
		panic(err)
	} else {
		convArr = convs
	}
	return
}

func GetOriginConv(devName string) (convArr []nodeattr.Converter, err error) {
	o := orm.NewOrm()
	defer func() {
		if r := recover(); r != nil {
			switch x := r.(type) {
			case string:
				err = errors.New(x)
			case error:
				err = x
			default:
				err = errors.New("Unknown panic")
			}
		}
	}()

	convArr, err = opcdao.GetDataConv(devName, o)
	return
}

// GetDataConv 取得轉換表
func GetDataConv(devName string, useTmpId int64) (convArr []nodeattr.Converter, err error) {
	o := orm.NewOrm()
	defer func() {
		if r := recover(); r != nil {
			switch x := r.(type) {
			case string:
				err = errors.New(x)
			case error:
				err = x
			default:
				err = errors.New("Unknown panic")
			}
		}
	}()

	if convInfos, err := opcdao.GetDataConv(devName, o); err != nil {
		panic(err)
	} else {
		tmpMap := make(map[string]nodeattr.Converter)
		for _, conv := range convInfos {
			tmpMap[conv.DstBrowse] = conv
		}
		for _, v := range opcuaservice.DefaultRels {
			var con nodeattr.Converter
			con.ID = v.ID
			con.ConvFunc = tmpMap[v.DstBrowse].ConvFunc
			con.DstBrowse = v.DstBrowse
			con.DstNodeid = devName + "." + v.DstNodeid
			con.RefBrowseName1 = tmpMap[v.DstBrowse].RefBrowseName1
			con.RefBrowseName2 = tmpMap[v.DstBrowse].RefBrowseName2
			con.RefBrowseName3 = tmpMap[v.DstBrowse].RefBrowseName3
			if tmpMap[v.DstBrowse].Modify == 0 && useTmpId == DefaultTemp.ID { // 設備欄位未修改過
				con.SrcBrowse = con.DstBrowse
			} else {
				con.SrcBrowse = tmpMap[v.DstBrowse].SrcBrowse
			}
			con.SrcNamespace = tmpMap[v.DstBrowse].SrcNamespace
			con.SrcNodeid = tmpMap[v.DstBrowse].SrcNodeid
			con.SrcUnit = tmpMap[v.DstBrowse].SrcUnit
			con.ScrDevName = devName
			con.Modify = tmpMap[v.DstBrowse].Modify
			con.Value = tmpMap[v.DstBrowse].Value
			convArr = append(convArr, con)
		}
	}
	return
}

// GetAllTemplate 取得所有範本資料
func GetAllTemplate() (tmp []nodeattr.Template, err error) {
	o := orm.NewOrm()
	defer func() {
		if r := recover(); r != nil {
			switch x := r.(type) {
			case string:
				err = errors.New(x)
			case error:
				err = x
			default:
				err = errors.New("Unknown panic")
			}
		}
	}()
	tmp, err = opcdao.GetAllTemplate(o)
	return
}

// GetAllConInfoNoPwd 取得不含密碼的連線資料
func GetAllConInfoNoPwd() (conInfos []nodeattr.ConInfo, err error) {
	o := orm.NewOrm()
	o.Begin()
	defer func() {
		if r := recover(); r != nil {
			o.Rollback()
			switch x := r.(type) {
			case string:
				err = errors.New(x)
			case error:
				err = x
			default:
				err = errors.New("Unknown panic")
			}
		} else {
			o.Commit()
		}
	}()
	conInfos, err = settingdao.GetAllConInfoNoPwd(o)
	return
}

// GetAllConInfo 取得連線資料
func GetAllConInfo() (conInfos []nodeattr.ConInfo, err error) {
	o := orm.NewOrm()
	o.Begin()
	defer func() {
		if r := recover(); r != nil {
			o.Rollback()
			switch x := r.(type) {
			case string:
				err = errors.New(x)
			case error:
				err = x
			default:
				err = errors.New("Unknown panic")
			}
		} else {
			o.Commit()
		}
	}()
	conInfos, err = settingdao.GetAllConInfo(o)
	return
}

// GetConnectRecord 取得所有連線資訊
func GetConnectRecord() (records []nodeattr.ConnectRecord, err error) {
	o := orm.NewOrm()
	o.Begin()
	defer func() {
		if r := recover(); r != nil {
			o.Rollback()
			switch x := r.(type) {
			case string:
				err = errors.New(x)
			case error:
				err = x
			default:
				err = errors.New("Unknown panic")
			}
		} else {
			o.Commit()
		}
	}()
	records, err = settingdao.GetConnectRecord(o)
	return
}

// GetAllDevInfo 取得所有設備資訊
func GetAllDevInfo() (devInfos []nodeattr.DevInfo, err error) {
	o := orm.NewOrm()
	o.Begin()
	defer func() {
		if r := recover(); r != nil {
			o.Rollback()
			switch x := r.(type) {
			case string:
				err = errors.New(x)
			case error:
				err = x
			default:
				err = errors.New("Unknown panic")
			}
		} else {
			o.Commit()
		}
	}()
	devInfos, err = settingdao.GetAllDevInfo(o)
	return
}

// GetDevInfoByConID 取得設備資訊By conID
func GetDevInfoByConID(conID int64) (devInfos []nodeattr.DevInfo, err error) {
	o := orm.NewOrm()
	o.Begin()
	defer func() {
		if r := recover(); r != nil {
			o.Rollback()
			switch x := r.(type) {
			case string:
				err = errors.New(x)
			case error:
				err = x
			default:
				err = errors.New("Unknown panic")
			}
		} else {
			o.Commit()
		}
	}()
	devInfos, err = settingdao.GetDevInfoByConID(conID, o)
	return
}

// GetDevInfoByProtocol 取得設備資訊By protocol
func GetDevInfoByProtocol(protocol string) (devInfos []nodeattr.DevInfo, err error) {
	o := orm.NewOrm()
	o.Begin()
	defer func() {
		if r := recover(); r != nil {
			o.Rollback()
			switch x := r.(type) {
			case string:
				err = errors.New(x)
			case error:
				err = x
			default:
				err = errors.New("Unknown panic")
			}
		} else {
			o.Commit()
		}
	}()
	devInfos, err = settingdao.GetDevInfoByProtocol(protocol, o)
	return
}

// GetDevInfoByName 取得設備資訊By name
func GetDevInfoByName(name string) (devInfo []nodeattr.DevInfo, err error) {
	o := orm.NewOrm()
	o.Begin()
	defer func() {
		if r := recover(); r != nil {
			o.Rollback()
			switch x := r.(type) {
			case string:
				err = errors.New(x)
			case error:
				err = x
			default:
				err = errors.New("Unknown panic")
			}
		} else {
			o.Commit()
		}
	}()
	devInfo, err = settingdao.GetDevInfoByName(name, o)
	return
}

// GetDevInfoByDevID 取得設備資訊
func GetDevInfoByDevID(devID int64) (devInfo nodeattr.DevInfo, err error) {
	o := orm.NewOrm()
	o.Begin()
	defer func() {
		if r := recover(); r != nil {
			o.Rollback()
			switch x := r.(type) {
			case string:
				err = errors.New(x)
			case error:
				err = x
			default:
				err = errors.New("Unknown panic")
			}
		} else {
			o.Commit()
		}
	}()
	devInfo, err = settingdao.GetDevInfoByDevID(devID, o)
	return
}

// InsertConInfo 新增連線資訊
func InsertConInfo(con nodeattr.ConInfo) (id int64, err error) {
	o := orm.NewOrm()
	o.Begin()
	defer func() {
		if r := recover(); r != nil {
			o.Rollback()
			switch x := r.(type) {
			case string:
				err = errors.New(x)
			case error:
				err = x
			default:
				err = errors.New("Unknown panic")
			}
		} else {
			o.Commit()
		}
	}()
	id, err = settingdao.InsertConInfo(con, o)
	return
}

// InsertMultiConInfo 新增多筆連線
func InsertMultiConInfo(con []nodeattr.ConInfo) (err error) {
	o := orm.NewOrm()
	o.Begin()
	defer func() {
		if r := recover(); r != nil {
			o.Rollback()
			switch x := r.(type) {
			case string:
				err = errors.New(x)
			case error:
				err = x
			default:
				err = errors.New("Unknown panic")
			}
		} else {
			o.Commit()
		}
	}()
	// for i := 0; i < len(con); i++ {
	// 	shaTmp512 := sha512.New()
	// 	shaTmp512.Write([]byte(con[i].Password))
	// 	sha512pwd := fmt.Sprintf("%x", shaTmp512.Sum(nil))
	// 	con[i].Password = sha512pwd
	// }
	err = settingdao.InsertMultiConInfo(con, o)
	return
}

// InsertDevInfo insert 設備資訊
func InsertDevInfo(devInfo nodeattr.DevInfo) (id int64, err error) {
	o := orm.NewOrm()
	o.Begin()
	defer func() {
		if r := recover(); r != nil {
			o.Rollback()
			switch x := r.(type) {
			case string:
				err = errors.New(x)
			case error:
				err = x
			default:
				err = errors.New("Unknown panic")
			}
		} else {
			o.Commit()
		}
	}()
	id, err = opcdao.InsertDevInfo(devInfo, o)
	return
}

// UpdateDevAuth 更新連線列表
func UpdateDevAuth(temp nodeattr.EditConInfo) (err error) {
	o := orm.NewOrm()
	o.Begin()
	defer func() {
		if r := recover(); r != nil {
			o.Rollback()
			switch x := r.(type) {
			case string:
				err = errors.New(x)
			case error:
				err = x
			default:
				err = errors.New("Unknown panic")
			}
		} else {
			o.Commit()
		}
	}()
	if err = settingdao.UpdateDevAuth(temp, o); err != nil {
		return
	}
	var dev nodeattr.DevInfo
	dev.ConID = temp.ID
	dev.ConName = temp.EditContent.Name
	err = settingdao.UpdateDevCon(dev, o)
	return
}

// UpdateDevTemplate 更新設備套用範本及名稱
func UpdateDevTemplate(editInfo nodeattr.EditDeviceInfo) (err error) {
	o := orm.NewOrm()
	o.Begin()
	defer func() {
		if r := recover(); r != nil {
			o.Rollback()
			switch x := r.(type) {
			case string:
				err = errors.New(x)
			case error:
				err = x
			default:
				err = errors.New("Unknown panic")
			}
		} else {
			o.Commit()
		}
	}()
	if err := settingdao.UpdateDevTemplate(editInfo, o); err != nil {
		panic(err)
	} else {
		if rels, err := opcdao.GetTempRel(editInfo.EditContent.TempID, o); err != nil {
			panic(err)
		} else {
			var convs []nodeattr.Converter
			for _, v := range rels {
				var tmp nodeattr.Converter
				tmp.ConvFunc = v.ConvFunc
				tmp.DstBrowse = v.DstBrowse
				tmp.DstNodeid = editInfo.EditContent.Name + "." + v.DstNodeid
				tmp.RefBrowseName1 = v.RefBrowseName1
				tmp.RefBrowseName2 = v.RefBrowseName2
				tmp.RefBrowseName3 = v.RefBrowseName3
				tmp.SrcBrowse = v.SrcBrowse
				tmp.SrcNamespace = v.SrcNamespace
				tmp.SrcNodeid = v.SrcNodeid
				tmp.SrcUnit = v.SrcUnit
				tmp.ScrDevName = editInfo.EditContent.Name
				convs = append(convs, tmp)
			}
			if err := opcdao.DelAllDataConv(editInfo.EditContent.Name, o); err != nil {
				panic(err)
			} else {
				if err := opcdao.InsertAllDataConv(convs, o); err != nil {
					panic(err)
				}
			}
		}
	}
	return nil
}

// UpdateDevStatusByID 更新設備狀態、廠區
func UpdateDevStatusByID(dev nodeattr.DevInfo) (err error) {
	o := orm.NewOrm()
	o.Begin()
	defer func() {
		if r := recover(); r != nil {
			o.Rollback()
			switch x := r.(type) {
			case string:
				err = errors.New(x)
			case error:
				err = x
			default:
				err = errors.New("Unknown panic")
			}
		} else {
			o.Commit()
		}
	}()
	err = settingdao.UpdateDevStatusByID(dev, o)
	return err
}

// UpdateDevStatusByMac 依照MAC更新設備狀態、廠區
func UpdateDevStatusByMac(dev nodeattr.DevInfo) (err error) {
	o := orm.NewOrm()
	o.Begin()
	defer func() {
		if r := recover(); r != nil {
			o.Rollback()
			switch x := r.(type) {
			case string:
				err = errors.New(x)
			case error:
				err = x
			default:
				err = errors.New("Unknown panic")
			}
		} else {
			o.Commit()
		}
	}()
	err = settingdao.UpdateDevStatusByMac(dev, o)
	return err
}

// DelConInfo 刪除連線
func DelConInfo(id int) (err error) {
	o := orm.NewOrm()
	o.Begin()
	defer func() {
		if r := recover(); r != nil {
			o.Rollback()
			switch x := r.(type) {
			case string:
				err = errors.New(x)
			case error:
				err = x
			default:
				err = errors.New("Unknown panic")
			}
		} else {
			o.Commit()
		}
	}()
	err = settingdao.DelConInfo(id, o)
	return
}

// DelDevInfo 刪除設備資料
func DelDevInfo(id int) (err error) {
	o := orm.NewOrm()
	o.Begin()
	defer func() {
		if r := recover(); r != nil {
			o.Rollback()
			switch x := r.(type) {
			case string:
				err = errors.New(x)
			case error:
				err = x
			default:
				err = errors.New("Unknown panic")
			}
		} else {
			o.Commit()
		}
	}()
	err = opcdao.DelDevInfo(id, o)
	return
}

// InsertTemplateRel 新增範本關聯表
func InsertTemplateRel(tempRels []nodeattr.TemplateRel, tempID int64) (err error) {
	o := orm.NewOrm()
	o.Begin()
	defer func() {
		if r := recover(); r != nil {
			o.Rollback()
			switch x := r.(type) {
			case string:
				err = errors.New(x)
			case error:
				err = x
			default:
				err = errors.New("Unknown panic")
			}
		} else {
			o.Commit()
		}
	}()
	for i := 0; i < len(tempRels); i++ {
		tempRels[i].ID = 0
		tempRels[i].TempID = tempID
		if tempID == DefaultTemp.ID {
			tempRels[i].SrcBrowse = tempRels[i].DstBrowse
		}
	}
	err = opcdao.InsertTemplateRel(tempRels, o)
	return
}

// UpdateConnectInfo 更新連線記錄
func UpdateConnectInfo(con nodeattr.ConInfo) (err error) {
	o := orm.NewOrm()
	o.Begin()
	defer func() {
		if r := recover(); r != nil {
			o.Rollback()
			switch x := r.(type) {
			case string:
				err = errors.New(x)
			case error:
				err = x
			default:
				err = errors.New("Unknown panic")
			}
		} else {
			o.Commit()
		}
	}()
	err = settingdao.UpdateConnectInfo(con, o)
	return
}

// UpdateDevInfo 更新設備列表
func UpdateDevInfo(dev nodeattr.DevInfo) (err error) {
	o := orm.NewOrm()
	o.Begin()
	defer func() {
		if r := recover(); r != nil {
			o.Rollback()
			switch x := r.(type) {
			case string:
				err = errors.New(x)
			case error:
				err = x
			default:
				err = errors.New("Unknown panic")
			}
		} else {
			o.Commit()
		}
	}()
	err = settingdao.UpdateDevInfo(dev, o)
	return
}

// UpdateDevInfoByConID 更新設備狀態
func UpdateDevInfoByConID(dev nodeattr.DevInfo) (err error) {
	o := orm.NewOrm()
	o.Begin()
	defer func() {
		if r := recover(); r != nil {
			o.Rollback()
			switch x := r.(type) {
			case string:
				err = errors.New(x)
			case error:
				err = x
			default:
				err = errors.New("Unknown panic")
			}
		} else {
			o.Commit()
		}
	}()
	err = settingdao.UpdateDevInfoByConID(dev, o)
	return
}

// GetEdgeAuth 取得帳戶資料
func GetEdgeAuth() (edgeAuths []nodeattr.EdgeAuth, err error) {
	o := orm.NewOrm()
	o.Begin()
	defer func() {
		if r := recover(); r != nil {
			o.Rollback()
			switch x := r.(type) {
			case string:
				err = errors.New(x)
			case error:
				err = x
			default:
				err = errors.New("Unknown panic")
			}
		} else {
			o.Commit()
		}
	}()
	edgeAuths, err = settingdao.GetEdgeAuth(o)
	return edgeAuths, err
}

// DelEdgeAuth 刪除帳戶資料
func DelEdgeAuth(id int) (err error) {
	o := orm.NewOrm()
	o.Begin()
	defer func() {
		if r := recover(); r != nil {
			o.Rollback()
			switch x := r.(type) {
			case string:
				err = errors.New(x)
			case error:
				err = x
			default:
				err = errors.New("Unknown panic")
			}
		} else {
			o.Commit()
		}
	}()
	err = settingdao.DelEdgeAuth(id, o)
	return
}

// InsertEdgeAuth 新增帳戶資料
func InsertEdgeAuth(accounts []nodeattr.EdgeAuth) (err error) {
	var tmpArr []nodeattr.EdgeAuth
	o := orm.NewOrm()
	o.Begin()
	defer func() {
		if r := recover(); r != nil {
			o.Rollback()
			switch x := r.(type) {
			case string:
				err = errors.New(x)
			case error:
				err = x
			default:
				err = errors.New("Unknown panic")
			}
		} else {
			o.Commit()
		}
	}()
	for _, account := range accounts {
		var tmp nodeattr.EdgeAuth
		shaTmp512 := sha512.New()
		shaTmp512.Write([]byte(account.Password))
		tmp.Account = account.Account
		tmp.Password = fmt.Sprintf("%x", shaTmp512.Sum(nil))
		tmp.Name = account.Name
		tmp.Timestamp = account.Timestamp
		tmpArr = append(tmpArr, tmp)
	}
	err = settingdao.InsertEdgeAuth(tmpArr, o)
	return
}

// UpdateAccountPwd 更新帳戶資料
func UpdateAccountPwd(account nodeattr.EditEdgeAuth) (err error) {
	o := orm.NewOrm()
	o.Begin()
	defer func() {
		if r := recover(); r != nil {
			o.Rollback()
			switch x := r.(type) {
			case string:
				err = errors.New(x)
			case error:
				err = x
			default:
				err = errors.New("Unknown panic")
			}
		} else {
			o.Commit()
		}
	}()
	// shaTmp512 := sha512.New()
	// shaTmp512.Write([]byte(account.EditContent.Password))
	// account.EditContent.Password = string(shaTmp512.Sum(nil))
	err = settingdao.UpdateAccountPwd(account, o)
	return
}

// UpdateAccountLoginTime 更新最後登入時間
func UpdateAccountLoginTime(account nodeattr.EdgeAuth) (err error) {
	o := orm.NewOrm()
	o.Begin()
	defer func() {
		if r := recover(); r != nil {
			o.Rollback()
			switch x := r.(type) {
			case string:
				err = errors.New(x)
			case error:
				err = x
			default:
				err = errors.New("Unknown panic")
			}
		} else {
			o.Commit()
		}
	}()
	// shaTmp512 := sha512.New()
	// shaTmp512.Write([]byte(account.EditContent.Password))
	// account.EditContent.Password = string(shaTmp512.Sum(nil))
	err = settingdao.UpdateAccountLoginTime(account, o)
	return
}

// UpdatePermission 更新帳戶權限
func UpdatePermission(permission nodeattr.EditPermission) (err error) {
	o := orm.NewOrm()
	o.Begin()
	defer func() {
		if r := recover(); r != nil {
			o.Rollback()
			switch x := r.(type) {
			case string:
				err = errors.New(x)
			case error:
				err = x
			default:
				err = errors.New("Unknown panic")
			}
		} else {
			o.Commit()
		}
	}()
	err = settingdao.UpdatePermission(permission, o)
	return
}

// InsertPermission 更新帳戶權限
func InsertPermission(permissions []nodeattr.AuthPermissopn) (err error) {
	o := orm.NewOrm()
	o.Begin()
	defer func() {
		if r := recover(); r != nil {
			o.Rollback()
			switch x := r.(type) {
			case string:
				err = errors.New(x)
			case error:
				err = x
			default:
				err = errors.New("Unknown panic")
			}
		} else {
			o.Commit()
		}
	}()
	err = settingdao.InsertPermission(permissions, o)
	return
}

// GetPermission 取得帳戶權限
func GetPermission(id int64) (permissions []nodeattr.AuthPermissopn, err error) {
	o := orm.NewOrm()
	o.Begin()
	defer func() {
		if r := recover(); r != nil {
			o.Rollback()
			switch x := r.(type) {
			case string:
				err = errors.New(x)
			case error:
				err = x
			default:
				err = errors.New("Unknown panic")
			}
		} else {
			o.Commit()
		}
	}()
	permissions, err = settingdao.GetPermission(id, o)
	return
}

// GetEdgeAuthByName 使用名稱取得帳戶資料
func GetEdgeAuthByName(name string) (edgeAuth nodeattr.EdgeAuth, err error) {
	o := orm.NewOrm()
	o.Begin()
	defer func() {
		if r := recover(); r != nil {
			o.Rollback()
			switch x := r.(type) {
			case string:
				err = errors.New(x)
			case error:
				err = x
			default:
				err = errors.New("Unknown panic")
			}
		} else {
			o.Commit()
		}
	}()
	edgeAuth, err = settingdao.GetEdgeAuthByName(name, o)
	return
}

// InsertConRecordInfo 新增連線記錄
func InsertConRecordInfo(record nodeattr.ConnectRecord) (err error) {
	o := orm.NewOrm()
	o.Begin()
	defer func() {
		if r := recover(); r != nil {
			o.Rollback()
			switch x := r.(type) {
			case string:
				err = errors.New(x)
			case error:
				err = x
			default:
				err = errors.New("Unknown panic")
			}
		} else {
			o.Commit()
		}
	}()
	err = settingdao.InsertConRecordInfo(record, o)
	return
}

// UpdateConvter 更新轉換表
func UpdateConvter(conv nodeattr.EditConverter) (err error) {
	o := orm.NewOrm()
	o.Begin()
	defer func() {
		if r := recover(); r != nil {
			o.Rollback()
			switch x := r.(type) {
			case string:
				err = errors.New(x)
			case error:
				err = x
			default:
				err = errors.New("Unknown panic")
			}
		} else {
			o.Commit()
		}
	}()
	convArr := conv.EditContent
	for _, v := range conv.EditContent { // 檢查公式
		if checkErr := checkFormula(v.ConvFunc); checkErr != nil {
			panic("Not a formula")
		}
	}
	var dev nodeattr.EditDeviceInfo
	for _, v := range global.Devs {
		if v.Name == conv.Name {
			dev.ID = v.ID
			dev.EditContent.TempID = v.TempID
			dev.EditContent.TempName = v.TempName
			break
		}
	}
	browseMap := make(map[string]nodeattr.TemplateRel)
	if tmpRel, err := GetTempRel(dev.EditContent.TempID); err != nil {
		return err
	} else {
		var tmp nodeattr.EditConverter
		tmp.Name = dev.EditContent.Name
		for _, v := range tmpRel {
			browseMap[v.DstBrowse] = v
		}
		for _, v := range convArr {
			v.ID = 0
			if browseMap[v.DstBrowse].SrcBrowse != v.SrcBrowse || browseMap[v.DstBrowse].SrcNamespace != v.SrcNamespace || browseMap[v.DstBrowse].SrcNodeid != v.SrcNodeid ||
				browseMap[v.DstBrowse].ConvFunc != v.ConvFunc || browseMap[v.DstBrowse].RefBrowseName1 != v.RefBrowseName1 || browseMap[v.DstBrowse].RefBrowseName2 != v.RefBrowseName2 ||
				browseMap[v.DstBrowse].RefBrowseName3 != v.RefBrowseName3 {
				// 有編輯過
				v.Modify = 1
			} else {
				v.Modify = 0
			}
			tmp.EditContent = append(tmp.EditContent, v)
		}

		if err := opcdao.DelDataConv(tmp, o); err != nil {
			panic(err)
		}
		if err := opcdao.InsertDataConv(tmp, o); err != nil {
			panic(err)
		}
	}
	return
}
func checkFormula(str string) error {
	if len(str) == 0 {
		return nil
	}
	expression, err := govaluate.NewEvaluableExpression(str)
	if err != nil {
		return err
	}
	parameters := make(map[string]interface{})
	parameters["x"] = 8
	parameters["y"] = 8
	parameters["z"] = 8
	_, err = expression.Evaluate(parameters)
	return err
}

// UpdateTempRel 更新關聯表
func UpdateTempRel(editInfo nodeattr.EditTemplateRel, templateArr []nodeattr.Template) (err error) {
	o := orm.NewOrm()
	o.Begin()
	defer func() {
		if r := recover(); r != nil {
			o.Rollback()
			switch x := r.(type) {
			case string:
				err = errors.New(x)
			case error:
				err = x
			default:
				err = errors.New("Unknown panic")
			}
		} else {
			o.Commit()
		}
	}()
	for _, v := range editInfo.EditContent { // 檢查公式
		if checkErr := checkFormula(v.ConvFunc); checkErr != nil {
			panic("Not a formula")
		}
	}

	if err := opcdao.DelTempRel(editInfo, o); err != nil {
		panic(err)
	}

	if err := opcdao.InsertTemplateRel(editInfo.EditContent, o); err != nil {
		panic(err)
	}
	return
}
func InsertDataConv(dev nodeattr.DevInfo) (editinfo nodeattr.EditConverter, err error) {
	o := orm.NewOrm()
	o.Begin()
	defer func() {
		if r := recover(); r != nil {
			o.Rollback()
			switch x := r.(type) {
			case string:
				err = errors.New(x)
			case error:
				err = x
			default:
				err = errors.New("Unknown panic")
			}
		} else {
			o.Commit()
		}
	}()
	if tmp, err := GetTemplateRel(dev.TempID); err != nil {
		panic(err)
	} else {
		var conv nodeattr.Converter
		editinfo.Name = dev.Name
		for _, v := range tmp {
			conv.ConvFunc = v.ConvFunc
			conv.ScrDevName = dev.Name
			conv.SrcBrowse = v.SrcBrowse
			conv.SrcNamespace = v.SrcNamespace
			conv.SrcNodeid = v.SrcNodeid
			conv.SrcUnit = v.SrcUnit
			conv.RefBrowseName1 = v.RefBrowseName1
			conv.RefBrowseName2 = v.RefBrowseName2
			conv.RefBrowseName3 = v.RefBrowseName3
			conv.DstBrowse = v.DstBrowse
			conv.DstNodeid = v.DstNodeid
			editinfo.EditContent = append(editinfo.EditContent, conv)
		}
		if err := opcdao.InsertDataConv(editinfo, o); err != nil {
			panic(err)
		}
	}
	return
}

func GetTempRel(id int64) ([]nodeattr.TemplateRel, error) {
	if tmpRels, err := GetTemplateRel(id); err != nil {
		return nil, err
	} else {
		if len(tmpRels) == 0 {
			// 此 temp_id 的範本關聯表未建立，用預設節點建立新的關聯表，存入DB
			InsertTemplateRel(DefaultRels, id)
			return DefaultRels, nil
		} else {
			return tmpRels, nil
		}
	}
}

// InsertTemp 新增複製範本
func InsertTemp(tmp nodeattr.Template, templateArr []nodeattr.Template) error {
	var newTmpID int64
	var oldTmpID int64
	for _, v := range templateArr {
		if tmp.Name == v.Name { // 檢查範本名稱是否重複
			return errors.New("template name is exist")
		}
	}
	if len(tmp.Model) != 0 { // 依照模型另存範本
		if convArr, err := GetDataConv(tmp.Model, 0); err != nil {
			return err
		} else {
			oldTmpID = tmp.ID
			tmp.ID = 0
			tmp.CreateTime = time.Now().UnixNano() / int64(time.Millisecond)
			tmp.ModifyTime = tmp.CreateTime
			if id, insertErr := InsertTemplate(tmp); insertErr != nil {
				return insertErr
			} else {
				newTmpID = id
			}
			var tmpRels []nodeattr.TemplateRel
			for _, v := range convArr {
				var tmpRel nodeattr.TemplateRel
				if v.Modify == 1 {
					tmpRel.SrcBrowse = v.SrcBrowse
					tmpRel.ConvFunc = v.ConvFunc
					tmpRel.SrcNamespace = v.SrcNamespace
					tmpRel.SrcNodeid = v.SrcNodeid
					tmpRel.SrcUnit = v.SrcUnit
					tmpRel.RefBrowseName1 = v.RefBrowseName1
					tmpRel.RefBrowseName2 = v.RefBrowseName2
					tmpRel.RefBrowseName3 = v.RefBrowseName3
				}
				tmpRel.DstBrowse = v.DstBrowse
				tmpRel.DstNodeid = v.DstNodeid
				tmpRels = append(tmpRels, tmpRel)
			}
			err := InsertTemplateRel(tmpRels, newTmpID)
			return err
		}
	}
	for _, v := range templateArr {
		if tmp.ID == v.ID { // 複製現有範本為新範本
			oldTmpID = tmp.ID
			tmp.ID = 0
			tmp.CreateTime = time.Now().UnixNano() / int64(time.Millisecond)
			tmp.ModifyTime = tmp.CreateTime
			if id, insertErr := InsertTemplate(tmp); insertErr != nil {
				return insertErr
			} else {
				newTmpID = id
			}
			if tmpRels, err := GetTempRel(oldTmpID); err != nil {
				return err
			} else {
				err := InsertTemplateRel(tmpRels, newTmpID)
				return err
			}
		}
	}
	tmp.CreateTime = time.Now().UnixNano() / int64(time.Millisecond)
	tmp.ModifyTime = tmp.CreateTime
	_, insertErr := InsertTemplate(tmp)
	return insertErr
}

func CheckConvert(editConvert nodeattr.EditConverter) (string, error) {
	var msg string
	if len(editConvert.EditContent) == 0 {
		return msg, nil
	}
	if convs, err := GetDataConv(editConvert.Name, 0); err != nil {
		return msg, err
	} else {
		m := make(map[string]nodeattr.Converter)
		for _, v := range editConvert.EditContent {
			m[v.DstBrowse] = v
		}
		for _, v := range convs {
			if v.SrcBrowse != m[v.DstBrowse].SrcBrowse || v.SrcNodeid != m[v.DstBrowse].SrcNodeid || v.ConvFunc != m[v.DstBrowse].ConvFunc ||
				v.RefBrowseName1 != m[v.DstBrowse].RefBrowseName1 || v.RefBrowseName2 != m[v.DstBrowse].RefBrowseName2 || v.RefBrowseName3 != m[v.DstBrowse].RefBrowseName3 ||
				v.SrcNamespace != m[v.DstBrowse].SrcNamespace {
				msg = "Converter has be modified"
				break
			}
		}
	}
	return msg, nil
}
func CopyConverter(editInfo nodeattr.EditConverter) (err error) {
	o := orm.NewOrm()
	o.Begin()
	defer func() {
		if r := recover(); r != nil {
			o.Rollback()
			switch x := r.(type) {
			case string:
				err = errors.New(x)
			case error:
				err = x
			default:
				err = errors.New("Unknown panic")
			}
		} else {
			o.Commit()
		}
	}()
	var tempRels []nodeattr.TemplateRel
	var template nodeattr.Template
	var srcDevName string
	template.CreateTime = time.Now().UnixNano() / int64(time.Millisecond)
	template.ModifyTime = template.CreateTime
	template.Name = editInfo.Name
	var tempID int64
	if id, err := InsertTemplate(template); err != nil {
		return err
	} else {
		tempID = id
	}
	for _, v := range editInfo.EditContent {
		var tmp nodeattr.TemplateRel
		srcDevName = v.ScrDevName
		tmp.ConvFunc = v.ConvFunc
		tmp.SrcBrowse = v.SrcBrowse
		tmp.SrcNodeid = v.SrcNodeid
		tmp.SrcNamespace = v.SrcNamespace
		tmp.DstBrowse = v.DstBrowse
		tmp.DstNodeid = v.DstNodeid
		tmp.RefBrowseName1 = v.RefBrowseName1
		tmp.RefBrowseName2 = v.RefBrowseName2
		tmp.RefBrowseName3 = v.RefBrowseName3
		tempRels = append(tempRels, tmp)
	}
	if err := InsertTemplateRel(tempRels, tempID); err != nil {
		return err
	}
	if devs, err := GetAllDevInfo(); err != nil {
		return err
	} else {
		for _, v := range devs {
			if v.Name == srcDevName {
				var dev nodeattr.EditDeviceInfo
				dev.ID = v.ID
				dev.EditContent.TempID = tempID
				dev.EditContent.TempName = editInfo.Name
				if err := settingdao.UpdateDevTemplate(dev, o); err != nil {
					return err
				}
				break
			}
		}
	}
	return
}
func DelTempRelByTempID(id int64) (err error) {
	o := orm.NewOrm()
	o.Begin()
	defer func() {
		if r := recover(); r != nil {
			o.Rollback()
			switch x := r.(type) {
			case string:
				err = errors.New(x)
			case error:
				err = x
			default:
				err = errors.New("Unknown panic")
			}
		} else {
			o.Commit()
		}
	}()
	err = opcdao.DelTempRelByTempID(id, o)
	return
}

// ImportTemp 匯入範本
func ImportTemp(tmp nodeattr.Template, templateArr []nodeattr.Template, editTmp nodeattr.EditTemplateRel) error {
	for _, v := range templateArr {
		if tmp.Name == v.Name { // 檢查範本名稱是否重複
			return errors.New("template name is exist")
		}
	}
	tmp.ID = 0
	var newTmpID int64
	tmp.CreateTime = time.Now().UnixNano() / int64(time.Millisecond)
	tmp.ModifyTime = tmp.CreateTime
	if id, insertErr := InsertTemplate(tmp); insertErr != nil {
		return insertErr
	} else {
		newTmpID = id
	}
	err := InsertTemplateRel(editTmp.EditContent, newTmpID)
	return err
}
