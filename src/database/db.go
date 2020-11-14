package database

import (
  "fmt"
  "os"
  "log"
  "strings"
  "strconv"
  "time"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/driver/postgres"
  "golang.org/x/crypto/bcrypt"
)

// DB is a global db connection to be shared
var DB *gorm.DB

// checkExists checks if a value exists and fails if it doesn't
func checkExists(exists bool, msg string) {
	if !exists {
		log.Fatal(msg)
	}
}

//ConnectDB sets up the initial connection to the database along with retrying attempts
func ConnectDB(dbType string) error {
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
		DB, err = gorm.Open(postgres.Open(connectionString), &gorm.Config{})
		if err != nil {
			if i != attempts {
				log.Printf(
					"WARNING: Could not connect to db on attempt %d. Trying again in %d seconds.\n",
					i,
					timeout,
				)
			} else {
				return fmt.Errorf("could not connect to db after %d attempts", attempts)
			}
			time.Sleep(time.Duration(timeout) * time.Second)
		} else {
			// No error to worry about
			break
		}
	}
	log.Println("Connection to db succeeded!")
	return nil
}

// CreatePerms creates all server perms
func CreatePerms() {
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
	DB.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "name"}},
		DoUpdates: clause.AssignmentColumns([]string{"name", "description"}),
	}).Create(&userPerms)

	DB.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "name"}},
		DoUpdates: clause.AssignmentColumns([]string{"name", "description"}),
	}).Create(&serverPerms)
}

// MakeAdmin upserts the default admin account
func MakeAdmin() {
	passwd, exists := os.LookupEnv("ADMIN_PASSWORD")
	if !exists {
		log.Fatal("You must set an admin password")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(passwd), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal(err)
	}

	// Upsert into table permissions
	admin := User{
		Username: "admin",
		Password: hash,
	}
	DB.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "username"}},
		DoUpdates: clause.AssignmentColumns([]string{"password"}),
	}).Create(&admin)

	// Get admin permission
	userPerm := UserPerm{}
	result := DB.Where("name = 'administrator'").First(&userPerm)
	if result.Error != nil {
		log.Fatal(err)
	}

	// Actually add admin perms to admin user
	DB.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "user_id"}, {Name: "user_perm_id"}},
		DoNothing: true,
	}).Create(PermsPerUser{
		UserID: admin.ID,
		UserPermID: userPerm.ID,
	})
}

// CreateTables sets up the db
func CreateTables() {
	// Create all regular tables
	DB.AutoMigrate(
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
	CreatePerms()
}

// DropTables drops everything in the db
func DropTables() {
	// Drop tables in an order that won't invoke errors from foreign key constraints
	DB.Migrator().DropTable(&ModsPerServer{})
	DB.Migrator().DropTable(&PermsPerUser{})
	DB.Migrator().DropTable(&ServerPermsPerUser{})
	DB.Migrator().DropTable(&UserPlayer{})
	DB.Migrator().DropTable(&ServerLog{})
	DB.Migrator().DropTable(&PlayerLog{})
	DB.Migrator().DropTable(&WebLog{})
	DB.Migrator().DropTable(&ServerPerm{})
	DB.Migrator().DropTable(&Server{})
	DB.Migrator().DropTable(&UserPerm{})
	DB.Migrator().DropTable(&User{})
	DB.Migrator().DropTable(&Player{})
	DB.Migrator().DropTable(&Mod{})
	DB.Migrator().DropTable(&Version{})
	DB.Migrator().DropTable(&Game{})
}