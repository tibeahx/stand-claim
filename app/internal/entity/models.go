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

type SenderInfo struct {
	ID              int64
	Username        string
	GroupID         int64
	IsGroup         bool
	IsBot           bool
	IsInlineMode    bool
	IsFirstInstance bool
}

func SenderInfoFromContext(c telebot.Context) SenderInfo {
	chat := c.Chat()
	sender := c.Sender()

	info := SenderInfo{
		ID:           sender.ID,
		Username:     sender.Username,
		GroupID:      chat.ID,
		IsGroup:      chat.Type == telebot.ChatGroup || chat.Type == telebot.ChatSuperGroup,
		IsBot:        sender.IsBot,
		IsInlineMode: c.Message().Via != nil,
	}

	if info.IsGroup {
		info.IsFirstInstance = chat.ID == sender.ID
	}

	return info
}

type Stand struct {
	ID           uuid.UUID `db:"id"`
	Name         string    `db:"name"`
	Released     bool      `db:"released"`
	TimeClaimed  time.Time `db:"time_claimed"`
	TimeReleased time.Time `db:"time_released"`
	Owner        Owner
}
