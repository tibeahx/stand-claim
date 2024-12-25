package telegram

import (
	"fmt"

	"github.com/tibeahx/claimer/app/internal/repo"
	"github.com/tibeahx/claimer/pkg/utils"

	"gopkg.in/telebot.v4"
)

type Handler struct {
	repo *repo.Repo
}

func NewHandler(b *Bot, repo *repo.Repo) *Handler {
	return &Handler{
		repo: repo,
	}
}

func (h *Handler) Ping(c telebot.Context) error {
	// owner := entity.OwnerFromContext(c)
	// c.Chat().
	return nil
}

func (h *Handler) ListStands(c telebot.Context) error {
	stands, err := h.repo.Stands()
	if err != nil {
		return err
	}

	if len(stands) == 0 {
		return c.Send("No environments found")
	}

	var response string

	for _, stand := range stands {
		if stand.OwnerUsername == "" {
			continue
		}

		status := utils.FormatStandStatus(stand)

		response += fmt.Sprintf(
			"%s %s %s\n",
			utils.Computer,
			stand.Name,
			status,
		)
	}

	return c.Reply(response)
}

func (h *Handler) Claim(c telebot.Context) error {
	return nil
}

func (h *Handler) Release(c telebot.Context) error {
	return nil
}

func (h *Handler) Greetings(c telebot.Context) error {
	greeting := fmt.Sprintf(
		"Hello %s, I'm a StandClaimer bot, I will help you to manage enviroments accross the team. Tap `/` on the group menu to see commands",
		c.Sender().Username,
	)
	return c.Send(greeting)
}

func (h *Handler) Handlers() map[string]telebot.HandlerFunc {
	return map[string]telebot.HandlerFunc{
		"/list":    h.ListStands,
		"/claim":   h.Claim,
		"/release": h.Release,
		"/ping":    h.Ping,
	}
}
