package main

import (
	"database/sql"
	"os"
	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

func initDB() {
	// 2.3 Create logs directory if it doesn't exist
	os.MkdirAll("logs", 0755)

	// 2.3 Create order.db if it doesn't exist
	var err error
	db, err = sql.Open("sqlite3", "order.db")
	if err != nil {
		panic(err)
	}

	// 5.3 SQL requirement: Create order_items table
	query := `
	CREATE TABLE IF NOT EXISTS order_items (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		order_no TEXT NOT NULL,
		terminal_no TEXT NOT NULL,
		order_status TEXT NOT NULL,
		item_no INTEGER NOT NULL,
		menu_name TEXT NOT NULL,
		unit_price INTEGER NOT NULL,
		quantity INTEGER NOT NULL,
		subtotal INTEGER NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`
	_, err = db.Exec(query)
	if err != nil {
		panic(err)
	}
}