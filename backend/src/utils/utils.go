package utils

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"time"

	"msmf/database"
)

var UserPattern = regexp.MustCompile("[a-z0-9]+")
var ServerPattern = regexp.MustCompile("[a-z0-9- ]+")

// ToJSON converts to json and logs errors. Simply here to reduce code duplication
func ToJSON(v interface{}) []byte {
	out, err := json.Marshal(v)
	if err != nil {
		log.Println(err)
	}
	return out
}

// WriteJSON writes out an error in json form and sets appropriate headers. Should be used by API
func WriteJSON(w http.ResponseWriter, status int, content interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(ToJSON(content))
}

// ErrorJSON writes out an error in json form and sets appropriate headers. Should be used by API
func ErrorJSON(w http.ResponseWriter, status int, err string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	resp := make(map[string]interface{})
	resp["error"] = err
	_, _ = w.Write(ToJSON(&resp))
}

// GenerateRandom generates a random byte array
func GenerateRandom(size int) (b []byte) {
	b = make([]byte, size)
	_, err := rand.Read(b)
	if err != nil {
		log.Println(err)
	}
	return
}

// GenerateToken returns a token representing a logged-in user, and their expiration time
func GenerateToken() (string, time.Time) {
	b := GenerateRandom(32)
	return fmt.Sprintf("%x", b), time.Now().Add(6 * time.Hour)
}

// GenerateCode returns a url safe referral code
func GenerateCode() string {
	b := GenerateRandom(4)
	return base64.URLEncoding.EncodeToString(b)[:6]
}

// ValidateToken verifies a token exists in the db and isn't expired
// It will then update any invalid tokens
func ValidateToken(token string) bool {
	user := database.User{}
	result := database.DB.Where("token = ?", token).First(&user)
	if result.Error != nil || user.TokenExpiration.Before(time.Now()) {
		return false
	}
	return true
}

// PermCheck is a struct to build a dynamic query to tell if
// the search entity has proper permissions
// Perms must be of type string or []string
type PermCheck struct {
	FKTable     string
	Perms       interface{} // Can be of type string or []string
	PermTable   string
	Search      interface{} // Can be of any type
	SearchCol   string
	SearchTable string
}

// CheckPermissions takes the permission type (PermsPerUser vs ServerPermsPerUser)
// The permission(s) as a string or []string
func CheckPermissions(permCheck *PermCheck) bool {
	// Build the query string
	query := fmt.Sprintf(`
		SELECT COUNT(*) FROM %s
		INNER JOIN %s
			ON %s.id = %s.%s_id
		INNER JOIN %s
			ON %s.id = %s.%s_id
		WHERE %s.%s = ?
		AND %s.name in ?`,
		permCheck.FKTable,
		permCheck.PermTable,
		permCheck.PermTable,
		permCheck.FKTable,
		permCheck.PermTable[:len(permCheck.PermTable)-1],
		permCheck.SearchTable,
		permCheck.SearchTable,
		permCheck.FKTable,
		permCheck.SearchTable[:len(permCheck.SearchTable)-1],
		permCheck.SearchTable,
		permCheck.SearchCol,
		permCheck.PermTable,
	)

	var perms []string
	switch p := permCheck.Perms.(type) {
	case string:
		perms = []string{
			"administrator",
			p,
		}
		break
	case []string:
		perms = append(p, "administrator")
	}

	result := database.DB.Raw(query, permCheck.Search, perms)
	if result.Error != nil {
		return false
	}

	var count int
	result.Scan(&count)
	return count > 0
}

// GetPermissions takes the permission type (PermsPerUser vs ServerPermsPerUser)
// The permission(s) as a string or []string
func GetPermissions(permCheck *PermCheck) bool {
	// Build the query string
	query := fmt.Sprintf(`
		SELECT * FROM %s
		INNER JOIN %s
			ON %s.id = %s.%s_id
		INNER JOIN %s
			ON %s.id = %s.%s_id
		WHERE %s.%s = ?
		AND %s.name in ?`,
		permCheck.FKTable,
		permCheck.PermTable,
		permCheck.PermTable,
		permCheck.FKTable,
		permCheck.PermTable[:len(permCheck.PermTable)-1],
		permCheck.SearchTable,
		permCheck.SearchTable,
		permCheck.FKTable,
		permCheck.SearchTable[:len(permCheck.SearchTable)-1],
		permCheck.SearchTable,
		permCheck.SearchCol,
		permCheck.PermTable,
	)

	var perms []string
	switch p := permCheck.Perms.(type) {
	case string:
		perms = []string{
			"administrator",
			p,
		}
		break
	case []string:
		perms = append(p, "administrator")
	}

	result := database.DB.Raw(query, permCheck.Search, perms)
	if result.Error != nil {
		return false
	}

	var count int
	result.Scan(&count)
	return count > 0
}
