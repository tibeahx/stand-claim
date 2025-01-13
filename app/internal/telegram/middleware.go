package telegram

import (
	"errors"
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

var (
	errNoUsersJoined = errors.New("no users joined within event")
	errNoUsersLeft   = errors.New("no users left within event")
)

func UserJoinedMiddleware(h *Handler) telebot.MiddlewareFunc {
	return func(next telebot.HandlerFunc) telebot.HandlerFunc {
		return func(c telebot.Context) error {
			msg := c.Message()

			if msg.UserJoined == nil {
				return errNoUsersJoined
			}

			username := msg.UserJoined.Username

			userFound, err := h.repo.FindUser(username)
			if err != nil {
				return err
			}

			if !userFound {
				if err := h.repo.CreateUser(username); err != nil {
					return err
				}
			} else {
				log.Zap().Infof("user %s already exists", username)
			}

			return next(c)
		}
	}
}

func UserLeftMiddleware(h *Handler) telebot.MiddlewareFunc {
	return func(next telebot.HandlerFunc) telebot.HandlerFunc {
		return func(c telebot.Context) error {
			msg := c.Message()

			if msg.UserLeft == nil {
				return errNoUsersLeft
			}

			username := msg.UserLeft.Username

			userFound, err := h.repo.FindUser(username)
			if err != nil {
				return err
			}

			if userFound {
				if err := h.repo.DeleteUser(username); err != nil {
					return err
				}
			} else {
				log.Zap().Info("user not found to be deleted")
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
