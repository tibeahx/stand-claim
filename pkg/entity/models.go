package entity

import (
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
	Name          string    `db:"name"`
	Released      bool      `db:"released,omitempty"`
	OwnerUsername string    `db:"owner_username,omitempty"`
	TimeClaimed   time.Time `db:"time_claimed,omitempty"`
}
