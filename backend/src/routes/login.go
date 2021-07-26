package routes

import (
	"net/http"

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
	result := database.DB.Where("username = ?", username).Find(&user)
	if result.Error != nil {
		w.WriteHeader(http.StatusForbidden)
		http.ServeFile(w, r, "static/login.html")
		return
	}

	if bcrypt.CompareHashAndPassword(user.Password, []byte(passwd)) == nil {
		// Save token
		user.Token, user.TokenExpiration = utils.GenerateToken()
		database.DB.Save(&user)

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
	w.WriteHeader(http.StatusForbidden)
	http.ServeFile(w, r, "static/login.html")
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
	newPass := r.FormValue("password")
	confirmPass := r.FormValue("confirm_password")

	// Check to see if passwords don't match
	if bcrypt.CompareHashAndPassword(user.Password, []byte(currPass)) != nil {
		w.WriteHeader(http.StatusUnauthorized)
		http.ServeFile(w, r, "static/change-password.html")
		return
	}

	// See if new passwords don't match
	if newPass != confirmPass {
		w.WriteHeader(http.StatusBadRequest)
		http.ServeFile(w, r, "static/change-password.html")
		return
	}

	// Set the new password
	hash, err := bcrypt.GenerateFromPassword([]byte(newPass), bcrypt.DefaultCost)
	if err != nil {
		// Should really add a descriptive message
		w.WriteHeader(http.StatusBadRequest)
		http.ServeFile(w, r, "static/change-password.html")
		return
	}

	// Update user password and save to database
	user.Password = hash
	database.DB.Save(&user)

	http.Redirect(w, r, "/", http.StatusFound)
}
