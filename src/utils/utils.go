package utils

import (
	"crypto/rand"
	"fmt"
	"time"

	"msmf/database"
)

//GenerateToken returns a token representing a logged in user
func GenerateToken() (string, time.Time) {
  b := make([]byte, 32)
  rand.Read(b)
  return fmt.Sprintf("%x", b), time.Now().Add(time.Hour)
}

//Verifies a token exists in the db and isn't expired
// It will then update any invalid tokens
func ValidateToken(token string) bool {
	user := database.User{}
	result := database.DB.Where("token = ?", token).First(&user)
	if result.Error != nil || user.TokenExpiration.Before(time.Now()) {
		return false
	}
	return true
}
