package entity

import (
	"time"

	"gopkg.in/telebot.v4"
)

type ChatInfo struct {
	ChatID  int64
	IsGroup bool
	Members []telebot.ChatMember
}

type User struct {
	Username string    `db:"username"`
	Created  time.Time `db:"created"`
}

type Stand struct {
	Name          string    `db:"name"`
	Released      bool      `db:"released"`
	OwnerUsername string    `db:"owner_username"`
	TimeClaimed   time.Time `db:"time_claimed"`
}
