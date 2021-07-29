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

	user := database.User{}

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
			Name:     "token",
			Value:    user.Token,
			Expires:  user.TokenExpiration,
			Secure:   true,
			HttpOnly: true,
			SameSite: http.SameSiteStrictMode,
		})
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	http.Error(w, "Forbidden", http.StatusForbidden)
}

// ChangePassword handles updating a password as long as the user knows their current password
func ChangePassword(w http.ResponseWriter, r *http.Request) {
	user := database.User{}
	// Middleware already checks to see if the user is authenticated
	tokenCookie, _ := r.Cookie("token")
	token := tokenCookie.Value

	// Get the valid user
	database.DB.Where("users.token = ?", token).Find(&user)

	// Get all of the form fields
	currPass := r.FormValue("current_password")
	newPass := r.FormValue("new_password")

	// TODO make it so the new password must meet certain requirements to be a good password

	// Check to see if passwords don't match
	if bcrypt.CompareHashAndPassword(user.Password, []byte(currPass)) != nil {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Set the new password
	hash, err := bcrypt.GenerateFromPassword([]byte(newPass), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Update user password and save to database
	user.Password = hash
	database.DB.Save(&user)

	http.Redirect(w, r, "/", http.StatusFound)
}
