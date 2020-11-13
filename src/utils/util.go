package utils

import (
  "encoding/base64"
  "golang.org/x/crypto/bcrypt"
	"log"
)

// CheckExists checks if a value exists and fails if it doesn't
func CheckExists(exists bool, msg string) {
	if !exists {
		log.Fatal(msg)
	}
}

//Hash takes a password and returns the base64 encoded hash
func Hash(password string) (string, error) {
  //Default cost of hash
  cost := 10
  passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), cost)
  if err != nil {
    return "", nil
  }
  encoded := base64.StdEncoding.EncodeToString([]byte(passwordHash))
  return encoded, nil
}