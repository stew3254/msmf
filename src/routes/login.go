package routes

import (
	"net/http"
	"golang.org/x/crypto/bcrypt"
	"encoding/json"

	"msmf/database"
	"msmf/utils"
)

//Login handler
func Login(w http.ResponseWriter, r *http.Request) {
  defer r.Body.Close()

  username := r.FormValue("username")
	passwd := r.FormValue("password")

	user := database.User{}

	// Create a response to send back to the user
	resp := make(map[string]interface{})
	result := database.DB.Where("username = ?", username).First(&user)
	if result.Error != nil {
		resp["error"] = "Authentication failed"
		json, _ := json.Marshal(resp)
		w.Write(json)
		return
	}

	if bcrypt.CompareHashAndPassword(user.Password, []byte(passwd)) == nil {
		// Save token
		user.Token, user.TokenExpiration = utils.GenerateToken()
		database.DB.Save(&user)

		// Send response back to client
		resp["token"] = user.Token
		resp["expiration"] = user.TokenExpiration
	} else {
		resp["error"] = "Authentication failed"
	}
	json, _ := json.Marshal(resp)
	w.Write(json)
}
