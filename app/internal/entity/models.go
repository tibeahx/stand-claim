package entity

import (
	"time"

	"github.com/google/uuid"
	"gopkg.in/telebot.v4"
)

type Owner struct {
	ID       int64  `db:"owner_id"`
	Username string `db:"owner_username"`
	GroupID  int64  `db:"owner_group_id"`
}

type UserInfo struct {
	ID       int64
	Username string
	GroupID  int64
	IsGroup  bool
}

func UserInfoFromContext(c telebot.Context) *UserInfo {
	info := UserInfo{
		ID:       c.Sender().ID,
		Username: c.Sender().Username,
		GroupID:  c.Chat().ID,
		IsGroup:  c.Chat().Type == telebot.ChatGroup || c.Chat().Type == telebot.ChatSuperGroup,
	}

	if info.IsGroup {
		c.Chat().ID = c.Sender().ID
	}

	return &info
}

type Stand struct {
	ID           uuid.UUID `db:"id"`
	Name         string    `db:"name"`
	Released     bool      `db:"released"`
	TimeClaimed  time.Time `db:"time_claimed"`
	TimeReleased time.Time `db:"time_released"`
	Owner        Owner
}
