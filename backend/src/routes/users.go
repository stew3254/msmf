package routes

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
	"msmf/database"
	"msmf/utils"
	"net/http"
)

// GetUsers lists all users on the server
func GetUsers(w http.ResponseWriter, r *http.Request) {
	var users []database.User
	database.DB.Find(&users)
	utils.WriteJSON(w, http.StatusOK, &users)
}

// GetUser gets information about a particular user
func GetUser(w http.ResponseWriter, r *http.Request) {
	tokenCookie, _ := r.Cookie("token")
	token := tokenCookie.Value

	// Get the user
	var user database.User

	params := mux.Vars(r)
	username, exists := params["user"]

	// The route must be for yourself
	if !exists {
		// return your own details
		database.DB.Where("token = ?", token).Find(&user)
	} else {
		err := database.DB.Where("username = ?", username).Find(&user).Error
		// See if the user doesn't exist
		if err != nil || user.ID == nil {
			utils.ErrorJSON(w, http.StatusNotFound, "User not found")
			return
		}
	}

	utils.WriteJSON(w, http.StatusOK, &user)
}

// UpdateUser handles updating user information
func UpdateUser(w http.ResponseWriter, r *http.Request) {
	user := database.User{}
	// Get token
	tokenCookie, _ := r.Cookie("token")
	token := tokenCookie.Value

	// Get JSON of body of PATCH
	body := make(map[string]string)
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	// Get the valid user
	database.DB.Where("users.token = ?", token).Find(&user)

	// Get all of the form fields
	currPass, currPassExists := body["current_password"]
	newPass, newPassExists := body["new_password"]

	// One is set but the other isn't
	if currPass != newPass {
		utils.ErrorJSON(w, http.StatusBadRequest,
			"Must supply both current_password and new_password together")
		return
	} else if currPassExists && newPassExists {
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

		// Update user password
		user.Password = hash
	}

	// See if user set a display name
	display, exists := body["display"]
	if exists {
		if len(display) > 32 {
			utils.ErrorJSON(w, http.StatusBadRequest, "Cannot exceed maximum length of display name")
			return
		}
		// If the user made it an empty string, remove it entirely
		if len(display) == 0 {
			user.Display = nil
		} else {
			// Set it to what they want
			user.Display = &display
		}
	}

	// Save the user to the database
	database.DB.Save(&user)

	http.Redirect(w, r, "/", http.StatusFound)
}
