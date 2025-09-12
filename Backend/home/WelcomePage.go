package home

import (
	"forum/Backend/errors"
	"forum/Backend/login"
	"html/template"
	"net/http"
)

func WelcomePage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" { // Only exact "/"
		errors.NotFound(w, r, "Page not found")
		return
	}

	session, err := login.GetSessionFromRequest(r)
	if err != nil || session == nil {
		// No session → treat as guest
		session = &login.Session{
			IsGuest: true,
		}
	}

	if !session.IsGuest {
		// Logged-in user → redirect to home
		http.Redirect(w, r, "/homePage", http.StatusSeeOther)
		return
	}

	// Guest → show welcome page
	tmpl, err := template.ParseFiles("Frontend/welcome.html")
	if err != nil {
		http.Error(w, "Error loading template", http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}
