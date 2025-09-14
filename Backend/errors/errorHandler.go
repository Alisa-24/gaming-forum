package errors

import (
	"html/template"
	"net/http"
)

// 400 Bad Request
func BadRequest(w http.ResponseWriter, r *http.Request, msg string) {
	tmpl, err := template.ParseFiles("templates/err/400.html")
	if err != nil {
		http.Error(w, "Error loading template: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusBadRequest)
	tmpl.Execute(w, map[string]string{"Error": msg})
}

// 404 Not Found
func NotFound(w http.ResponseWriter, r *http.Request, msg string) {
	tmpl, err := template.ParseFiles("templates/err/404.html")
	if err != nil {
		http.Error(w, "Error loading template: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNotFound)
	tmpl.Execute(w, map[string]string{"Error": msg})
}

// 405 Method Not Allowed
func MethodNotAllowed(w http.ResponseWriter, r *http.Request, msg string) {
	tmpl, err := template.ParseFiles("templates/err/405.html")
	if err != nil {
		http.Error(w, "Error loading template: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusMethodNotAllowed)
	tmpl.Execute(w, map[string]string{"Error": msg})
}

// 500 Internal Server Error
func InternalServerError(w http.ResponseWriter, r *http.Request, msg string) {
	tmpl, err := template.ParseFiles("templates/err/500.html")
	if err != nil {
		http.Error(w, "Error loading template: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusInternalServerError)
	tmpl.Execute(w, map[string]string{"Error": msg})
}
