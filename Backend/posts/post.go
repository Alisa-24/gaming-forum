package posts

import (
	"database/sql"
	"fmt"
	db "forum/Backend/DB"
	"forum/Backend/errors"
	"html/template"
	"net/http"
	"sort"
	"strconv"
	"time"
)

// PostShow struct for template rendering
type PostShow struct {
	ID         int
	Username   string
	Title      string
	Content    string
	Categories []string
	CreatedAt  string
	Likes      int
	Dislikes   int
}

// Comment struct for template rendering
type Comment struct {
	ID        int
	Username  string
	Content   string
	CreatedAt string
	Likes     int
	Dislikes  int
}

// PostShowHandler handles GET /post?id=ID
func PostShowHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		errors.MethodNotAllowed(w, r, "Only GET requests are allowed")
		return
	}

	postIDStr := r.URL.Query().Get("id")
	if postIDStr == "" {
		errors.BadRequest(w, r, "Post ID is required")
		return
	}

	postID, err := strconv.Atoi(postIDStr)
	if err != nil || postID < 1 {
		errors.BadRequest(w, r, "Invalid Post ID")
		return
	}

	ShowPost(w, r, postID)
}

// ShowPost fetches post + comments + likes/dislikes and renders template
func ShowPost(w http.ResponseWriter, r *http.Request, postID int) {
	conn := db.DB
	if conn == nil {
		errors.InternalServerError(w, r, "Database connection not available")
		return
	}

	tmpl, err := template.ParseFiles("Frontend/post.html")
	if err != nil {
		errors.InternalServerError(w, r, fmt.Sprintf("Template parsing error: %v", err))
		return
	}

	loc := time.FixedZone("UTC+3", 3*3600)

	// Fetch post using DB layer
	p, err := db.GetPostWithCategories(conn, postID)
	if err != nil {
		if err == sql.ErrNoRows {
			errors.NotFound(w, r, "Post not found")
			return
		}
		errors.InternalServerError(w, r, fmt.Sprintf("Error fetching post: %v", err))
		return
	}

	post := PostShow{
		ID:         p.ID,
		Username:   p.Username,
		Title:      p.Title,
		Content:    p.Content,
		Categories: p.Categories,
		CreatedAt:  p.CreatedAt.In(loc).Format("Jan 02, 2006 3:04 PM"),
		Likes:      p.Likes,
		Dislikes:   p.Dislikes,
	}

	// Fetch comments using DB layer
	commentsRaw, err := db.GetCommentsForPost(conn, postID)
	if err != nil {
		errors.InternalServerError(w, r, fmt.Sprintf("Error fetching comments: %v", err))
		return
	}

	var comments []Comment
	for _, c := range commentsRaw {
		comments = append(comments, Comment{
			ID:        c.ID,
			Username:  c.Username,
			Content:   c.Content,
			CreatedAt: c.CreatedAt.In(loc).Format("Jan 02, 2006 3:04 PM"),
			Likes:     c.Likes,
			Dislikes:  c.Dislikes,
		})
	}

	sortCommentsByLikes(comments)

	err = tmpl.Execute(w, map[string]interface{}{
		"Post":     post,
		"Comments": comments,
	})
	if err != nil {
		errors.InternalServerError(w, r, fmt.Sprintf("Error rendering template: %v", err))
		return
	}
}

// sortCommentsByLikes sorts comments by likes (most liked first)
func sortCommentsByLikes(comments []Comment) {
	sort.SliceStable(comments, func(i, j int) bool {
		return comments[i].Likes > comments[j].Likes
	})
}
