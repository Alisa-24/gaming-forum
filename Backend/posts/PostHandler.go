package posts

import (
	db "forum/Backend/DB"
	"forum/Backend/errors"
	"forum/Backend/login"
	"html/template"
	"net/http"
	"strings"
	"time"
)

var categories = []string{"minecraft", "online games", "souls games", "general", "story games"}

// PostHandler handles GET and POST requests for creating posts
func PostHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/createpost.html")
	if err != nil {
		errors.InternalServerError(w, r, "Error loading template: "+err.Error())
		return
	}

	session, _ := login.GetSessionFromRequest(r)
	if session.IsGuest || session.Username == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if r.Method == http.MethodGet {
		tmpl.Execute(w, map[string]interface{}{"Categories": categories})
		return
	}

	if r.Method != http.MethodPost {
		errors.MethodNotAllowed(w, r, "Method not allowed")
		return
	}

	if err := r.ParseForm(); err != nil {
		errors.InternalServerError(w, r, "Error parsing form: "+err.Error())
		return
	}

	title := strings.ReplaceAll(r.FormValue("title"), "\n", " ")
	content := r.FormValue("content")
	categoriesSelected := r.Form["category[]"]

	if title == "" || content == "" || len(categoriesSelected) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		tmpl.Execute(w, map[string]interface{}{"Error": "Title, content, and at least one category required", "Categories": categories})
		return
	}

	if len(title) > 50 || len(content) > 1000 {
		w.WriteHeader(http.StatusBadRequest)
		tmpl.Execute(w, map[string]interface{}{"Error": "Too long title or content", "Categories": categories})
		return
	}

	// Normalize categories
	for i, c := range categoriesSelected {
		categoriesSelected[i] = strings.ToLower(strings.TrimSpace(c))
	}

	// Validate categories
	for _, c := range categoriesSelected {
		valid := false
		for _, cat := range categories {
			if c == cat {
				valid = true
				break
			}
		}
		if !valid {
			w.WriteHeader(http.StatusBadRequest)
			tmpl.Execute(w, map[string]interface{}{"Error": "Invalid category: " + c, "Categories": categories})
			return
		}
	}

	userID, err := db.GetUserIDByUsername(db.DB, session.Username)
	if err != nil {
		errors.InternalServerError(w, r, "Error finding user ID: "+err.Error())
		return
	}

	postID, err := db.CreatePost(db.DB, userID, title, content, time.Now())
	if err != nil {
		errors.InternalServerError(w, r, "Error saving post: "+err.Error())
		return
	}

	for _, c := range categoriesSelected {
		if err := db.EnsureCategory(db.DB, c); err != nil {
			errors.InternalServerError(w, r, "Error saving category: "+err.Error())
			return
		}
		categoryID, err := db.GetCategoryID(db.DB, c)
		if err != nil {
			errors.InternalServerError(w, r, "Error fetching category ID: "+err.Error())
			return
		}
		if err := db.LinkPostToCategory(db.DB, postID, categoryID); err != nil {
			errors.InternalServerError(w, r, "Error linking post to category: "+err.Error())
			return
		}
	}

	http.Redirect(w, r, "/homePage", http.StatusSeeOther)
}
