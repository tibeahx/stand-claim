package telegram

import (
	"fmt"
	"time"

	"github.com/tibeahx/claimer/app/internal/config"

	"gopkg.in/telebot.v4"
)

type Bot struct {
	tele *telebot.Bot
}

func NewBot(cfg *config.Config) (*Bot, error) {
	b, err := telebot.NewBot(telebot.Settings{
		Verbose: cfg.Bot.Verbose,
		Token:   cfg.Bot.Token,
		Poller: &telebot.LongPoller{
			Timeout: 10 * time.Second,
			AllowedUpdates: []string{
				"message",
				"edited_message",
				"inline_query",
				"callback_query",
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to build bot: %w", err)
	}

	b.Use(Middleware)

	return &Bot{tele: b}, nil
}

func (b *Bot) Tele() *telebot.Bot {
	return b.tele
}
