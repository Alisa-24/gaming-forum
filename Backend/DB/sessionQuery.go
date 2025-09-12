package db

import (
	"database/sql"
	"time"
)

type SessionRecord struct {
	Token    string
	UserID   sql.NullInt64
	Username sql.NullString
	Expires  time.Time
}

func GetSessionByToken(conn *sql.DB, token string) (*SessionRecord, error) {
	var s SessionRecord
	err := conn.QueryRow(`
		SELECT s.token, s.user_id, u.username, s.expires_at
		FROM sessions s
		LEFT JOIN users u ON s.user_id = u.id
		WHERE s.token=? AND (s.expires_at IS NULL OR s.expires_at>CURRENT_TIMESTAMP)
	`, token).Scan(&s.Token, &s.UserID, &s.Username, &s.Expires)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func CreateSession(conn *sql.DB, userID int, token string, expires time.Time) error {
	_, err := conn.Exec(`INSERT INTO sessions (token, user_id, expires_at) VALUES (?, ?, ?)`,
		token, userID, expires)
	return err
}

func DeleteSessionByToken(conn *sql.DB, token string) error {
	_, err := conn.Exec(`DELETE FROM sessions WHERE token=?`, token)
	return err
}

func DeleteUserSessions(conn *sql.DB, userID int) error {
	_, err := conn.Exec(`DELETE FROM sessions WHERE user_id=?`, userID)
	return err
}
