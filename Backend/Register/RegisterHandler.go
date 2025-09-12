package register

import (
	db "forum/Backend/DB"
	"forum/Backend/errors"
	"forum/Backend/login"
	"html/template"
	"net/http"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

// RegisterHandler handles user registration
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	dbConn := db.DB
	tmpl, err := template.ParseFiles("Frontend/register.html")
	if err != nil {
		errors.InternalServerError(w, r, "Error loading template: "+err.Error())
		return
	}

	// End active session if user already logged in
	session, _ := login.GetSessionFromRequest(r)
	if session != nil && !session.IsGuest {
		if dbConn != nil {
			_ = db.DeleteSessionByToken(dbConn, session.Token)
		}

		// Expire session cookie
		http.SetCookie(w, &http.Cookie{
			Name:     "session_token",
			Value:    "",
			Path:     "/",
			HttpOnly: true,
			MaxAge:   -1, // expire immediately
		})
	}

	// Serve registration form on GET request
	if r.Method == http.MethodGet {
		tmpl.Execute(w, nil)
		return
	}

	// Only POST requests are allowed
	if r.Method != http.MethodPost {
		errors.MethodNotAllowed(w, r, "Method not allowed")
		return
	}

	// Collect form values
	username := r.FormValue("username")
	email := strings.ToLower(r.FormValue("email"))
	password := r.FormValue("password")

	// Validate required fields
	if username == "" || email == "" || password == "" {
		w.WriteHeader(http.StatusBadRequest)
		tmpl.Execute(w, map[string]string{"Error": "All fields are required"})
		return
	}

	// Check if username already exists
	usernameExists, err := db.UsernameExists(dbConn, username)
	if err != nil {
		errors.InternalServerError(w, r, "Database error: "+err.Error())
		return
	}
	if usernameExists {
		w.WriteHeader(http.StatusBadRequest)
		tmpl.Execute(w, map[string]string{"Error": "Username already exists"})
		return
	}

	// Check if email already exists
	emailExists, err := db.EmailExists(dbConn, email)
	if err != nil {
		errors.InternalServerError(w, r, "Database error: "+err.Error())
		return
	}
	if emailExists {
		w.WriteHeader(http.StatusBadRequest)
		tmpl.Execute(w, map[string]string{"Error": "Email already registered"})
		return
	}

	// Validate password, email, username
	if !isValidEmail(email) {
		w.WriteHeader(http.StatusBadRequest)
		tmpl.Execute(w, map[string]string{"Error": "Invalid email format"})
		return
	}
	if err := validateUsername(username); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		tmpl.Execute(w, map[string]string{"Error": err.Error()})
		return
	}
	if valid, err := isValidPassword(password); !valid {
		w.WriteHeader(http.StatusBadRequest)
		tmpl.Execute(w, map[string]string{"Error": err.Error()})
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		errors.InternalServerError(w, r, "Error hashing password: "+err.Error())
		return
	}

	// Convert []byte to string
	if err := db.InsertUser(dbConn, username, email, string(hashedPassword)); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		tmpl.Execute(w, map[string]string{"Error": "Database error: " + err.Error()})
		return
	}

	// Redirect to login page after successful registration
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
