package routes

import (
	"encoding/json"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm/clause"
	"log"
	"math"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
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
	// Make sure to seed the random
	// Could use crypto random, but this doesn't really need to be secure
	// It's not a secret, nor does it really matter
	rand.Seed(time.Now().Unix())

	// Loop until a valid code is found
	for {
		code := (rand.Int() % (int(math.Pow(10, 8)) - int(math.Pow(10, 7)-1))) + int(math.Pow(10, 7))
		// Give user 24 hours until referral code expires
		referral := database.Referrer{
			Code:       code,
			UserID:     user.ID,
			Expiration: time.Now().Add(24 * time.Hour),
		}
		err = database.DB.Create(&referral).Error
		if err == nil {
			resp["status"] = "Success"
			resp["code"] = code
			break
		} else {
			// TODO fix this
			log.Println(result.Error.Error())
		}
	}

	// Write out the body
	utils.WriteJSON(w, http.StatusOK, &resp)
}

// Refer allows users with proper permissions to send someone an invite code
// in order to make an account
func Refer(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.String(), "/")
	// Can't error due to regex checking on route
	code, _ := strconv.Atoi(parts[len(parts)-1])
	resp := make(map[string]interface{})
	if code < int(math.Pow(10, 7)) || code > int(math.Pow(10, 8))-1 {
		utils.ErrorJSON(w, http.StatusNotFound, "Not Found")
		return
	}

	// Grab the user and referral code
	var referrer database.Referrer
	err := database.DB.Preload(clause.Associations).First(&referrer, code).Error
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

	// Check if any of the following are missing
	if len(body["username"]) == 0 || len(body["password"]) == 0 {
		utils.ErrorJSON(w, http.StatusBadRequest, "Bad Request")
		return
	}

	// Generate the password hash
	hash, err := bcrypt.GenerateFromPassword([]byte(body["password"]), bcrypt.DefaultCost)
	// Something wrong with hashing
	if err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	// Attempt to make the user
	err = database.DB.Create(&database.User{
		Username:   body["username"],
		Password:   hash,
		ReferredBy: referrer.UserID,
	}).Error

	// Owner already exists
	if err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	// All good, remove one time referrer code
	database.DB.Delete(&referrer)
	resp["status"] = "Success"

	// Write out the body
	utils.WriteJSON(w, http.StatusOK, &resp)
}
