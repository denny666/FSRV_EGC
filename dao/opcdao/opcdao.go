package opcdao

import (
	"FSRV_Edge/nodeattr"
	"context"
	"reflect"
	"strings"

	"github.com/astaxie/beego/orm"

	// need import
	_ "github.com/mattn/go-sqlite3"
)

var (
	flag   = false
	ctx    context.Context
	cancel context.CancelFunc
)

const (
	tmpRelMaxNum     = 10
	nodeStructMaxNum = 100
)

// MySQLArrayToQuestionStr input:[1,2,4,5] output:(?,?,?,?)
func MySQLArrayToQuestionStr(arr interface{}) string {
	var str string
	s := reflect.ValueOf(arr)
	if s.Kind() != reflect.Slice {
		return str
	}

	str += "("

	for i := 0; i < s.Len(); i++ {
		str += "?,"
	}

	str = strings.TrimRight(str, ",")
	str += ")"
	return str
}

// GetTableSize 取得 table size
func GetTableSize(tableName string, o orm.Ormer) (int64, error) {
	var count int64
	rs := o.Raw("SELECT count(*) FROM " + tableName)
	err := rs.QueryRow(&count)
	return count, err
}

// GetAllTemplate 取得範本列表
func GetAllTemplate(o orm.Ormer) (tmpArr []nodeattr.Template, err error) {
	rs := o.Raw("SELECT * FROM node_template")
	_, err = rs.QueryRows(&tmpArr)
	return tmpArr, err
}

// GetTemplateByName 取得範本列表By Name
func GetTemplateByName(name string, o orm.Ormer) (tmp nodeattr.Template, err error) {
	rs := o.Raw("SELECT * FROM node_template WHERE name=?", name)
	err = rs.QueryRow(&tmp)
	return tmp, err
}

// GetTempRel 取得關聯表資料
func GetTempRel(tempID int64, o orm.Ormer) (tmpRelArr []nodeattr.TemplateRel, err error) {
	rs := o.Raw("SELECT * FROM node_template_rel WHERE template_id=?", tempID)
	_, err = rs.QueryRows(&tmpRelArr)
	return tmpRelArr, err
}

// GetDataConv 取得轉換表DB資料
func GetDataConv(devName string, o orm.Ormer) (convArr []nodeattr.Converter, err error) {
	rs := o.Raw("SELECT * FROM data_converter WHERE src_dev_name=?", devName)
	_, err = rs.QueryRows(&convArr)
	return convArr, err
}

// GetDataConvDev 取得轉換表設備名稱
func GetDataConvDev(o orm.Ormer) (convArr []string, err error) {
	rs := o.Raw(`SELECT DISTINCT src_dev_name FROM data_converter`)
	_, err = rs.QueryRows(&convArr)
	return convArr, err
}

// InsertDevInfo 新增設備
func InsertDevInfo(devinfo nodeattr.DevInfo, o orm.Ormer) (int64, error) {
	id, err := o.Insert(&devinfo)
	return id, err
}

// InsertTemplateRel 儲存範本關聯資料至DB
func InsertTemplateRel(tempRel []nodeattr.TemplateRel, o orm.Ormer) error {
	if len(tempRel) == 0 {
		return nil
	}
	_, err := o.InsertMulti(tmpRelMaxNum, tempRel)
	return err
}

// InsertAllDataConv 儲存轉換表資料至DB
func InsertAllDataConv(conv []nodeattr.Converter, o orm.Ormer) error {
	if len(conv) == 0 {
		return nil
	}
	_, err := o.InsertMulti(tmpRelMaxNum, conv)
	return err
}

// InsertDataConv 儲存範本列表至DB
func InsertDataConv(conv nodeattr.EditConverter, o orm.Ormer) error {
	if len(conv.EditContent) == 0 {
		return nil
	}
	_, err := o.InsertMulti(tmpRelMaxNum, conv.EditContent)
	return err
}

// InsertTemplate 儲存範本列表至DB
func InsertTemplate(temp nodeattr.Template, o orm.Ormer) (int64, error) {
	id, err := o.Insert(&temp)
	return id, err
}

// InsertNodeStruct 儲存預設節點資料至DB
func InsertNodeStruct(nodeStruct []nodeattr.NodeStruct, o orm.Ormer) error {
	if len(nodeStruct) == 0 {
		return nil
	}
	_, err := o.InsertMulti(nodeStructMaxNum, nodeStruct)
	return err
}

// DelTemplate 刪除範本列表
func DelTemplate(id int, o orm.Ormer) error {
	_, err := o.Raw("DELETE FROM node_template WHERE id=?", id).Exec()
	return err
}

// DelTempRelByTempID 刪除指定TempID範本關聯
func DelTempRelByTempID(id int64, o orm.Ormer) error {
	_, err := o.Raw("DELETE FROM node_template_rel WHERE template_id=?", id).Exec()
	return err
}

// DelTempRel 刪除範本關聯表
func DelTempRel(tempRel nodeattr.EditTemplateRel, o orm.Ormer) error {
	if len(tempRel.EditContent) == 0 {
		return nil
	}
	var arr []string
	for _, v := range tempRel.EditContent {
		arr = append(arr, v.DstBrowse)
	}
	sql := `DELETE FROM node_template_rel WHERE template_id=? and dst_browse_name IN ` + MySQLArrayToQuestionStr(arr)
	_, err := o.Raw(sql, tempRel.TempID, arr).Exec()
	return err
}

// DelAllDataConv 刪除設備轉換表
func DelAllDataConv(name string, o orm.Ormer) error {
	_, err := o.Raw("DELETE FROM data_converter WHERE src_dev_name=?", name).Exec()
	return err
}

// DelDataConv 刪除多筆轉換表
func DelDataConv(conv nodeattr.EditConverter, o orm.Ormer) error {
	if len(conv.EditContent) == 0 {
		return nil
	}
	var arr []string
	for _, v := range conv.EditContent {
		arr = append(arr, v.DstBrowse)
	}
	sql := `DELETE FROM data_converter WHERE src_dev_name=? and dst_browse_name IN ` + MySQLArrayToQuestionStr(arr)
	_, err := o.Raw(sql, conv.Name, arr).Exec()
	return err
}

// DelDevInfo 刪除設備列表
func DelDevInfo(id int, o orm.Ormer) error {
	_, err := o.Raw("DELETE FROM device_info WHERE id=?", id).Exec()
	return err
}

// UpdateTemplate 更新DB範本列表資訊
func UpdateTemplate(temp nodeattr.EditTemplate, o orm.Ormer) error {
	_, err := o.Raw("UPDATE node_template SET name=?,modify_time=? WHERE id=?", temp.EditContent.Name, temp.EditContent.ModifyTime, temp.ID).Exec()
	return err
}

// GetAllNodeStruct 取得 parent_browse_name 底下 node
func GetAllNodeStruct(o orm.Ormer) (nodeArr []nodeattr.NodeStruct, err error) {
	rs := o.Raw("SELECT * FROM node_structure")
	_, err = rs.QueryRows(&nodeArr)
	return nodeArr, err
}
