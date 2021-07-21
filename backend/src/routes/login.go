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
	result := database.DB.Where("username = ?", username).First(&user)
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
