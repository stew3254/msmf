package utils

import (
	"crypto/rand"
	"fmt"
	"log"
	"time"
)

// CheckExists checks if a value exists and fails if it doesn't
func CheckExists(exists bool, msg string) {
	if !exists {
		log.Fatal(msg)
	}
}

//GenerateToken returns a token representing a logged in user
func GenerateToken() (string, time.Time) {
  b := make([]byte, 32)
  rand.Read(b)
  return fmt.Sprintf("%x", b), time.Now().Add(time.Hour)
}