package login

import (
	db "forum/Backend/DB"
	"log"
	"net/http"
)

// Session represents an app-level user session
type Session struct {
	Token    string
	UserID   *int
	Username string
	IsGuest  bool
}

// GetSessionFromRequest retrieves a session from a request
func GetSessionFromRequest(r *http.Request) (*Session, error) {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		// No cookie → guest
		return &Session{
			Token:    "",
			UserID:   nil,
			Username: "Guest",
			IsGuest:  true,
		}, nil
	}

	record, err := db.GetSessionByToken(db.DB, cookie.Value)
	if err != nil {
		// Log DB error but return a guest session to avoid nil-pointer panics in callers.
		log.Println("GetSessionFromRequest: DB error:", err)
		return &Session{
			Token:    "",
			UserID:   nil,
			Username: "Guest",
			IsGuest:  true,
		}, nil
	}

	if record == nil {
		// Not found / expired → guest
		return &Session{
			Token:    "",
			UserID:   nil,
			Username: "Guest",
			IsGuest:  true,
		}, nil
	}

	s := &Session{
		Token:   record.Token,
		IsGuest: false,
	}

	if record.UserID.Valid {
		uid := int(record.UserID.Int64)
		s.UserID = &uid
	}

	if record.Username.Valid {
		s.Username = record.Username.String
	}

	return s, nil
}
