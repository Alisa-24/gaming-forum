package db

import (
	"database/sql"
	"strings"
	"time"
)

// -------------------- Structs --------------------

type PostShow struct {
	ID                 int
	Username           string
	Title              string
	Content            string
	Categories         []string
	CreatedAt          time.Time
	CreatedAtFormatted string // Add nice readable format
	Likes              int
	Dislikes           int
	Comments           int  // total number of comments
	UserLiked          *int // nil = not liked, 0 = disliked, 1 = liked
}

type Comment struct {
	ID        int
	Username  string
	Content   string
	CreatedAt time.Time
	Likes     int
	Dislikes  int
}

// -------------------- Post Functions --------------------

// GetUserPosts fetches posts created or liked by a user
func GetUserPosts(conn *sql.DB, userID int, fetchType string) ([]PostShow, error) {
	var query string

	switch fetchType {
	case "created":
		query = `
			SELECT p.id, u.username, p.title, p.content, p.created_at,
				GROUP_CONCAT(c.name, ',') AS categories
			FROM posts p
			JOIN users u ON p.user_id = u.id
			LEFT JOIN post_categories pc ON p.id = pc.post_id
			LEFT JOIN categories c ON pc.category_id = c.id
			WHERE p.user_id = ?
			GROUP BY p.id
			ORDER BY p.created_at DESC
		`
	case "liked":
		query = `
			SELECT p.id, u.username, p.title, p.content, p.created_at,
				GROUP_CONCAT(c.name, ',') AS categories
			FROM likes l
			JOIN posts p ON l.post_id = p.id
			JOIN users u ON p.user_id = u.id
			LEFT JOIN post_categories pc ON p.id = pc.post_id
			LEFT JOIN categories c ON pc.category_id = c.id
			WHERE l.user_id = ? AND l.is_like = 1
			GROUP BY p.id
			ORDER BY p.created_at DESC
		`
	default:
		return nil, nil
	}

	rows, err := conn.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []PostShow
	for rows.Next() {
		var p PostShow
		var created time.Time
		var categories sql.NullString
		if err := rows.Scan(&p.ID, &p.Username, &p.Title, &p.Content, &created, &categories); err != nil {
			return nil, err
		}
		p.CreatedAt = created
		p.CreatedAtFormatted = created.Format("Jan 2, 2006 at 3:04 PM") // Add nice readable format

		if categories.Valid && categories.String != "" {
			p.Categories = parseCategories(categories.String)
		} else {
			p.Categories = []string{}
		}

		// Fetch likes/dislikes/comments
		_ = conn.QueryRow(`SELECT COUNT(*) FROM likes WHERE post_id=? AND is_like=1`, p.ID).Scan(&p.Likes)
		_ = conn.QueryRow(`SELECT COUNT(*) FROM likes WHERE post_id=? AND is_like=0`, p.ID).Scan(&p.Dislikes)
		_ = conn.QueryRow(`SELECT COUNT(*) FROM comments WHERE post_id=?`, p.ID).Scan(&p.Comments)

		posts = append(posts, p)
	}

	return posts, nil
}

// GetPostWithCategories fetches a post with its categories and likes/dislikes
func GetPostWithCategories(conn *sql.DB, postID int) (*PostShow, error) {
	var post PostShow
	var categoriesStr sql.NullString
	var createdAt time.Time

	query := `
		SELECT p.id, u.username, p.title, p.content, p.created_at,
			   GROUP_CONCAT(c.name, ',') AS categories
		FROM posts p
		JOIN users u ON p.user_id = u.id
		LEFT JOIN post_categories pc ON p.id = pc.post_id
		LEFT JOIN categories c ON pc.category_id = c.id
		WHERE p.id = ?
		GROUP BY p.id
	`
	err := conn.QueryRow(query, postID).Scan(
		&post.ID, &post.Username, &post.Title, &post.Content, &createdAt, &categoriesStr,
	)
	if err != nil {
		return nil, err
	}
	post.CreatedAt = createdAt

	if categoriesStr.Valid && categoriesStr.String != "" {
		post.Categories = parseCategories(categoriesStr.String)
	} else {
		post.Categories = []string{"general"}
	}

	_ = conn.QueryRow(`SELECT COUNT(*) FROM likes WHERE post_id=? AND is_like=1`, postID).Scan(&post.Likes)
	_ = conn.QueryRow(`SELECT COUNT(*) FROM likes WHERE post_id=? AND is_like=0`, postID).Scan(&post.Dislikes)
	_ = conn.QueryRow(`SELECT COUNT(*) FROM comments WHERE post_id=?`, postID).Scan(&post.Comments)

	return &post, nil
}

// GetCommentsForPost fetches all comments for a post with likes/dislikes
func GetCommentsForPost(conn *sql.DB, postID int) ([]Comment, error) {
	query := `
		SELECT c.id, u.username, c.content, c.created_at
		FROM comments c
		JOIN users u ON c.user_id = u.id
		WHERE c.post_id = ?
		ORDER BY c.created_at ASC
	`
	rows, err := conn.Query(query, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []Comment
	for rows.Next() {
		var c Comment
		var createdAt time.Time
		if err := rows.Scan(&c.ID, &c.Username, &c.Content, &createdAt); err != nil {
			return nil, err
		}
		c.CreatedAt = createdAt

		_ = conn.QueryRow(`SELECT COUNT(*) FROM likes WHERE comment_id=? AND is_like=1`, c.ID).Scan(&c.Likes)
		_ = conn.QueryRow(`SELECT COUNT(*) FROM likes WHERE comment_id=? AND is_like=0`, c.ID).Scan(&c.Dislikes)

		comments = append(comments, c)
	}
	return comments, rows.Err()
}

// -------------------- Category Functions --------------------

// parseCategories splits a comma-separated string into a slice
func parseCategories(s string) []string {
	parts := strings.Split(s, ",")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}

// EnsureCategory inserts a category if it doesn't exist
func EnsureCategory(conn *sql.DB, name string) error {
	_, err := conn.Exec(`INSERT OR IGNORE INTO categories (name) VALUES (?)`, name)
	return err
}

// GetCategoryID returns the ID of a category
func GetCategoryID(conn *sql.DB, name string) (int, error) {
	var id int
	err := conn.QueryRow(`SELECT id FROM categories WHERE name = ?`, name).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

// LinkPostToCategory associates a post with a category
func LinkPostToCategory(conn *sql.DB, postID, categoryID int) error {
	_, err := conn.Exec(`INSERT INTO post_categories (post_id, category_id) VALUES (?, ?)`, postID, categoryID)
	return err
}

// -------------------- Post Creation --------------------

// CreatePost inserts a new post and returns its ID
func CreatePost(conn *sql.DB, userID int, title, content string, createdAt time.Time) (int, error) {
	res, err := conn.Exec(
		`INSERT INTO posts (user_id, title, content, created_at) VALUES (?, ?, ?, ?)`,
		userID, title, content, createdAt,
	)
	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(id), nil
}
