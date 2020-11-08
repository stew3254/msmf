package main

import (
  "fmt"
  "os"
  "log"
  "strings"
  "strconv"
  "time"
  "github.com/jinzhu/gorm"
  
  _ "github.com/jinzhu/gorm/dialects/postgres"
)

func connectDB(dbType string) (*gorm.DB, error) {
	// Get user credentials
	dbTypeUpper := strings.ToUpper(dbType)

	user, exists := os.LookupEnv(dbTypeUpper+"_USER")
	checkExists(exists, "Couldn't find database user")
	password, exists := os.LookupEnv(dbTypeUpper+"_PASSWORD")
	checkExists(exists, "Couldn't find database password")

	// Get database params
	dbServer, exists := os.LookupEnv(dbTypeUpper+"_SERVER")
	checkExists(exists, "Couldn't find database server")
	dbPort, exists := os.LookupEnv(dbTypeUpper+"_PORT")
	checkExists(exists, "Couldn't find database port")
	dbName, exists := os.LookupEnv(dbTypeUpper+"_DB")
	checkExists(exists, "Couldn't find database name")
	connectionString := fmt.Sprintf(
		"sslmode=disable host=%s port=%s dbname=%s user=%s password=%s",
		dbServer,
		dbPort,
		dbName,
		user,
		password,
  )
  
	// Check how many times to try the db before quitting
	attemptsStr, exists := os.LookupEnv("DB_ATTEMPTS")
	if !exists {
		attemptsStr = "5"
	}

	attempts, err := strconv.Atoi(attemptsStr)
	if err != nil {
		attempts = 5
	}

	timeoutStr, exists := os.LookupEnv("DB_CONNECTION_TIMEOUT")
	if !exists {
		timeoutStr = "5"
	}
	timeout, err := strconv.Atoi(timeoutStr)
	if err != nil {
		timeout = 5
	}

	for i := 1; i <= attempts; i++ {
		db, err = gorm.Open(dbType, connectionString)
		if err != nil {
			if i != attempts {
				log.Printf(
					"WARNING: Could not connect to db on attempt %d. Trying again in %d seconds.\n",
					i,
					timeout,
				)
			} else {
				return db, fmt.Errorf("could not connect to db after %d attempts", attempts)
			}
			time.Sleep(time.Duration(timeout) * time.Second)
		} else {
			// No error to worry about
			break
		}
	}
	log.Println("Connection to db succeeded!")
	return db, nil
}