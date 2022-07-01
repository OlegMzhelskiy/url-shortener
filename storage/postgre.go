package storage

import (
	"database/sql"
	"fmt"
)

type StoreDB struct {
	//db sql.DB
	dbdns string
}

func NewStoreDB(DBDNS string) *StoreDB {
	return &StoreDB{DBDNS}
}

func (store *StoreDB) Ping() bool {
	//DBDNS := "host=localhost dbname=shortener user=postgres password=123 sslmode=disable"
	db, err := sql.Open("postgres", store.dbdns)
	if err != nil {
		fmt.Println(err)
		return false
	}
	defer db.Close()
	if err = db.Ping(); err != nil {
		fmt.Println(err)
		return false
	}
	return true
}
