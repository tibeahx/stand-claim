package telegram

import (
	"strings"

	"github.com/tibeahx/claimer/app/internal/config"
	"gopkg.in/telebot.v4"
)

func Middleware(handler telebot.HandlerFunc) telebot.HandlerFunc {
	return validateCmdMiddleware(handler)
}

func validateCmdMiddleware(next telebot.HandlerFunc) telebot.HandlerFunc {
	return func(c telebot.Context) error {
		if strings.HasPrefix(c.Text(), "/") {
			if !isValidCommand(c.Text()) {
				return c.Send("unknown command, see `/` for available commands")
			}
		}
		return next(c)
	}
}

func isValidCommand(cmd string) bool {
	for _, command := range config.Commands {
		if command.Text != cmd {
			return false
		}
	}
	return true
}
