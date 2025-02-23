package main

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"log"
)

var db *sqlx.DB

func initDB() {
	var err error
	dsn := "host=localhost user=wallet_user password=wallet_password dbname=wallet_db sslmode=disable"
	db, err = sqlx.Connect("postgres", dsn)
	if err != nil {
		log.Fatalf("Ошибка подключения к БД: %v", err)
	}
	fmt.Println("Успешное подключение к БД")

	var err1 error
	_, err1 = db.Exec("CREATE TABLE IF NOT EXISTS users (id SERIAL PRIMARY KEY, username VARCHAR(255) NOT NULL UNIQUE)")
	if err1 != nil {
		panic(err1)
	}

	var err2 error
	_, err2 = db.Exec("CREATE TABLE IF NOT EXISTS wallets (id SERIAL PRIMARY KEY, user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE, balance NUMERIC NOT NULL DEFAULT 0)")
	if err2 != nil {
		panic(err2)
	}
}
