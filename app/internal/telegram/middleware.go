package telegram

import (
	"strings"

	"github.com/tibeahx/claimer/app/internal/config"
	"github.com/tibeahx/claimer/pkg/entity"
	"github.com/tibeahx/claimer/pkg/log"
	"gopkg.in/telebot.v4"
)

var ChatInfo = entity.ChatInfo{}

func ChatInfoMiddleware(next telebot.HandlerFunc) telebot.HandlerFunc {
	return func(c telebot.Context) error {
		if c.Chat().Type == telebot.ChatGroup || c.Chat().Type == telebot.ChatSuperGroup {
			ChatInfo.ChatID = c.Chat().ID
		}
		log.Zap().Infof("chat id set to: %d", ChatInfo.ChatID)
		return next(c)
	}
}

func UserMiddleware(h *Handler) telebot.MiddlewareFunc {
	return func(next telebot.HandlerFunc) telebot.HandlerFunc {
		return func(c telebot.Context) error {
			u := c.ChatMember().NewChatMember.User.Username
			userFound, err := h.repo.FindUser(u)
			if err != nil {
				return err
			}
			if !userFound {
				h.repo.CreateUser(u)
				return next(c)
			}

			oldMember := c.ChatMember().OldChatMember
			u2 := oldMember.User.Username

			if oldMember.Role != telebot.Administrator {
				switch oldMember.Role {
				case telebot.Left, telebot.Kicked, telebot.Restricted:
					userFound, err := h.repo.FindUser(u2)
					if err != nil {
						return err
					}
					if userFound {
						h.repo.DeleteUser(u2)
					}
				}
			}
			return next(c)
		}
	}
}

func ValidateCmdMiddleware(next telebot.HandlerFunc) telebot.HandlerFunc {
	return func(c telebot.Context) error {
		if c.Callback() != nil {
			return next(c)
		}

		if c.Message() != nil && strings.HasPrefix(c.Message().Text, "/") {
			cmdText := strings.Split(c.Text(), "@")[0]
			log.Zap().Infof("got command %s", cmdText)
			if !isValidCommand(cmdText) {
				return c.Send("unknown command, see `/` for available commands")
			}
		}

		return next(c)
	}
}

func isValidCommand(cmd string) bool {
	for _, command := range config.TeleCommands {
		if command.Text == cmd {
			return true
		}
	}
	return false
}
