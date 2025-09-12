package login

import (
	db "forum/Backend/DB"
	"forum/Backend/errors"
	"net/http"
)

// LogoutHandler logs the user out by deleting the session and expiring the cookie
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	// Get the current session
	session, err := GetSessionFromRequest(r)
	if err != nil {
		errors.InternalServerError(w, r, "Error retrieving session: "+err.Error())
		return
	}

	// If guest, just redirect
	if session.IsGuest {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// Delete the session from the database
	err = db.DeleteSessionByToken(db.DB, session.Token)
	if err != nil {
		errors.InternalServerError(w, r, "Error deleting session: "+err.Error())
		return
	}

	// Expire the session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	})

	// Redirect to homepage
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
