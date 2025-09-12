package db

import "database/sql"

// AddComment adds a comment to a post
func AddComment(conn *sql.DB, postID, userID int, content string) error {
	_, err := conn.Exec(`
		INSERT INTO comments (post_id, user_id, content)
		VALUES (?, ?, ?)
	`, postID, userID, content)
	return err
}
