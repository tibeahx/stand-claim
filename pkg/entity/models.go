package entity

import (
	"database/sql"
	"time"
)

type ChatInfo struct {
	ChatID int64
}

type User struct {
	Username string    `db:"username"`
	Created  time.Time `db:"created"`
}

type Stand struct {
	Name          string         `db:"name"`
	Released      bool           `db:"released,omitempty"`
	OwnerUsername sql.NullString `db:"owner_username"`
	TimeClaimed   sql.NullTime   `db:"time_claimed"`
}
