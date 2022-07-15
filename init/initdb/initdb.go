package initdb

import (
	"FSRV_Edge/nodeattr"

	"github.com/astaxie/beego/orm"
)

func init() {

}
func dbRigister() {
	orm.RegisterDriver("sqlite", orm.DRSqlite)
	orm.RegisterDataBase("default", "sqlite3", "./edge.db")
	orm.RegisterModel(new(nodeattr.ConInfo), new(nodeattr.Template), new(nodeattr.TemplateRel), new(nodeattr.NodeStruct), new(nodeattr.Converter), new(nodeattr.DevInfo))
	orm.RunSyncdb("default", false, false)
}
