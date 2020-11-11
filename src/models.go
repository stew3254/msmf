package main

import (
	"time"

	"github.com/jinzhu/gorm"

	_ "github.com/jinzhu/gorm/dialects/postgres"
)

// Game Model
type Game struct {
	ID *int `gorm:"primaryKey; type:serial"`
	Name string `gorm:"type:varchar(64) not null unique"`
}

// Version Model
type Version struct {
	ID *int `gorm:"primaryKey; type:serial"`
	Tag string `gorm:"type:text not null"`
	GameID *int
	Game Game
}

// Mod Model
type Mod struct {
	ID *int `gorm:"primaryKey; type:serial"`
	URL string `gorm:"type: text not null"`
	Name string `gorm:"type: varchar(64) not null"`
	GameID *int `gorm:"not null"`
	Game Game 
	VersionID *int
	Version Version 
}

// Server Model
type Server struct {
	ID *int `gorm:"primaryKey; type:serial"`
	Port int16 `gorm:"not null; unique; check: Port < 65536, Port > 0"`
	GameID *int `gorm:"not null"`
	Game Game
	VersionID *int
	Version Version
}

// ServerPerm Model
type ServerPerm struct {
	ID *int `gorm:"primaryKey; type:serial"`
	Name string `gorm:"type: varchar(64) not null unique"`
}

// User Model. ReferredBy is self referencing Foreign Key
type User struct {
	ID *int `gorm:"primaryKey; type:serial"`
	Username string `gorm:"type: varchar(32) not null unique"`
	Password string `gorm:"type: varchar(128) not null"`
	ReferredBy *int
	Referrer *User
}

// UserPerm Model
type UserPerm struct {
	ID *int `gorm:"primaryKey; type:serial"`
	Name string `gorm:"type: varchar(64) not null unique"`
}

// Player Model
type Player struct {
	ID *int `gorm:"primaryKey; type:serial"`
	Name string `gorm:"type: varchar(64) not null unique"`
}

// ModsPerServer Model. Foriegn Key table
type ModsPerServer struct {
	ModID *int
	Mod Mod
	ServerID *int
	Server Server
}

// PermsPerUser Model. Foriegn Key table
type PermsPerUser struct {
	UserID *int
	User User
	UserPermID *int
	UserPerm UserPerm
}

// ServerPermsPerUser Model. Foriegn Key table
type ServerPermsPerUser struct {
	ServerID *int
	Server Server
	ServerPermID *int
	ServerPerm ServerPerm
	UserID *int
	User User
}

// UserPlayer Model. Foriegn Key table
type UserPlayer struct {
	UserID *int
	User User
	PlayerID *int
	Player Player
}

// ServerLog Model
type ServerLog struct {
	ID *int `gorm:"primaryKey; type:serial"`
	Time time.Time `gorm:"not null"`
	Command string `gorm:"type: text not null"`
	PlayerID *int `gorm:"not null"`
	Player Player
	ServerID *int `gorm:"not null"`
	Server Server
}

// PlayerLog Model
type PlayerLog struct {
	Time time.Time `gorm:"primaryKey; type:serial"`
	Action string `gorm:"type: text not null"`
	PlayerID *int `gorm:"not null"`
	Player Player
	ServerID *int `gorm:"not null"`
	Server Server
}

// WebLog Model
type WebLog struct {
	ID *int `gorm:"primaryKey; type:serial"`
	Time time.Time `gorm:"not null"`
	IP string `gorm:"type: varchar(128) not null"`
	Method string `gorm:"type: text not null"`
	StatusCode int `gorm:"not null"`
	QueryParams string `gorm:"type:json"`
	PostData string `gorm:"type:json"`
	Cookies string `gorm:"type:json"`
	UserID *int
	User User
}

func createTables(db *gorm.DB) {
	// Create all tables
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

	// Add rest of constraints
	db.Model(&Version{}).AddForeignKey("game_id", "games(id)", "CASCADE", "CASCADE")

	db.Model(&Mod{}).AddForeignKey("game_id", "games(id)", "CASCADE", "CASCADE")
	db.Model(&Mod{}).AddForeignKey("version_id", "versions(id)", "SET NULL", "CASCADE")
	db.Model(&Mod{}).AddUniqueIndex("mod_unique_index", "url", "name", "game_id", "version_id")

	db.Model(&Server{}).AddForeignKey("game_id", "games(id)", "CASCADE", "CASCADE")
	db.Model(&Server{}).AddForeignKey("version_id", "versions(id)", "SET NULL", "CASCADE")

	db.Model(&User{}).AddForeignKey("referred_by", "users(id)", "SET NULL", "CASCADE")

	db.Model(&ModsPerServer{}).AddForeignKey("mod_id", "mods(id)", "CASCADE", "CASCADE")
	db.Model(&ModsPerServer{}).AddForeignKey("server_id", "servers(id)", "CASCADE", "CASCADE")
	db.Model(&ModsPerServer{}).AddUniqueIndex("mods_per_server_unique_index", "mod_id", "server_id")

	db.Model(&PermsPerUser{}).AddForeignKey("user_id", "users(id)", "CASCADE", "CASCADE")
	db.Model(&PermsPerUser{}).AddForeignKey("user_perm_id", "user_perms(id)", "CASCADE", "CASCADE")
	db.Model(&PermsPerUser{}).AddUniqueIndex("perms_per_user_unique_index", "user_id", "user_perm_id")

	db.Model(&ServerPermsPerUser{}).AddForeignKey("user_id", "users(id)", "CASCADE", "CASCADE")
	db.Model(&ServerPermsPerUser{}).AddForeignKey("server_perm_id", "server_perms(id)", "CASCADE", "CASCADE")
	db.Model(&ServerPermsPerUser{}).AddForeignKey("server_id", "servers(id)", "CASCADE", "CASCADE")
	db.Model(&ServerPermsPerUser{}).AddUniqueIndex("server_perms_per_user_unique_index", "server_id", "server_perm_id", "user_id")

	db.Model(&UserPlayer{}).AddForeignKey("user_id", "users(id)", "CASCADE", "CASCADE")
	db.Model(&UserPlayer{}).AddForeignKey("player_id", "players(id)", "CASCADE", "CASCADE")
	db.Model(&UserPlayer{}).AddUniqueIndex("user_player_unique_index", "user_id", "player_id")

	db.Model(&ServerLog{}).AddForeignKey("player_id", "players(id)", "CASCADE", "CASCADE")
	db.Model(&ServerLog{}).AddForeignKey("server_id", "servers(id)", "CASCADE", "CASCADE")

	db.Model(&PlayerLog{}).AddForeignKey("player_id", "players(id)", "CASCADE", "CASCADE")
	db.Model(&PlayerLog{}).AddForeignKey("server_id", "servers(id)", "CASCADE", "CASCADE")

	db.Model(&WebLog{}).AddForeignKey("user_id", "users(id)", "SET NULL", "CASCADE")
}

func dropTables(db *gorm.DB) {
	// Drop tables in an order that won't invoke errors from foreign key constraints
	db.DropTableIfExists(
		&ModsPerServer{},
		&PermsPerUser{},
		&ServerPermsPerUser{},
		&UserPlayer{},
		&ServerLog{},
		&PlayerLog{},
		&WebLog{},
		&ServerPerm{},
		&Server{},
		&UserPerm{},
		&User{},
		&Player{},
		&Mod{},
		&Version{},
		&Game{},
	)
}