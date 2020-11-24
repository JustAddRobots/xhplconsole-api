package db

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

func Conn() (*sqlx.DB, error) {
	dblogin := "root:b1cker1ng" +
		"@tcp(hosaka.local)" +
		"/xhplconsole?parseTime=true"
	db, err := sqlx.Connect("mysql", dblogin)
	return db, err
}
