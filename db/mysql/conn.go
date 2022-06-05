package mysql

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

func init() {
	var err error
	// root:123456@tcp(127.0.0.1)
	db, err = sql.Open("mysql", "root:123456@tcp(127.0.0.1:3340)/fileserver?charset=utf8")
	db.SetMaxOpenConns(1000)
	err = db.Ping()
	if err != nil {
		fmt.Println("Failed to connect to mysql,err:", err.Error())
		os.Exit(1)
	}
}

// DBConn 初始化db
func DBConn() *sql.DB {
	return db
}
