package telegram

import (
	"fmt"

	"github.com/tibeahx/claimer/app/internal/config"
	"github.com/tibeahx/claimer/app/internal/repo"

	"gopkg.in/telebot.v4"
)

type Handler struct {
	bot  *Bot
	repo *repo.Repo
}

func NewHandler(b *Bot, repo *repo.Repo) *Handler {
	return &Handler{
		bot:  b,
		repo: repo,
	}
}

func (h *Handler) Ping(c telebot.Context) error {
	// owner := entity.OwnerFromContext(c)
	// c.Chat().
	return nil
}

func (h *Handler) ListStands(c telebot.Context) error {
	stands, err := h.repo.Stands(c)
	if err != nil {
		return err

	}
	return c.Send(stands)
}

func (h *Handler) ListFreeStands(c telebot.Context) error {
	return nil
}

func (h *Handler) Claim(c telebot.Context) error {
	return nil
}

func (h *Handler) Release(c telebot.Context) error {
	return nil
}

func (h *Handler) Status(c telebot.Context) error {
	return nil
}

func (h *Handler) Greetings(c telebot.Context) error {
	greeting := fmt.Sprintf(
		"Hello %s, I'm a StandClaimer bot, I will help you to track testing stands for deployments accross the team. You may refer my by query @StandClaimBot",
		c.Sender().Username,
	)
	return c.Send(greeting)
}

func (h *Handler) Handlers(cfg *config.Config) map[string]telebot.HandlerFunc {
	funcMap := make(map[string]telebot.HandlerFunc)

	handlers := []telebot.HandlerFunc{
		h.Ping,
		h.ListStands,
		h.ListFreeStands,
		h.Claim,
		h.Release,
		h.Status,
	}

	for i, cmd := range config.Commands {
		if _, ok := funcMap[cmd.Text]; !ok {
			funcMap[cmd.Text] = handlers[i]
		}
	}

	return funcMap
}
