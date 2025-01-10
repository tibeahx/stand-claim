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
			msg := c.Message()
			if msg != nil && msg.UserJoined != nil {
				username := msg.UserJoined.Username
				if username != "" {
					userFound, err := h.repo.FindUser(username)
					if err != nil {
						log.Zap().Errorf("failed to find user %s: %v", username, err)
						return err
					}
					if !userFound {
						if err := h.repo.CreateUser(username); err != nil {
							log.Zap().Errorf("failed to create user %s: %v", username, err)
							return err
						}
						log.Zap().Infof("user %s created", username)
					}
				}
			}

			if msg != nil && msg.UserLeft != nil {
				username := msg.UserLeft.Username
				if username != "" {
					if err := h.repo.DeleteUser(username); err != nil {
						log.Zap().Errorf("failed to delete user %s: %v", username, err)
						return err
					}
					log.Zap().Infof("user %s deleted", username)
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
