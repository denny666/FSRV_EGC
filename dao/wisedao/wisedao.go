package wisedao

import (
	"github.com/jmoiron/sqlx"
)

var (
	db *sqlx.DB
)

func init() {
	var connectErr error
	db, connectErr = sqlx.Connect("sqlite3", "./edge.db")
	if connectErr != nil {
		panic(connectErr.Error())
	}
}
