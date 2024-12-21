package entity

import (
	"github.com/google/uuid"
	"gopkg.in/telebot.v3"
)

type Owner struct {
	ID       int64  `db:"owner_id"`
	Username string `db:"owner_username"`
	GroupID  int64  `db:"owner_group_id"`
}

func OwnerFromContext(c telebot.Context) Owner {
	return Owner{
		ID:       c.Sender().ID,
		Username: c.Sender().Username,
		GroupID:  c.Message().Chat.ID,
	}
}

type StandOptions struct {
	Claimed bool `db:"claimed"`
	Owner   Owner
}

type Stand struct {
	ID   uuid.UUID `db:"id"`
	Name string    `db:"name"`
	Opts StandOptions
}
