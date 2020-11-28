package routes

import (
	"golang.org/x/crypto/bcrypt"
	"encoding/json"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"
	"math/rand"

	"msmf/database"
	"msmf/utils"
)

// GetReferrals shows all active referral links
func GetReferrals(w http.ResponseWriter, r *http.Request) {
	// Response to send later
	resp := make(map[string]interface{})

	// Remove expired codes first
	database.DB.Where("expiration < ?", time.Now()).Delete(&database.Referrer{})

	referrals := make([]map[string]interface{}, 0)
	rows, err := database.DB.Table("referrers").Select("users.username as referrer, referrers.code, referrers.expiration").Joins("inner join users on referrers.user_id = users.id").Rows()
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		resp["error"] = err.Error()
	} else {
		// Populate the referrers
		for rows.Next() {
			row := make(map[string]interface{})
			database.DB.ScanRows(rows, &row)
			referrals = append(referrals, row)
		}
		resp["referrers"] = referrals
	}

	// Write out the body
	out, err := json.Marshal(resp)
	if err != nil {
		log.Println(err)
	}
	w.Write(out)
}

// CreateReferral makes a new referral link
func CreateReferral(w http.ResponseWriter, r *http.Request) {
	// Response to send later
	resp := make(map[string]interface{})

	// Get Token
	tokenCookie, err := r.Cookie("token")
	// Invalid token
	if err != nil || !utils.ValidateToken(tokenCookie.Value) {
		w.WriteHeader(http.StatusUnauthorized)
		resp["error"] = "Unauthorized"
	} else {
		user := database.User{}
		result := database.DB.Where("token = ?", tokenCookie.Value).First(&user)
		// Error grabbing the user
		if result.Error != nil {
			w.WriteHeader(http.StatusUnauthorized)
			resp["error"] = "Unauthorized"
		// Check user permissions
		} else {
			hasPerms := utils.CheckPermissions(&utils.PermCheck{
				FKTable: "perms_per_users",
				Perms: "invite_user",
				PermTable: "user_perms",
				Search: user.Token,
				SearchCol: "token",
				SearchTable: "users",
			})
			// User doesn't have correct perms
			if !hasPerms {
				w.WriteHeader(http.StatusUnauthorized)
				resp["error"] = "Unauthorized"
			// Create the invite code
			} else {
				// Make sure to seed the random
				// Could use crypto random, but this doesn't really need to be secure
				// It's not a secret, nor does it really matter
				rand.Seed(time.Now().Unix())

				// Loop until a valid code is found
				for {
					code := (rand.Int() % (int(math.Pow(10, 8)) - int(math.Pow(10,7)-1))) + int(math.Pow(10, 7))
					referral := database.Referrer{
						Code: code,
						UserID: user.ID,
						Expiration: time.Now().Add(time.Hour*6),
					}
					result = database.DB.Create(&referral)
					if result.Error == nil {
						resp["status"] = "Success"
						resp["code"] = code
						break
					} else {
						log.Println(result.Error.Error())
					}
				}
			}
		}
	}


	// Write out the body
	out, err := json.Marshal(resp)
	if err != nil {
		log.Println(err)
	}
	w.Write(out)
}
// Refer allows users with proper permissions to send someone an invite code
// in order to make an account
func Refer(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.String(), "/")
	// Can't error due to regex checking on subrouter
	code, _ := strconv.Atoi(parts[len(parts)-1])
	resp := make(map[string]interface{})
	if code < int(math.Pow(10, 7)) || code > int(math.Pow(10, 8)) - 1 {
		resp["error"] = "Not Found"
		out, err := json.Marshal(resp)
		if err != nil {
			log.Println(err)
		}
		w.Write(out)
		return
	}

	// Grab the user and referral code
	referrer := database.Referrer{}
	result := database.DB.Preload("User").First(&referrer, code)
	if result.Error != nil {
		w.WriteHeader(http.StatusNotFound)
		resp["error"] = "Not Found"
	// Handle referral expiration
	} else if referrer.Expiration.Before(time.Now()) {
		// Remove the code from the db
		database.DB.Delete(&referrer)
		w.WriteHeader(http.StatusNotFound)
		resp["error"] = "Not Found"
	} else if r.Method == "GET" {
		resp["username"] = referrer.User.Username
		resp["expiration"] = referrer.Expiration
	// Must be Method PUT
	} else {
		// Get JSON of body of PUT
		body := make(map[string]string)
		err := json.NewDecoder(r.Body).Decode(&body)
		// Invalid data
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			resp["error"] = "Bad Request"
		// Could decode body
		} else {
			// Check if any of the following are missing
			if len(body["username"]) == 0 || len(body["password"]) == 0 {
				w.WriteHeader(http.StatusBadRequest)
				resp["error"] = "Bad Request"
			} else {
				// Generate the password hash
				hash, err := bcrypt.GenerateFromPassword([]byte(body["password"]), bcrypt.DefaultCost)
				// Something wrong with hashing
				if err != nil {
					w.WriteHeader(http.StatusBadRequest)
					resp["error"] = err.Error()
				} else {
					// Attempt to make the user
					result = database.DB.Create(&database.User{
						Username: body["username"],
						Password: hash,
						ReferredBy: referrer.UserID,
					})
					// User already exists
					if result.Error != nil {
						w.WriteHeader(http.StatusUnauthorized)
						resp["error"] = "user already exists"
					} else {
						// All good, remove one time referrer code
						database.DB.Delete(&referrer)
						resp["status"] = "Success"
					}
				}
			}
		}
	}

	// Finally write out the body
	out, err := json.Marshal(resp)
	if err != nil {
		log.Println(err)
	}
	w.Write(out)
}