package middleware

import (
	"gopkg.in/telebot.v4"
)

func Middleware(handler telebot.HandlerFunc) telebot.HandlerFunc {
	return invalidCommandMiddleware(handler)
}

func invalidCommandMiddleware(next telebot.HandlerFunc) telebot.HandlerFunc {
	return func(c telebot.Context) error {
		if !isValidCommand(c.Text()) {
			return c.Send("unknown command, see `/commands` for available commands")
		}
		return next(c)
	}
}

func authMiddleware(next telebot.HandlerFunc) telebot.HandlerFunc {
	return func(ctx telebot.Context) error {
		sender := ctx.Sender()
		allowedUsers := ctx.Get("allowed_users").([]int)

		for _, usr := range allowedUsers {
			if usr > 0 {
				if sender.ID == int64(usr) {
					return next(ctx)
				}
			}
		}
		return ctx.Send("not allowed")
	}
}

func isValidCommand(cmd string) bool {
	for _, validCmd := range []string{
		"/ping",
		"/claim",
		"/list",
		"/list_free",
		"/release",
	} {
		if cmd == validCmd {
			return true
		}
	}
	return false
}
