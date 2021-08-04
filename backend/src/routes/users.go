package routes

import "net/http"

// GetUsers lists all users on the server
func GetUsers(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "501 not implemented", http.StatusNotImplemented)
}

// GetUser gets information about a particular user
func GetUser(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "501 not implemented", http.StatusNotImplemented)
}
