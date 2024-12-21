package telegram

import (
	"fmt"

	middleware "github.com/tibeahx/claimer/app/internal/transport"
	"github.com/tibeahx/claimer/pkg/opts"
	"gopkg.in/telebot.v4"
)

type BotOptions struct {
	Verbose    bool
	GroupID    int64
}

type Bot struct {
	tele       *telebot.Bot
	middleware telebot.MiddlewareFunc
}

func NewBot(token string, opts BotOptions) (*Bot, error) {
	b, err := telebot.NewBot(telebot.Settings{
		Verbose: opts.Verbose,
		Token:   token,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to build bot: %w", err)
	}

	bot := &Bot{
		tele:       b,
		middleware: middleware.Middleware,
	}

	b.Use(bot.middleware)

	return bot, nil
}

func (b *Bot) Middleware() telebot.MiddlewareFunc {
	return b.middleware
}

func (b *Bot) Tele() *telebot.Bot {
	return b.tele
}

func (b *Bot) SetCommands(opts ...opts.Options) error {
	return b.tele.SetCommands(opts)
}
