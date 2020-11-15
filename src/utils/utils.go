package utils

import (
	"crypto/rand"
	"fmt"
	"time"

	"msmf/database"
)

// RefInt returns reference to the int suppied
func RefInt(i int) *int {
	return &i
}

// GenerateToken returns a token representing a logged in user
func GenerateToken() (string, time.Time) {
  b := make([]byte, 32)
  rand.Read(b)
  return fmt.Sprintf("%x", b), time.Now().Add(time.Hour)
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
	FKTable string
	Perms interface{} // Can be of type string or []string
	PermTable string
	Search interface{} // Can be of any type
	SearchCol string
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
		permCheck.PermTable[:len(permCheck.PermTable) - 1],
		permCheck.SearchTable,
		permCheck.SearchTable,
		permCheck.FKTable,
		permCheck.SearchTable[:len(permCheck.SearchTable) - 1],
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