package database

import (
	"time"
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
	Game Game `gorm:"constraint:OnUpdate:CASCADE,ONDELETE:CASCADE"`
}

// Mod Model
type Mod struct {
	ID *int `gorm:"primaryKey; type:serial; index:mod,unique"`
	URL string `gorm:"type: text not null; index:mod,unique"`
	Name string `gorm:"type: varchar(64) not null; index:mod,unique"`
	GameID *int `gorm:"not null; index:mod,unique"`
	Game Game `gorm:"constraint:OnUpdate:CASCADE,ONDELETE:CASCADE"`
	VersionID *int `gorm:"index:mod,unique"`
	Version Version `gorm:"constraint:OnUpdate:CASCADE,ONDELETE:SET NULL"`
}

// Server Model
type Server struct {
	ID *int `gorm:"primaryKey; type:serial"`
	Port int16 `gorm:"not null; unique; check: Port < 65536; check: Port > 0"`
	GameID *int `gorm:"not null"`
	Game Game `gorm:"constraint:OnUpdate:CASCADE,ONDELETE:CASCADE"`
	VersionID *int
	Version Version `gorm:"constraint:OnUpdate:CASCADE,ONDELETE:SET NULL"`
}

// ServerPerm Model
type ServerPerm struct {
	ID *int `gorm:"primaryKey; type:serial"`
	Name string `gorm:"type: varchar(64) not null unique"`
	Description string `gorm:"type: text"`
}

// User Model. ReferredBy is self referencing Foreign Key
type User struct {
	ID *int `gorm:"primaryKey; type:serial"`
	Username string `gorm:"type: varchar(32) not null unique"`
	Password []byte `gorm:"type: bytea not null"`
	Token string `gorm: type varchar(64)`
	TokenExpiration time.Time
	ReferredBy *int
	// Referrer *User `gorm:"foreignKey:ReferredBy;constraint:OnUpdate:CASCADE,ONDELETE:SET NULL"`
}

// UserPerm Model
type UserPerm struct {
	ID *int `gorm:"primaryKey; type:serial"`
	Name string `gorm:"type: varchar(64) not null unique"`
	Description string `gorm:"type: text"`
}

// Referrer Model. Where active user referrals reside
type Referrer struct {
	Code int `gorm:"primaryKey"`
	UserID *int
	User User `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Expiration time.Time `gorm:"not null"`
}

// Player Model
type Player struct {
	ID *int `gorm:"primaryKey; type:serial"`
	Name string `gorm:"type: varchar(64) not null unique"`
}

// ModsPerServer Model. Foriegn Key table
type ModsPerServer struct {
	ModID int `gorm:"not null; index:mods_per_server,unique"`
	Mod Mod `gorm:"constraint:OnUpdate:CASCADE,ONDELETE:CASCADE"`
	ServerID int `gorm:"not null; index:mods_per_server,unique"`
	Server Server `gorm:"constraint:OnUpdate:CASCADE,ONDELETE:CASCADE"`
}

// PermsPerUser Model. Foriegn Key table
type PermsPerUser struct {
	UserID int `gorm:"not null; index:perms_per_user,unique"`
	User User `gorm:"constraint:OnUpdate:CASCADE,ONDELETE:CASCADE"`
	UserPermID int `gorm:"not null; index:perms_per_user,unique"`
	UserPerm UserPerm `gorm:"constraint:OnUpdate:CASCADE,ONDELETE:CASCADE"`
}

// ServerPermsPerUser Model. Foriegn Key table
type ServerPermsPerUser struct {
	ServerID int `gorm:"not null; index:server_perms_per_user,unique"`
	Server Server `gorm:"constraint:OnUpdate:CASCADE,ONDELETE:CASCADE"`
	ServerPermID int `gorm:"not null; index:server_perms_per_user,unique"`
	ServerPerm ServerPerm `gorm:"constraint:OnUpdate:CASCADE,ONDELETE:CASCADE"`
	UserID int `gorm:"not null; index:server_perms_per_user,unique"`
	User User `gorm:"constraint:OnUpdate:CASCADE,ONDELETE:CASCADE"`
}

// UserPlayer Model. Foriegn Key table
type UserPlayer struct {
	UserID int `gorm:"not null; index:user_player,unique"`
	User User `gorm:"constraint:OnUpdate:CASCADE,ONDELETE:CASCADE"`
	PlayerID int `gorm:"not null; index:user_player,unique"`
	Player Player `gorm:"constraint:OnUpdate:CASCADE,ONDELETE:CASCADE"`
}

// ServerLog Model
type ServerLog struct {
	ID *int `gorm:"primaryKey; type:serial"`
	Time time.Time `gorm:"type: timestamp not null"`
	Command string `gorm:"type: text not null"`
	PlayerID *int
	Player Player `gorm:"constraint:OnUpdate:CASCADE,ONDELETE:SET NULL"`
	ServerID *int
	Server Server `gorm:"constraint:OnUpdate:CASCADE,ONDELETE:SET NULL"`
}

// PlayerLog Model
type PlayerLog struct {
	ID *int `gorm:"primaryKey; type: serial"`
	Time time.Time `gorm:"type: timestamp not null"`
	Action string `gorm:"type: text not null"`
	PlayerID *int
	Player Player `gorm:"constraint:OnUpdate:CASCADE,ONDELETE:SET NULL"`
	ServerID *int
	Server Server `gorm:"constraint:OnUpdate:CASCADE,ONDELETE:SET NULL"`
}

// WebLog Model
type WebLog struct {
	ID *int `gorm:"primaryKey; type:serial"`
	Time time.Time `gorm:"type: timestamp not null"`
	IP string `gorm:"type: varchar(128) not null"`
	Method string `gorm:"type: text not null"`
	StatusCode int `gorm:"not null"`
	QueryParams string `gorm:"type:json"`
	PostData string `gorm:"type:json"`
	Cookies string `gorm:"type:json"`
	UserID *int
	User User `gorm:"foreignKey:ID;constraint:OnUpdate:CASCADE,ONDELETE:SET NULL"`
}