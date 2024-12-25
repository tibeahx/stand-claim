package entity

import (
	"time"

	"github.com/google/uuid"
)

type Owner struct {
	ID       int64  `db:"owner_id"`
	Username string `db:"owner_username"`
}

type Stand struct {
	ID            uuid.UUID `db:"id"`
	Name          string    `db:"name"`
	Released      bool      `db:"released"`
	TimeClaimed   time.Time `db:"time_claimed"`
	TimeReleased  time.Time `db:"time_released"`
	OwnerID       int64     `db:"owner_id"`
	OwnerUsername string    `db:"owner_username"`
}
