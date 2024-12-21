package telegram

import (
	"fmt"

	"gopkg.in/telebot.v3"
)

type BotOptions struct {
	ErrHandler func(error, telebot.Context)
	Verbose    bool
	GroupID    int64
}

type Bot struct {
	tele       *telebot.Bot
	errHandler func(error, telebot.Context)
}

func NewBot(token string, opts BotOptions) (*Bot, error) {
	b, err := telebot.NewBot(telebot.Settings{
		Verbose: opts.Verbose,
		Token:   token,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to build bot: %w", err)
	}
	return &Bot{tele: b}, nil
}

func (b *Bot) ErrHandler() func(error, telebot.Context) {
	return b.errHandler
}

func (b *Bot) Tele() *telebot.Bot {
	return b.tele
}