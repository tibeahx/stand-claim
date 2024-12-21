package middleware

import "gopkg.in/telebot.v3"

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
