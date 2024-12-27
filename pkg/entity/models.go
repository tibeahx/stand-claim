package entity

import (
	"time"

	"github.com/google/uuid"
	"gopkg.in/telebot.v4"
)

type ChatInfo struct {
	ChatID  int64
	IsGroup bool
	Members []telebot.ChatMember
}

type Owner struct {
	ID       int64  `db:"owner_id"`
	Username string `db:"owner_username"`
}

func NewOwner(c telebot.Context) Owner {
	return Owner{
		ID:       c.Message().Sender.ID,
		Username: c.Message().Sender.Username,
	}
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
