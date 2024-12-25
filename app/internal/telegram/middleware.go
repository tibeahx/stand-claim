package telegram

import (
	"strings"

	"github.com/tibeahx/claimer/app/internal/config"
	"github.com/tibeahx/claimer/pkg/log"
	"gopkg.in/telebot.v4"
)

func Middleware(handler telebot.HandlerFunc) telebot.HandlerFunc {
	return validateCmdMiddleware(handler)
}

func validateCmdMiddleware(next telebot.HandlerFunc) telebot.HandlerFunc {
	return func(c telebot.Context) error {
		if strings.HasPrefix(c.Message().Text, "/") {
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
		if command.Text != cmd {
			return false
		}
	}
	return true
}
