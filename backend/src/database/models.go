package database

import (
	"time"
)

// Game Model
type Game struct {
	ID      *int   `gorm:"primaryKey; type:serial" json:"-"`
	Name    string `gorm:"type:varchar(64) not null unique" json:"name"`
	Image   string `gorm:"type:varchar(64) not null" json:"-"`
	IsImage bool   `gorm:"type:bool not null" json:"-"`
}

// Version Model
type Version struct {
	ID     *int   `gorm:"primaryKey; type:serial" json:"-"`
	Tag    string `gorm:"type:text not null; index:version,unique" json:"tag"`
	GameID *int   `gorm:"not null; index:version,unique" json:"-"`
	Game   Game   `gorm:"constraint:OnUpdate:CASCADE,ONDELETE:CASCADE" json:"game"`
}

// Mod Model
type Mod struct {
	ID        *int    `gorm:"primaryKey; type:serial" json:"-"`
	URL       string  `gorm:"type: text not null; index:mod,unique" json:"url"`
	Name      string  `gorm:"type: varchar(64) not null; index:mod,unique" json:"name"`
	GameID    *int    `gorm:"not null; index:mod,unique" json:"-"`
	Game      Game    `gorm:"constraint:OnUpdate:CASCADE,ONDELETE:CASCADE" json:"game"`
	VersionID *int    `gorm:"index:mod,unique" json:"-"`
	Version   Version `gorm:"constraint:OnUpdate:CASCADE,ONDELETE:SET NULL" json:"version"`
}

// DiscordIntegration Model
type DiscordIntegration struct {
	ID         *int    `gorm:"primaryKey; type:serial" json:"-"`
	Type       string  `gorm:"type: varchar(64)" json:"type"`
	DiscordURL string  `gorm:"type: text" json:"discord_url"`
	Username   *string `gorm:"type: varchar(64)" json:"username"`
	AvatarURL  *string `gorm:"type: text" json:"avatar_url"`
	Active     bool    `gorm:"not null; type: bool" json:"active"`
	ServerID   *int    `gorm:"index:server_integration,unique" json:"-"`
	Server     Server  `gorm:"constraint:OnUpdate:CASCADE,ONDELETE:SET NULL" json:"-"`
}

// Server Model
type Server struct {
	ID        *int    `gorm:"primaryKey; type:serial" json:"id"`
	Port      uint16  `gorm:"not null; unique; check: Port < 65536; check: Port > 0" json:"port"`
	Name      string  `gorm:"type: varchar(64)" json:"name"`
	Running   bool    `gorm:"not null; type: bool" json:"running"`
	GameID    *int    `gorm:"not null" json:"-"`
	Game      Game    `gorm:"constraint:OnUpdate:CASCADE,ONDELETE:CASCADE" json:"game"`
	OwnerID   *int    `gorm:"not null" json:"-"`
	Owner     User    `gorm:"constraint:OnUpdate:CASCADE,ONDELETE:CASCADE" json:"owner"`
	VersionID *int    `json:"-"`
	Version   Version `gorm:"constraint:OnUpdate:CASCADE,ONDELETE:SET NULL" json:"version"`
}

// ServerPerm Model
type ServerPerm struct {
	ID          *int   `gorm:"primaryKey; type:serial" json:"-"`
	Name        string `gorm:"type: varchar(64) not null unique" json:"name"`
	Description string `gorm:"type: text" json:"description"`
}

// User Model. ReferredBy is self referencing Foreign Key
type User struct {
	ID              *int      `gorm:"primaryKey; type:serial" json:"-"`
	Username        string    `gorm:"type: varchar(32) not null unique" json:"username"`
	Password        []byte    `gorm:"type: bytea not null" json:"-"`
	Token           string    `gorm:"type varchar(64) not null unique" json:"-"`
	TokenExpiration time.Time `json:"-"`
	ReferredBy      *int      `json:"-"`
	// Referrer *Owner `gorm:"foreignKey:ReferredBy;constraint:OnUpdate:CASCADE,ONDELETE:SET NULL"`
}

// UserPerm Model
type UserPerm struct {
	ID          *int   `gorm:"primaryKey; type:serial" json:"-"`
	Name        string `gorm:"type: varchar(64) not null unique" json:"name"`
	Description string `gorm:"type: text" json:"description"`
}

// Referrer Model. Where active user referrals reside
type Referrer struct {
	Code       int       `gorm:"primaryKey" json:"code"`
	Expiration time.Time `gorm:"not null" json:"expiration"`
	UserID     *int      `gorm:"not null" json:"-"`
	User       User      `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"user"`
}

// Player Model
type Player struct {
	ID   *int   `gorm:"primaryKey; type:serial" json:"-"`
	Name string `gorm:"type: varchar(64) not null unique" json:"name"`
}

// ModsPerServer Model. Foriegn Key table
type ModsPerServer struct {
	ModID     int     `gorm:"not null; index:mods_per_server,unique" json:"-"`
	Mod       Mod     `gorm:"constraint:OnUpdate:CASCADE,ONDELETE:CASCADE" json:"mod"`
	ServerID  int     `gorm:"not null; index:mods_per_server,unique" json:"-"`
	Server    Server  `gorm:"constraint:OnUpdate:CASCADE,ONDELETE:CASCADE" json:"server"`
	VersionID int     `gorm:"not null; index:mods_per_server,unique" json:"-"`
	Version   Version `gorm:"constraint:OnUpdate:CASCADE,ONDELETE:CASCADE" json:"version"`
}

// PermsPerUser Model. Foriegn Key table
type PermsPerUser struct {
	UserID     int      `gorm:"not null; index:perms_per_user,unique" json:"user_id"`
	User       User     `gorm:"constraint:OnUpdate:CASCADE,ONDELETE:CASCADE" json:"user"`
	UserPermID int      `gorm:"not null; index:perms_per_user,unique" json:"user_perm_id"`
	UserPerm   UserPerm `gorm:"constraint:OnUpdate:CASCADE,ONDELETE:CASCADE" json:"user_perm"`
}

// ServerPermsPerUser Model. Foriegn Key table
type ServerPermsPerUser struct {
	ServerID     int        `gorm:"not null; index:server_perms_per_user,unique" json:"server_id"`
	Server       Server     `gorm:"constraint:OnUpdate:CASCADE,ONDELETE:CASCADE" json:"server"`
	ServerPermID int        `gorm:"not null; index:server_perms_per_user, unique" json:"server_perm_id"`
	ServerPerm   ServerPerm `gorm:"constraint:OnUpdate:CASCADE,ONDELETE:CASCADE" json:"server_perm"`
	UserID       int        `gorm:"not null; index:server_perms_per_user,unique" json:"user_id"`
	User         User       `gorm:"constraint:OnUpdate:CASCADE,ONDELETE:CASCADE" json:"user"`
}

// UserPlayer Model. Foriegn Key table
type UserPlayer struct {
	UserID   int    `gorm:"not null; index:user_player,unique" json:"-"`
	User     User   `gorm:"constraint:OnUpdate:CASCADE,ONDELETE:CASCADE" json:"user"`
	PlayerID int    `gorm:"not null; index:user_player,unique" json:"-"`
	Player   Player `gorm:"constraint:OnUpdate:CASCADE,ONDELETE:CASCADE" json:"player"`
}

// ServerLog Model
type ServerLog struct {
	ID       *int      `gorm:"primaryKey; type:serial" json:"id"`
	Time     time.Time `gorm:"type: timestamp not null" json:"time"`
	Command  string    `gorm:"type: text not null" json:"command"`
	PlayerID *int      ` gorm:"not null" json:"player_id"`
	Player   Player    `gorm:"constraint:OnUpdate:CASCADE,ONDELETE:CASCADE" json:"player"`
	ServerID *int      `gorm:"not null" json:"server_id"`
	Server   Server    `gorm:"constraint:OnUpdate:CASCADE,ONDELETE:CASCADE" json:"server"`
}

// PlayerLog Model
type PlayerLog struct {
	ID       *int      `gorm:"primaryKey; type: serial" json:"id"`
	Time     time.Time `gorm:"type: timestamp not null" json:"time"`
	Action   string    `gorm:"type: text not null" json:"action"`
	PlayerID *int      ` gorm:"not null" json:"-" json:"player_id"`
	Player   Player    `gorm:"constraint:OnUpdate:CASCADE,ONDELETE:CASCADE" json:"player"`
}

// WebLog Model
type WebLog struct {
	ID          *int      `gorm:"primaryKey; type:serial" json:"id"`
	Time        time.Time `gorm:"type: timestamp not null" json:"time"`
	IP          string    `gorm:"type: varchar(128) not null" json:"ip"`
	Method      string    `gorm:"type: text not null" json:"method"`
	StatusCode  int       `gorm:"not null" json:"status_code"`
	QueryParams string    `gorm:"type:json" json:"query_params"`
	PostData    string    `gorm:"type:json" json:"post_data"`
	Cookies     string    `gorm:"type:json" json:"cookies"`
	UserID      *int      ` json:"user_id"`
	User        User      `gorm:"foreignKey:ID;constraint:OnUpdate:CASCADE,ONDELETE:SET NULL" json:"user"`
}
