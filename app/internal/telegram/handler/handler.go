package handler

import (
	"github.com/tibeahx/claimer/app/internal/service"
	telegram "github.com/tibeahx/claimer/app/internal/telegram/bot"
	"gopkg.in/telebot.v4"
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
	// c.Chat().
	return nil
}

func (h *Handler) ListStands(c telebot.Context) error {
	stands, err := h.service.ListStands(c)
	if err != nil {
		return err

	}
	return reply(c, stands)
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

func reply(c telebot.Context, content any) error {
	if c.Chat().Type == telebot.ChatChannel || c.Chat().Type == telebot.ChatChannelPrivate {
		return c.Send(content, telebot.SendOptions{
			ReplyTo:   c.Message(),
			Protected: true,
		})
	}
	return c.Send(content, telebot.SendOptions{Protected: true})
}
