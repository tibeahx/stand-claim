package handler

import (
	"github.com/tibeahx/claimer/app/internal/service"
	telegram "github.com/tibeahx/claimer/app/internal/telegram/bot"
	"gopkg.in/telebot.v3"
)

type Handler struct {
	bot     *telegram.Bot
	service *service.Service
}

func NewHandler(b *telegram.Bot, service *service.Service) *Handler {
	return &Handler{
		bot:     b,
		service: service,
	}
}

func (h *Handler) Handlers() map[string]telebot.HandlerFunc {
	return map[string]telebot.HandlerFunc{
		"/ping":      h.Ping,
		"/list":      h.ListStands,
		"/list_free": h.ListFreeStands,
		"/claim":     h.Claim,
		"/release":   h.Release,
	}
}

func (h *Handler) Ping(c telebot.Context) error {
	// owner := entity.OwnerFromContext(c)

	return nil
}

func (h *Handler) ListStands(c telebot.Context) error {
	// stands, err := h.service.ListStands(c)
	// if err != nil {
	// 	return fmt.Errorf("failed to list stands due to: %w", err)
	// }

	return nil
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
