package main

import (
  "fmt"
  "os"
  "log"
  "strings"
  "strconv"
  "time"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"msmf/utils"
  
	"gorm.io/driver/postgres"
)

func connectDB(dbType string) (*gorm.DB, error) {
	// Get user credentials
	dbTypeUpper := strings.ToUpper(dbType)

	user, exists := os.LookupEnv(dbTypeUpper+"_USER")
	utils.CheckExists(exists, "Couldn't find database user")
	password, exists := os.LookupEnv(dbTypeUpper+"_PASSWORD")
	utils.CheckExists(exists, "Couldn't find database password")

	// Get database params
	dbServer, exists := os.LookupEnv(dbTypeUpper+"_SERVER")
	utils.CheckExists(exists, "Couldn't find database server")
	dbPort, exists := os.LookupEnv(dbTypeUpper+"_PORT")
	utils.CheckExists(exists, "Couldn't find database port")
	dbName, exists := os.LookupEnv(dbTypeUpper+"_DB")
	utils.CheckExists(exists, "Couldn't find database name")
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
		db, err = gorm.Open(postgres.Open(connectionString), &gorm.Config{})
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

// Create all server perms
func createPerms(db *gorm.DB) {
	// User Permissions
	userPerms := []UserPerm{
		{
			Name: "administrator",
			Description: "Enables full control over all user permissions",
		},
		{
			Name: "create_server",
			Description: "Enables creation of servers and the deletion of your own servers",
		},
		{
			Name: "delete_server",
			Description: "Enables deletion of all servers regardless of server owner",
		},
		{
			Name: "manage_user_permission",
			Description: "Allows management of other users's permissions. You cannot add permissions to others that you do not have already",
		},
		{
			Name: "manage_server_permission",
			Description: "Enables the ability to modify all server permissions for all servers and users. Note, you cannot remove permissions from people who own servers",
		},
		{
			Name: "create_user",
			Description: "Enables the ability to add more users to the web portal",
		},
		{
			Name: "delete_user",
			Description: "Enables the ability to delete users from the web portal",
		},
	}

	// Server Permissions
	serverPerms := []ServerPerm{
		{
			Name: "administrator",
			Description: "Enables full control over all server permissions for a server",
		},
		{
			Name: "restart",
			Description: "Enables stopping and starting of the server",
		},
		{
			Name: "edit_configuration",
			Description: "Enables changing the port amoung other features",
		},
		{
			Name: "manage_mods",
			Description: "Enables adding and removing mods from the server",
		},
		{
			Name: "kick",
			Description: "Allows kicking of players from a server",
		},
		{
			Name: "ban",
			Description: "Allows banning of players from a server",
		},
		{
			Name: "view_logs",
			Description: "Enables viewing of server logs, but not being able to send commands",
		},
		{
			Name: "manage_server_console",
			Description: "Enables attaching to the server console directly in order to run commands. Note, this will make you a server operator as well on games that have support for that",
		},
	}

	// Upsert into table permissions
	db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "name"}},
		DoUpdates: clause.AssignmentColumns([]string{"name", "description"}),
	}).Create(&userPerms)

	db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "name"}},
		DoUpdates: clause.AssignmentColumns([]string{"name", "description"}),
	}).Create(&serverPerms)
}

// Creates all tables
func createTables(db *gorm.DB) {
	// Create all regular tables
	db.AutoMigrate(
		&Game{},
		&Version{},
		&Mod{},
		&Server{},
		&ServerPerm{},
		&User{},
		&UserPerm{},
		&Player{},
		&ModsPerServer{},
		&PermsPerUser{},
		&ServerPermsPerUser{},
		&UserPlayer{},
		&ServerLog{},
		&PlayerLog{},
		&WebLog{},
	)

	// Create base permissions
	createPerms(db)
}

// Drop all tables
func dropTables(db *gorm.DB) {
	// Drop tables in an order that won't invoke errors from foreign key constraints
	db.Migrator().DropTable(&ModsPerServer{})
	db.Migrator().DropTable(&PermsPerUser{})
	db.Migrator().DropTable(&ServerPermsPerUser{})
	db.Migrator().DropTable(&UserPlayer{})
	db.Migrator().DropTable(&ServerLog{})
	db.Migrator().DropTable(&PlayerLog{})
	db.Migrator().DropTable(&WebLog{})
	db.Migrator().DropTable(&ServerPerm{})
	db.Migrator().DropTable(&Server{})
	db.Migrator().DropTable(&UserPerm{})
	db.Migrator().DropTable(&User{})
	db.Migrator().DropTable(&Player{})
	db.Migrator().DropTable(&Mod{})
	db.Migrator().DropTable(&Version{})
	db.Migrator().DropTable(&Game{})
}