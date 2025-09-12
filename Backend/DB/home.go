package db

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// -------------------- Fetch with stats --------------------

// FetchPostsByCategory fetches posts filtered by a single category
func FetchPostsByCategory(conn *sql.DB, category string, userID *int) ([]PostShow, error) {
	var posts []PostShow
	var rows *sql.Rows
	var err error

	if category == "" {
		rows, err = conn.Query(`
			SELECT p.id, u.username, p.title, p.content, p.created_at,
				COALESCE(GROUP_CONCAT(DISTINCT c.name), 'general') AS categories
			FROM posts p
			JOIN users u ON u.id = p.user_id
			LEFT JOIN post_categories pc ON p.id = pc.post_id
			LEFT JOIN categories c ON c.id = pc.category_id
			GROUP BY p.id
			ORDER BY p.created_at DESC
		`)
	} else {
		rows, err = conn.Query(`
			SELECT p.id, u.username, p.title, p.content, p.created_at,
				COALESCE(GROUP_CONCAT(DISTINCT c.name), 'general') AS categories
			FROM posts p
			JOIN users u ON u.id = p.user_id
			JOIN post_categories pc ON p.id = pc.post_id
			JOIN categories c ON pc.category_id = c.id
			WHERE LOWER(TRIM(c.name)) = LOWER(TRIM(?))
			GROUP BY p.id
			ORDER BY p.created_at DESC
		`, category)
	}

	if err != nil {
		return nil, fmt.Errorf("query error: %w", err)
	}
	defer rows.Close()

	var postIDs []int
	for rows.Next() {
		var p PostShow
		var categoriesStr string
		var created time.Time
		if err := rows.Scan(&p.ID, &p.Username, &p.Title, &p.Content, &created, &categoriesStr); err != nil {
			return nil, err
		}
		p.CreatedAt = created
		if categoriesStr != "" {
			parts := strings.Split(categoriesStr, ",")
			for i := range parts {
				parts[i] = strings.TrimSpace(parts[i])
			}
			p.Categories = parts
		} else {
			p.Categories = []string{"general"}
		}
		posts = append(posts, p)
		postIDs = append(postIDs, p.ID)
	}

	if len(posts) == 0 {
		return posts, nil
	}

	if err := populateStats(conn, posts, postIDs, userID); err != nil {
		return nil, err
	}

	return posts, nil
}

// FetchPostsByCategories fetches posts matching multiple categories
func FetchPostsByCategories(conn *sql.DB, categories []string, userID *int) ([]PostShow, error) {
	if len(categories) == 0 {
		return FetchPostsByCategory(conn, "", userID)
	}

	var cleanCats []string
	for _, c := range categories {
		cleaned := strings.ToLower(strings.TrimSpace(c))
		if cleaned != "" {
			cleanCats = append(cleanCats, cleaned)
		}
	}
	if len(cleanCats) == 0 {
		return FetchPostsByCategory(conn, "", userID)
	}

	placeholders := strings.Trim(strings.Repeat("?,", len(cleanCats)), ",")
	query := fmt.Sprintf(`
		SELECT DISTINCT p.id, u.username, p.title, p.content, p.created_at,
		       COALESCE(GROUP_CONCAT(DISTINCT c.name), 'general') AS categories
		FROM posts p
		JOIN users u ON u.id = p.user_id
		LEFT JOIN post_categories pc ON pc.post_id = p.id
		LEFT JOIN categories c ON pc.category_id = c.id
		WHERE p.id IN (
			SELECT pc2.post_id
			FROM post_categories pc2
			JOIN categories c2 ON pc2.category_id = c2.id
			WHERE LOWER(TRIM(c2.name)) IN (%s)
		)
		GROUP BY p.id
		ORDER BY p.created_at DESC
	`, placeholders)

	args := make([]interface{}, len(cleanCats))
	for i, c := range cleanCats {
		args[i] = c
	}

	rows, err := conn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []PostShow
	var postIDs []int
	for rows.Next() {
		var p PostShow
		var categoriesStr string
		var created time.Time
		if err := rows.Scan(&p.ID, &p.Username, &p.Title, &p.Content, &created, &categoriesStr); err != nil {
			return nil, err
		}
		p.CreatedAt = created
		if categoriesStr != "" {
			parts := strings.Split(categoriesStr, ",")
			for i := range parts {
				parts[i] = strings.TrimSpace(parts[i])
			}
			p.Categories = parts
		} else {
			p.Categories = []string{"general"}
		}
		posts = append(posts, p)
		postIDs = append(postIDs, p.ID)
	}

	if len(posts) == 0 {
		return posts, nil
	}

	if err := populateStats(conn, posts, postIDs, userID); err != nil {
		return nil, err
	}

	return posts, nil
}

// populateStats fills likes, dislikes, comments, and user-liked info
func populateStats(conn *sql.DB, posts []PostShow, postIDs []int, userID *int) error {
	if len(postIDs) == 0 {
		return nil
	}

	ids := make([]interface{}, len(postIDs))
	for i, id := range postIDs {
		ids[i] = id
	}
	placeholders := strings.Trim(strings.Repeat("?,", len(postIDs)), ",")

	// Helper to fetch counts
	fetchCounts := func(query string) (map[int]int, error) {
		result := make(map[int]int)
		rows, err := conn.Query(query, ids...)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		for rows.Next() {
			var id, count int
			if err := rows.Scan(&id, &count); err != nil {
				return nil, err
			}
			result[id] = count
		}
		return result, nil
	}

	likesMap, err := fetchCounts(fmt.Sprintf("SELECT post_id, COUNT(*) FROM likes WHERE is_like=1 AND post_id IN (%s) GROUP BY post_id", placeholders))
	if err != nil {
		return err
	}
	dislikesMap, err := fetchCounts(fmt.Sprintf("SELECT post_id, COUNT(*) FROM likes WHERE is_like=0 AND post_id IN (%s) GROUP BY post_id", placeholders))
	if err != nil {
		return err
	}
	commentsMap, err := fetchCounts(fmt.Sprintf("SELECT post_id, COUNT(*) FROM comments WHERE post_id IN (%s) GROUP BY post_id", placeholders))
	if err != nil {
		return err
	}

	userLikedMap := make(map[int]int)
	if userID != nil {
		args := append([]interface{}{*userID}, ids...)
		rows, err := conn.Query(fmt.Sprintf("SELECT post_id, is_like FROM likes WHERE user_id=? AND post_id IN (%s)", placeholders), args...)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var id int
			var isLike sql.NullBool
			if err := rows.Scan(&id, &isLike); err != nil {
				return err
			}
			if isLike.Valid {
				if isLike.Bool {
					userLikedMap[id] = 1
				} else {
					userLikedMap[id] = 0
				}
			}
		}
	}

	for i := range posts {
		id := posts[i].ID
		posts[i].Likes = likesMap[id]
		posts[i].Dislikes = dislikesMap[id]
		posts[i].Comments = commentsMap[id]
		if val, ok := userLikedMap[id]; ok {
			val := val
			posts[i].UserLiked = &val
		}
	}

	return nil
}
