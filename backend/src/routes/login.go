package routes

import (
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"

	"msmf/database"
	"msmf/utils"
)

// Login handler
func Login(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	passwd := r.FormValue("password")

	var user database.User

	// Create a response to send back to the user
	err := database.DB.Where("username = ?", username).Find(&user).Error
	if err != nil {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	if bcrypt.CompareHashAndPassword(user.Password, []byte(passwd)) == nil {
		// See if a token already exists that isn't expired
		now := time.Now()
		if now.Before(user.TokenExpiration) {
			// Set the expiration to 6 hours from now and return the old token
			user.TokenExpiration = now.Add(6 * time.Hour)
		} else {
			// Make a new token
			user.Token, user.TokenExpiration = utils.GenerateToken()
			database.DB.Save(&user)
		}

		http.SetCookie(w, &http.Cookie{
			Path:     "/",
			Name:     "token",
			Value:    user.Token,
			Expires:  user.TokenExpiration,
			Secure:   true,
			HttpOnly: true,
			SameSite: http.SameSiteStrictMode,
		})
		// http.Redirect(w, r, "/", http.StatusFound)
		// DEBUG
		http.ServeFile(w, r, "static/index.html")
		return
	}
	http.Error(w, "Forbidden", http.StatusForbidden)
}
