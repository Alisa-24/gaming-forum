package db

import "database/sql"

type LikeTarget struct {
	ID     int
	UserID int
	IsPost bool
	IsLike bool
}

func ToggleLike(conn *sql.DB, target LikeTarget) error {
	var existingID int
	var existingVal bool

	if target.IsPost {
		err := conn.QueryRow(`
			SELECT id, is_like FROM likes
			WHERE user_id=? AND post_id=? AND comment_id IS NULL
		`, target.UserID, target.ID).Scan(&existingID, &existingVal)
		switch err {
		case sql.ErrNoRows:
			_, err = conn.Exec(`INSERT INTO likes (user_id, post_id, is_like) VALUES (?, ?, ?)`,
				target.UserID, target.ID, target.IsLike)
		case nil:
			if existingVal == target.IsLike {
				_, err = conn.Exec(`DELETE FROM likes WHERE id=?`, existingID)
			} else {
				_, err = conn.Exec(`UPDATE likes SET is_like=? WHERE id=?`, target.IsLike, existingID)
			}
		default:
			return err
		}
		return err
	} else {
		err := conn.QueryRow(`
			SELECT id, is_like FROM likes
			WHERE user_id=? AND comment_id=? AND post_id IS NULL
		`, target.UserID, target.ID).Scan(&existingID, &existingVal)
		switch err {
		case sql.ErrNoRows:
			_, err = conn.Exec(`INSERT INTO likes (user_id, comment_id, is_like) VALUES (?, ?, ?)`,
				target.UserID, target.ID, target.IsLike)
		case nil:
			if existingVal == target.IsLike {
				_, err = conn.Exec(`DELETE FROM likes WHERE id=?`, existingID)
			} else {
				_, err = conn.Exec(`UPDATE likes SET is_like=? WHERE id=?`, target.IsLike, existingID)
			}
		default:
			return err
		}
		return err
	}
}

func CheckPostExists(conn *sql.DB, postID int) (bool, error) {
	var exists int
	err := conn.QueryRow(`SELECT id FROM posts WHERE id=?`, postID).Scan(&exists)
	if err == sql.ErrNoRows {
		return false, nil
	}
	return err == nil, err
}

func CheckCommentExists(conn *sql.DB, commentID int) (bool, error) {
	var exists int
	err := conn.QueryRow(`SELECT id FROM comments WHERE id=?`, commentID).Scan(&exists)
	if err == sql.ErrNoRows {
		return false, nil
	}
	return err == nil, err
}
