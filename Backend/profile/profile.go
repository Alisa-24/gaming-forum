package profile

import (
	db "forum/Backend/DB"
	"forum/Backend/errors"
	"net/http"
	"strconv"
	"text/template"
)

func ProfileHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		errors.MethodNotAllowed(w, r, "Only GET requests are allowed")
		return
	}

	tmpl, err := template.ParseFiles("Frontend/profile.html")
	if err != nil {
		errors.InternalServerError(w, r, "Error loading template: "+err.Error())
		return
	}

	userIDStr := r.URL.Query().Get("id")
	if userIDStr == "" {
		errors.BadRequest(w, r, "User ID is required")
		return
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil || userID < 1 {
		errors.BadRequest(w, r, "Invalid User ID")
		return
	}

	dbConn := db.DB
	if dbConn == nil {
		errors.InternalServerError(w, r, "Database connection not available")
		return
	}

	exists, err := db.UserExists(dbConn, userID)
	if err != nil {
		errors.InternalServerError(w, r, "Database error: "+err.Error())
		return
	}
	if !exists {
		errors.NotFound(w, r, "User not found")
		return
	}

	// Get username
	username, err := db.GetUsernameByID(dbConn, userID)
	if err != nil {
		errors.InternalServerError(w, r, "Database error: "+err.Error())
		return
	}

	// Get created posts
	createdPosts, err := db.GetUserPosts(dbConn, userID, "created")
	if err != nil {
		errors.InternalServerError(w, r, "Error fetching created posts: "+err.Error())
		return
	}

	// Get liked posts
	likedPosts, err := db.GetUserPosts(dbConn, userID, "liked")
	if err != nil {
		errors.InternalServerError(w, r, "Error fetching liked posts: "+err.Error())
		return
	}

	// Render template
	err = tmpl.Execute(w, map[string]interface{}{
		"Username":     username,
		"CreatedPosts": createdPosts,
		"LikedPosts":   likedPosts,
	})
	if err != nil {
		errors.InternalServerError(w, r, "Template execution error: "+err.Error())
	}
}
