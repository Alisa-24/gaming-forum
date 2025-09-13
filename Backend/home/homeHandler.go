package home

import (
	db "forum/Backend/DB"
	"forum/Backend/errors"
	"forum/Backend/login"
	"html/template"
	"net/http"
	"os"
	"strings"
)

// Post represents a forum post with author info
type Post struct {
	ID            int
	Title         string
	Content       string
	CreatedAt     string
	Username      string
	Category      string
	Categories    []string
	Likes         int
	Dislikes      int
	CommentsCount int
	UserLiked     *int // nil = not liked, 0 = disliked, 1 = liked
}

// PageData holds data for templates
type PageData struct {
	Category           string
	Posts              []Post
	UserID             *int
	SelectedCategories []string
	FilterApplied      bool
}

// ---------------- DB Fetching Functions ----------------

func fetchPostsWithUserLikes(category string, userID *int) ([]Post, error) {
	ps, err := db.FetchPostsByCategory(db.DB, category, userID)
	if err != nil {
		return nil, err
	}
	return convertPostShow(ps), nil
}

func fetchFilteredPosts(categories []string, userID *int) ([]Post, error) {
	ps, err := db.FetchPostsByCategories(db.DB, categories, userID)
	if err != nil {
		return nil, err
	}
	return convertPostShow(ps), nil
}

// Convert DB PostShow struct to Post struct
func convertPostShow(ps []db.PostShow) []Post {
	var posts []Post
	for _, p := range ps {
		posts = append(posts, Post{
			ID:            p.ID,
			Title:         p.Title,
			Content:       p.Content,
			Username:      p.Username,
			Category:      strings.Join(p.Categories, ","),
			Categories:    p.Categories,
			Likes:         p.Likes,
			Dislikes:      p.Dislikes,
			CommentsCount: p.Comments,
			UserLiked:     p.UserLiked,
			CreatedAt:     p.CreatedAt.Format("Jan 02, 2006 3:04 PM"),
		})
	}
	return posts
}

// ---------------- Template Helpers ----------------

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

func renderTemplate(w http.ResponseWriter, r *http.Request, templatePath string, data interface{}) {
	if !fileExists(templatePath) {
		errors.InternalServerError(w, r, "Template file not found")
		return
	}

	funcMap := template.FuncMap{
		"upper": func(s string) string {
			if len(s) == 0 {
				return ""
			}
			return strings.ToUpper(string(s[0]))
		},
		"slice": func(s string, start, end int) string {
			if len(s) == 0 || start >= len(s) {
				return ""
			}
			if end > len(s) {
				end = len(s)
			}
			if start < 0 {
				start = 0
			}
			return s[start:end]
		},
		"eq": func(a, b interface{}) bool {
			return a == b
		},
		"ptrVal": func(p *int) int {
			if p == nil {
				return -1
			}
			return *p
		},
	}

	tmpl, err := template.New("").Funcs(funcMap).ParseFiles(templatePath)
	if err != nil {
		errors.InternalServerError(w, r, "Template parsing error")
		return
	}

	templateName := templatePath[strings.LastIndex(templatePath, "/")+1:]
	if err := tmpl.ExecuteTemplate(w, templateName, data); err != nil {
		errors.InternalServerError(w, r, "Template execution error")
	}
}

// ---------------- Rendering Functions ----------------

func renderPostsWithPageData(w http.ResponseWriter, r *http.Request, templatePath string, category string) {
	session, err := login.GetSessionFromRequest(r)
	if err != nil {
		// Instead of showing an error, treat as guest user
		session = &login.Session{
			IsGuest:  true,
			Username: "Guest",
		}
	}

	var userID *int
	if !session.IsGuest && session.UserID != nil {
		userID = session.UserID
	}

	posts, err := fetchPostsWithUserLikes(category, userID)
	if err != nil {
		errors.InternalServerError(w, r, "Failed to fetch posts")
		return
	}

	data := PageData{
		Category: category,
		Posts:    posts,
		UserID:   userID,
	}

	renderTemplate(w, r, templatePath, data)
}

func renderFilteredPosts(w http.ResponseWriter, r *http.Request, templatePath string) {
	session, err := login.GetSessionFromRequest(r)
	if err != nil {
		// Instead of showing an error, treat as guest user
		session = &login.Session{
			IsGuest:  true,
			Username: "Guest",
		}
	}

	var userID *int
	if !session.IsGuest && session.UserID != nil {
		userID = session.UserID
	}

	// Parse filter parameters
	var categories []string
	categories = r.URL.Query()["categories"]
	if len(categories) == 0 {
		if param := r.URL.Query().Get("categories"); param != "" {
			categories = strings.Split(param, ",")
		}
	}

	filterApplied := len(categories) > 0
	var posts []Post
	if filterApplied {
		posts, err = fetchFilteredPosts(categories, userID)
	} else {
		posts, err = fetchPostsWithUserLikes("", userID)
	}
	if err != nil {
		errors.InternalServerError(w, r, "Failed to load posts")
		return
	}

	data := PageData{
		Category:           "",
		Posts:              posts,
		UserID:             userID,
		SelectedCategories: categories,
		FilterApplied:      filterApplied,
	}

	renderTemplate(w, r, templatePath, data)
}

// ---------------- HTTP Handlers ----------------

func AllPosts(w http.ResponseWriter, r *http.Request) {
	if len(r.URL.Query()) > 0 {
		renderFilteredPosts(w, r, "templates/index.html")
	} else {
		renderPostsWithPageData(w, r, "templates/index.html", "")
	}
}

func MinecraftPosts(w http.ResponseWriter, r *http.Request) {
	renderPostsWithPageData(w, r, "templates/minecraft.html", "minecraft")
}

func SoulsPosts(w http.ResponseWriter, r *http.Request) {
	renderPostsWithPageData(w, r, "templates/souls.html", "souls games")
}

func OnlinePosts(w http.ResponseWriter, r *http.Request) {
	renderPostsWithPageData(w, r, "templates/online.html", "online games")
}

func StoryPosts(w http.ResponseWriter, r *http.Request) {
	renderPostsWithPageData(w, r, "templates/story.html", "story games")
}

func GeneralPosts(w http.ResponseWriter, r *http.Request) {
	renderPostsWithPageData(w, r, "templates/general.html", "general")
}

func AboutPage(w http.ResponseWriter, r *http.Request) {
	renderPostsWithPageData(w, r, "templates/about.html", "")
}
