package db

import (
	"database/sql"
	"fmt"
	"forum/Backend/errors"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

// InitDB initializes the SQLite database and applies schema.sql
func InitDB() {
	var err error
	DB, err = sql.Open("sqlite3", "./forum.db")
	if err != nil {
		errors.InternalServerError(nil, nil, "Error opening database: "+err.Error())
		return
	}

	// Read schema.sql file
	schema, err := os.ReadFile("Backend/DB/schema.sql")
	if err != nil {
		errors.InternalServerError(nil, nil, "Error reading schema.sql: "+err.Error())
		return
	}

	// Execute schema
	_, err = DB.Exec(string(schema))
	if err != nil {
		errors.InternalServerError(nil, nil, "Error executing schema.sql: "+err.Error())
		return
	}

	fmt.Println("Database initialized successfully ✅")
}
