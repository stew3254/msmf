package routes

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm/clause"
	"net/http"
	"time"

	"msmf/database"
	"msmf/utils"
)

// GetReferrals shows all active referral links
func GetReferrals(w http.ResponseWriter, r *http.Request) {
	tokenCookie, _ := r.Cookie("token")
	token := tokenCookie.Value

	// See if the person has permissions to invite users in the first place
	hasPerms := utils.CheckPermissions(&utils.PermCheck{
		FKTable:     "perms_per_users",
		Perms:       "invite_user",
		PermTable:   "user_perms",
		Search:      token,
		SearchCol:   "token",
		SearchTable: "users",
	})

	// Don't let a user see all referral codes at once
	// If they really want to try, they can bruteforce the API until a GET doesn't fail
	if !hasPerms {
		utils.ErrorJSON(w, http.StatusForbidden, "Forbidden")
		return
	}

	// Remove expired codes first
	database.DB.Where("expiration < ?", time.Now()).Delete(&database.Referrer{})

	var referrers []database.Referrer
	err := database.DB.Preload(clause.Associations).Find(&referrers).Error
	if err != nil {
		utils.ErrorJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Write out the body
	utils.WriteJSON(w, http.StatusOK, &referrers)
}

// CreateReferral makes a new referral link
func CreateReferral(w http.ResponseWriter, r *http.Request) {
	// Response to send later
	resp := make(map[string]interface{})

	// Get Token
	tokenCookie, err := r.Cookie("token")
	// Invalid token
	if err != nil {
		utils.ErrorJSON(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	user := database.User{}
	result := database.DB.Where("token = ?", tokenCookie.Value).First(&user)
	// Error grabbing the user
	if result.Error != nil {
		utils.ErrorJSON(w, http.StatusUnauthorized, "Unauthorized")
		return
		// Check user permissions
	}
	hasPerms := utils.CheckPermissions(&utils.PermCheck{
		FKTable:     "perms_per_users",
		Perms:       "invite_user",
		PermTable:   "user_perms",
		Search:      user.Token,
		SearchCol:   "token",
		SearchTable: "users",
	})
	// Owner doesn't have correct perms
	if !hasPerms {
		utils.ErrorJSON(w, http.StatusUnauthorized, "Unauthorized")
		return
		// Create the invite code
	}

	// Loop until a valid code is found
	for {
		// Set count to something non-zero
		var count int64 = 1
		var code string
		// See if code already exists
		for count != 0 {
			// Make a new code
			code = utils.GenerateCode()
			database.DB.Where("referrers.code = ?", code).Count(&count)
		}

		// Give user 24 hours until referral code expires
		referral := database.Referrer{
			Code:       code,
			UserID:     user.ID,
			Expiration: time.Now().Add(24 * time.Hour),
		}
		err = database.DB.Create(&referral).Error
		if err == nil {
			resp["code"] = code
			break
		} else {
			utils.ErrorJSON(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	// Write out the body
	utils.WriteJSON(w, http.StatusOK, &resp)
}

// Refer allows users with proper permissions to send someone an invite code
// in order to make an account
func Refer(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	code := params["code"]

	// Grab the user and referral code
	var referrer database.Referrer
	err := database.DB.Preload(clause.Associations).Where("code = ?", code).First(&referrer).Error
	if err != nil {
		utils.ErrorJSON(w, http.StatusNotFound, "Not Found")
		return
		// Handle referral expiration
	} else if referrer.Expiration.Before(time.Now()) {
		// Remove the code from the db
		database.DB.Delete(&referrer)
		utils.ErrorJSON(w, http.StatusNotFound, "Not Found")
		return
	} else if r.Method == "GET" {
		utils.WriteJSON(w, http.StatusOK, &referrer)
		return
		// Must be Method POST
	}

	// Get JSON of body of POST
	body := make(map[string]string)
	err = json.NewDecoder(r.Body).Decode(&body)

	// Invalid data
	if err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, "Bad Request")
		return
	}

	// Make sure username matches pattern
	if !utils.UserPattern.MatchString(body["username"]) {
		utils.ErrorJSON(w, http.StatusBadRequest, "Invalid username format")
		return
	}

	username, exists := body["username"]
	if !exists {
		utils.ErrorJSON(w, http.StatusBadRequest, "Must supply a username")
		return
	}
	password, exists := body["password"]
	if !exists {
		utils.ErrorJSON(w, http.StatusBadRequest, "Must supply a password")
		return
	}

	// Check if any of the following are missing
	if len(username) == 0 || len(password) == 0 {
		utils.ErrorJSON(w, http.StatusBadRequest, "Cannot submit empty user or password")
		return
	}

	// Don't let the fields be too long
	if len(username) > 32 || len(password) > 64 {
		utils.ErrorJSON(w, http.StatusBadRequest,
			"Cannot exceed maximum length of username or password")
		return
	}

	// Cannot allow them to have the same name as the routes
	if username == "me" || password == "perm" {
		utils.ErrorJSON(w, http.StatusBadRequest, "Cannot set name to something reserved as a route")
		return
	}

	// Generate the password hash
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	// Something wrong with hashing
	if err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	// See if user set a display name
	display, exists := body["display"]
	if exists {
		if len(display) > 32 {
			utils.ErrorJSON(w, http.StatusBadRequest, "Cannot exceed maximum length of display name")
			return
		}
	}

	// Create user
	user := database.User{
		Username:   body["username"],
		Password:   hash,
		ReferredBy: referrer.UserID,
	}

	// Add the display if they have it
	if exists {
		user.Display = &display
	}

	// Attempt to make the user
	err = database.DB.Create(&user).Error

	// User already exists
	if err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	// All good, remove one time referrer code
	database.DB.Delete(&referrer)

	http.Error(w, "", http.StatusNoContent)
}
