package utils

import (
	"crypto/rand"
	"fmt"
	"time"

	"gorm.io/gorm/clause"

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

// CheckPermissions takes the permission type (PermsPerUser vs ServerPermsPerUser)
// The permission(s) as a string or []string
func CheckPermissions(permType interface{}, perms, v interface{}) bool {
	switch p := perms.(type) {
	case string:
		res := database.DB.Model(permType).Preload(clause.Associations).Where("name = ?", p).First(v)
		return res.Error == nil
	case []string:
		res := database.DB.Model(permType).Preload(clause.Associations).Where("name in ?", p).First(v)
		return res.Error == nil
	default:
		return false
	}
	return true
}