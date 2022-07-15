package settingdao

import (
	"FSRV_Edge/nodeattr"
	"strconv"
	"sync"

	"github.com/astaxie/beego/orm"
)

const (
	defaultMsg       = "無法取得連線狀態"
	tmpRelMaxNum     = 10
	disConMsg        = 2
	nodeStructMaxNum = 100
)

// GetAllConInfoNoPwd 取得不含密碼的連線資料
func GetAllConInfoNoPwd(o orm.Ormer) (conInfos []nodeattr.ConInfo, err error) {
	rs := o.Raw("SELECT id, ip, port, name, protocol, status, account, timestamp, exist FROM device_auth")
	_, err = rs.QueryRows(&conInfos)
	return conInfos, err
}

// GetAllConInfo 取得連線資料
func GetAllConInfo(o orm.Ormer) (conInfos []nodeattr.ConInfo, err error) {
	rs := o.Raw("SELECT * FROM device_auth")
	_, err = rs.QueryRows(&conInfos)
	return conInfos, err
}

var MacData macData

type macData struct {
	DataMap map[string]string
	Lock    sync.Mutex
}

func init() {
	MacData.Lock.Lock()
	MacData.DataMap = make(map[string]string)
	MacData.Lock.Unlock()
}

// GetAllDevInfo 取得全部設備資訊
func GetAllDevInfo(o orm.Ormer) (devInfos []nodeattr.DevInfo, err error) {
	rs := o.Raw("SELECT * FROM device_info")
	var tmpArr []nodeattr.DevInfo
	_, err = rs.QueryRows(&tmpArr)
	for _, v := range tmpArr {
		if v.Mac != "" {
			MacData.Lock.Lock()
			idStr := strconv.FormatInt(v.ID, 10)
			MacData.DataMap[v.Mac] = idStr
			MacData.DataMap[idStr] = v.Mac
			MacData.Lock.Unlock()
		}
		v.Name = "Device" + strconv.FormatInt(v.ID, 10)
		devInfos = append(devInfos, v)
	}
	return devInfos, err
}

// GetDevInfoByConID 取得設備資訊By conID
func GetDevInfoByConID(conID int64, o orm.Ormer) (devInfos []nodeattr.DevInfo, err error) {
	rs := o.Raw("SELECT * FROM device_info WHERE connect_id=?", conID)
	_, err = rs.QueryRows(&devInfos)
	return devInfos, err
}

// GetDevInfoByProtocol 取得設備資訊By Protocol
func GetDevInfoByProtocol(protocol string, o orm.Ormer) (devInfos []nodeattr.DevInfo, err error) {
	rs := o.Raw("SELECT * FROM device_info WHERE protocol=?", protocol)
	_, err = rs.QueryRows(&devInfos)
	return devInfos, err
}

// GetDevInfoByName 取得設備資訊By name
func GetDevInfoByName(name string, o orm.Ormer) (devInfos []nodeattr.DevInfo, err error) {
	rs := o.Raw("SELECT * FROM device_info WHERE name=?", name)
	_, err = rs.QueryRows(&devInfos)
	return devInfos, err
}

// GetDevInfoByDevID 取得設備資訊By devID
func GetDevInfoByDevID(devID int64, o orm.Ormer) (devInfo nodeattr.DevInfo, err error) {
	rs := o.Raw("SELECT * FROM device_info WHERE id=?", devID)
	err = rs.QueryRow(&devInfo)
	return devInfo, err
}

// GetEdgeAuth 取得帳戶資料
func GetEdgeAuth(o orm.Ormer) (edgeAuths []nodeattr.EdgeAuth, err error) {
	rs := o.Raw("SELECT id,name,account,timestamp FROM edge_auth")
	_, err = rs.QueryRows(&edgeAuths)
	return edgeAuths, err
}

// GetConnectRecord 取得連線記錄
func GetConnectRecord(o orm.Ormer) (records []nodeattr.ConnectRecord, err error) {
	rs := o.Raw("SELECT * FROM connect_record")
	_, err = rs.QueryRows(&records)
	return records, err
}

// GetEdgeAuthByName 使用名稱取得帳戶資料
func GetEdgeAuthByName(name string, o orm.Ormer) (edgeAuth nodeattr.EdgeAuth, err error) {
	rs := o.Raw("SELECT * FROM edge_auth WHERE account=?", name)
	err = rs.QueryRow(&edgeAuth)
	return edgeAuth, err
}

// GetSetting 取得系統設定資料
func GetSetting(o orm.Ormer) (setting nodeattr.SystemSetting, err error) {
	rs := o.Raw("SELECT * FROM system_setting")
	err = rs.QueryRow(&setting)
	return setting, err
}

// GetPermission 取得帳戶資料
func GetPermission(id int64, o orm.Ormer) (permissions []nodeattr.AuthPermissopn, err error) {
	rs := o.Raw("SELECT * FROM auth_permission WHERE account_id=?", id)
	_, err = rs.QueryRows(&permissions)
	return permissions, err
}

// UpdateDevStatusByMac 依照DC MAC更新設備狀態、廠區
func UpdateDevStatusByMac(dev nodeattr.DevInfo, o orm.Ormer) error {
	_, err := o.Raw("UPDATE device_info SET brand=?, status=?, connect_name=? WHERE mac=?", dev.Brand, dev.Status, dev.ConName, dev.Mac).Exec()
	return err
}

// UpdateDevAuth 更新DB連線列表資訊
func UpdateDevAuth(temp nodeattr.EditConInfo, o orm.Ormer) error {
	_, err := o.Raw("UPDATE device_auth SET ip=?,port=?,name=?,protocol=?,account=?,password=? WHERE id=?", temp.EditContent.IP, temp.EditContent.Port, temp.EditContent.Name, temp.EditContent.Protocol, temp.EditContent.Account, temp.EditContent.Password, temp.ID).Exec()
	return err
}

// UpdateConnectInfo 更新連線狀態、時間
func UpdateConnectInfo(con nodeattr.ConInfo, o orm.Ormer) error {
	_, err := o.Raw("UPDATE device_auth SET ip=?,timestamp=?, status=?, exist=? WHERE name=?", con.IP, con.Timestamp, con.Status, con.Exist, con.Name).Exec()
	return err
}

// UpdateDevInfo 更新設備列表
func UpdateDevInfo(dev nodeattr.DevInfo, o orm.Ormer) error {
	_, err := o.Raw("UPDATE device_info SET connect_id=?, connect_name=?, brand=?, protocol=?, status=?, basic_auth=? WHERE id=?",
		dev.ConID, dev.ConName, dev.Brand, dev.Protocol, dev.Status, dev.Auth, dev.ID).Exec()
	return err
}

// UpdateDevInfoByConID 更新設備狀態by conID
func UpdateDevInfoByConID(dev nodeattr.DevInfo, o orm.Ormer) error {
	_, err := o.Raw("UPDATE device_info SET status=? WHERE connect_id=?",
		dev.Status, dev.ConID).Exec()
	return err
}

// UpdateDevCon 更新設備連線名稱
func UpdateDevCon(dev nodeattr.DevInfo, o orm.Ormer) error {
	_, err := o.Raw("UPDATE device_info SET connect_name=? WHERE connect_id=?",
		dev.ConName, dev.ConID).Exec()
	return err
}

// UpdateDevTemplate 更新設備套用範本ID、名稱
func UpdateDevTemplate(dev nodeattr.EditDeviceInfo, o orm.Ormer) error {
	_, err := o.Raw("UPDATE device_info SET template_id=?, template_name=? WHERE id=?",
		dev.EditContent.TempID, dev.EditContent.TempName, dev.ID).Exec()
	return err
}

// UpdateDevStatusByID 更新設備狀態
func UpdateDevStatusByID(dev nodeattr.DevInfo, o orm.Ormer) error {
	_, err := o.Raw("UPDATE device_info SET brand=?, status=? WHERE id=?", dev.Brand, dev.Status, dev.ID).Exec()
	return err
}

// UpdateAccountPwd 更新帳戶密碼
func UpdateAccountPwd(account nodeattr.EditEdgeAuth, o orm.Ormer) error {
	_, err := o.Raw("UPDATE edge_auth SET password=? WHERE id=?", account.EditContent.Password, account.ID).Exec()
	return err
}

// UpdateAccountLoginTime 更新帳戶登入時間
func UpdateAccountLoginTime(account nodeattr.EdgeAuth, o orm.Ormer) error {
	_, err := o.Raw("UPDATE edge_auth SET timestamp=? WHERE account=?", account.Timestamp, account.Account).Exec()
	return err
}

// UpdateSetting 更新系統設定
func UpdateSetting(setting nodeattr.SystemSetting, o orm.Ormer) error {
	_, err := o.Raw("UPDATE system_setting SET ip=?, port=? WHERE id=?", setting.IP, setting.Port, setting.ID).Exec()
	return err
}

// UpdatePermission 更新帳戶權限
func UpdatePermission(permission nodeattr.EditPermission, o orm.Ormer) error {
	for _, content := range permission.EditContent {
		if _, err := o.Raw("UPDATE auth_permission SET action=? WHERE account_id=? AND browse_name=?", content.Action, permission.ID, content.BrowseName).Exec(); err != nil {
			return err
		}
	}
	return nil
}

// InsertConInfo 新增連線
func InsertConInfo(con nodeattr.ConInfo, o orm.Ormer) (int64, error) {
	id, err := o.Insert(&con)
	return id, err
}

// InsertMultiConInfo 新增多筆連線
func InsertMultiConInfo(con []nodeattr.ConInfo, o orm.Ormer) error {
	if len(con) == 0 {
		return nil
	}
	_, err := o.InsertMulti(nodeStructMaxNum, con)
	return err
}

// InsertConRecordInfo 新增連線紀錄
func InsertConRecordInfo(con nodeattr.ConnectRecord, o orm.Ormer) error {
	_, err := o.Insert(&con)
	return err
}

// InsertSetting 新增系統設定
func InsertSetting(setting nodeattr.SystemSetting, o orm.Ormer) error {
	_, err := o.Insert(&setting)
	return err
}

// InsertEdgeAuth 新增帳戶
func InsertEdgeAuth(accounts []nodeattr.EdgeAuth, o orm.Ormer) error {
	if len(accounts) == 0 {
		return nil
	}
	_, err := o.InsertMulti(tmpRelMaxNum, accounts)
	return err
}

// InsertPermission 新增帳戶權限
func InsertPermission(accounts []nodeattr.AuthPermissopn, o orm.Ormer) error {
	if len(accounts) == 0 {
		return nil
	}
	_, err := o.InsertMulti(nodeStructMaxNum, accounts)
	return err
}

// DelConInfo 刪除連線列表
func DelConInfo(id int, o orm.Ormer) error {
	_, err := o.Raw("DELETE FROM device_auth WHERE id=?", id).Exec()
	return err
}

// DelEdgeAuth 刪除帳戶
func DelEdgeAuth(id int, o orm.Ormer) error {
	_, err := o.Raw("DELETE FROM edge_auth WHERE id=?", id).Exec()
	return err
}
