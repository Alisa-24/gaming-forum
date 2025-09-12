package login

import (
	db "forum/Backend/DB"
	"forum/Backend/errors"
	"html/template"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"
)

const sessionDuration = 24 * time.Hour

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	dbConn := db.DB
	tmpl, err := template.ParseFiles("Frontend/login.html")
	if err != nil {
		errors.InternalServerError(w, r, "Error loading template: "+err.Error())
		return
	}

	// 1️⃣ If user already has a valid session, redirect
	if cookie, err := r.Cookie("session_token"); err == nil {
		sessionRecord, _ := db.GetSessionByToken(dbConn, cookie.Value)
		if sessionRecord != nil {
			http.Redirect(w, r, "/homePage", http.StatusSeeOther)
			return
		}
	}

	// 2️⃣ Serve login page on GET
	if r.Method == http.MethodGet {
		tmpl.Execute(w, nil)
		return
	}

	// Only POST allowed
	if r.Method != http.MethodPost {
		errors.MethodNotAllowed(w, r, "Method not allowed")
		return
	}

	// 3️⃣ Parse form values
	identifier := r.FormValue("identifier")
	password := r.FormValue("password")
	if identifier == "" || password == "" {
		w.WriteHeader(http.StatusBadRequest)
		tmpl.Execute(w, map[string]string{"Error": "All fields are required"})
		return
	}

	// 4️⃣ Fetch user by username/email
	user, err := db.GetUserByIdentifier(dbConn, identifier)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		tmpl.Execute(w, map[string]string{"Error": "User does not exist"})
		return
	}

	// 5️⃣ Verify password
	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)) != nil {
		w.WriteHeader(http.StatusBadRequest)
		tmpl.Execute(w, map[string]string{"Error": "Invalid password"})
		return
	}

	// 6️⃣ Try to delete any existing sessions for this user, but continue even if it fails
	err = db.DeleteUserSessions(dbConn, user.ID)
	if err != nil {
		// Just log the error but continue with login process
		// This ensures we don't show 500 errors to users
		// and allows a new session to be created
		// The old session will eventually expire anyway
		http.SetCookie(w, &http.Cookie{
			Name:     "session_token",
			Value:    "",
			Path:     "/",
			HttpOnly: true,
			MaxAge:   -1, // expire immediately
		})
	}

	// 7️⃣ Create a new session token
	token, err := GenerateToken()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		tmpl.Execute(w, map[string]string{"Error": "Token generation failed: " + err.Error()})
		return
	}

	expiration := time.Now().Add(sessionDuration)
	if err := db.CreateSession(dbConn, user.ID, token, expiration); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		tmpl.Execute(w, map[string]string{"Error": "Failed to create session: " + err.Error()})
		return
	}

	// 8️⃣ Set session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Expires:  expiration,
	})

	// 9️⃣ Redirect to home page
	http.Redirect(w, r, "/homePage", http.StatusSeeOther)
}
