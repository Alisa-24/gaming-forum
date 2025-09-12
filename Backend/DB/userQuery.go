package db

import "database/sql"

type User struct {
	ID       int
	Username string
	Email    string
	Password string
}

func GetUserByIdentifier(conn *sql.DB, identifier string) (*User, error) {
	var u User
	err := conn.QueryRow(`
		SELECT id, username, password, email
		FROM users
		WHERE LOWER(username)=LOWER(?) OR LOWER(email)=LOWER(?)
	`, identifier, identifier).Scan(&u.ID, &u.Username, &u.Password, &u.Email)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func InsertUser(conn *sql.DB, username, email, hashedPassword string) error {
	_, err := conn.Exec(`INSERT INTO users (username, email, password) VALUES (?, ?, ?)`,
		username, email, hashedPassword)
	return err
}

func UsernameExists(conn *sql.DB, username string) (bool, error) {
	var exists bool
	err := conn.QueryRow(`SELECT EXISTS(SELECT 1 FROM users WHERE LOWER(username)=LOWER(?))`, username).Scan(&exists)
	return exists, err
}

func EmailExists(conn *sql.DB, email string) (bool, error) {
	var exists bool
	err := conn.QueryRow(`SELECT EXISTS(SELECT 1 FROM users WHERE LOWER(email)=LOWER(?))`, email).Scan(&exists)
	return exists, err
}

func GetUserIDByUsername(conn *sql.DB, username string) (int, error) {
	var id int
	err := conn.QueryRow(`SELECT id FROM users WHERE username = ?`, username).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

// GetUsernameByID fetches the username for a given user ID
func GetUsernameByID(conn *sql.DB, userID int) (string, error) {
	var username string
	err := conn.QueryRow(`SELECT username FROM users WHERE id = ?`, userID).Scan(&username)
	if err != nil {
		return "", err
	}
	return username, nil
}

// UserExists checks if a user with the given ID exists in the database
func UserExists(conn *sql.DB, userID int) (bool, error) {
	var exists bool
	err := conn.QueryRow(`SELECT EXISTS(SELECT 1 FROM users WHERE id = ?)`, userID).Scan(&exists)
	return exists, err
}
