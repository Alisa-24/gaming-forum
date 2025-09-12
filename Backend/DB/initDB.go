package db

import (
	"database/sql"
	"forum/Backend/errors"
	"log"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB // Capitalized to export

func InitDB() {
	var err error
	DB, err = sql.Open("sqlite3", "./forum.db") // use relative path
	if err != nil {
		log.Fatal("Error connecting to the database:", err)
	}

	// Define schema as a string slice
	schema := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT NOT NULL UNIQUE,
			email TEXT NOT NULL UNIQUE,
			password TEXT NOT NULL
		);`,

		`CREATE TABLE IF NOT EXISTS sessions (
          token TEXT PRIMARY KEY,
          user_id INTEGER NOT NULL,
          created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
          expires_at TIMESTAMP,
          FOREIGN KEY (user_id) REFERENCES users(id)
        );`,

		`CREATE TABLE IF NOT EXISTS posts (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            user_id INTEGER NOT NULL,
            title TEXT NOT NULL,
            content TEXT NOT NULL,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            FOREIGN KEY (user_id) REFERENCES users(id)
        );`,

		`CREATE TABLE IF NOT EXISTS categories (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE
		);`,

		`CREATE TABLE IF NOT EXISTS post_categories (
			post_id INTEGER,
			category_id INTEGER,
			FOREIGN KEY (post_id) REFERENCES posts(id),
			FOREIGN KEY (category_id) REFERENCES categories(id),
			PRIMARY KEY (post_id, category_id)
		);`,

		`CREATE TABLE IF NOT EXISTS comments (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			post_id INTEGER NOT NULL,
			user_id INTEGER NOT NULL,
			content TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (post_id) REFERENCES posts(id),
			FOREIGN KEY (user_id) REFERENCES users(id)
		);`,

		`CREATE TABLE IF NOT EXISTS likes (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			post_id INTEGER,
			comment_id INTEGER,
			is_like BOOLEAN NOT NULL,
			FOREIGN KEY (user_id) REFERENCES users(id),
			FOREIGN KEY (post_id) REFERENCES posts(id),
			FOREIGN KEY (comment_id) REFERENCES comments(id),
			CHECK (
				(post_id IS NOT NULL AND comment_id IS NULL) OR
				(post_id IS NULL AND comment_id IS NOT NULL)
			),
				UNIQUE(user_id, post_id),
				UNIQUE(user_id, comment_id)
		);`,
	}

	// Execute each SQL statement
	for _, stmt := range schema {
		_, err := DB.Exec(strings.TrimSpace(stmt))
		if err != nil {
			log.Fatalf("Error executing statement: %s\n%v", stmt, err)
		}
	}
}

func InitCategories() {
	categories := []string{"general", "minecraft", "souls games", "online games", "story games"}
	for _, category := range categories {
		_, err := DB.Exec(`INSERT OR IGNORE INTO categories (name) VALUES (?)`, category)
		if err != nil {
			errors.InternalServerError(nil, nil, "Error initializing categories: "+err.Error())
		}
	}
}
