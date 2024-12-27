package telegram

import (
	"fmt"
	"strings"

	"github.com/tibeahx/claimer/app/internal/repo"
	"github.com/tibeahx/claimer/pkg/entity"
	"github.com/tibeahx/claimer/pkg/utils"

	"gopkg.in/telebot.v4"
)

type notifierFunc func(chatID int64, users ...string) error
type Handler struct {
	repo *repo.Repo
	bot  *Bot
}

func NewHandler(b *Bot, repo *repo.Repo) *Handler {
	return &Handler{
		repo: repo,
		bot:  b,
	}
}

func (h *Handler) Notify(chatID int64) notifierFunc {
	return func(chatID int64, users ...string) error {
		mentions := make([]string, 0, len(users))
		mentions = append(mentions, users...)

		response := fmt.Sprintf(
			"%s would you mind to release the stand? It's been busy for more than 100 hours",
			strings.Join(mentions, ", "),
		)
		_, err := h.bot.Tele().Send(&telebot.Chat{ID: chatID}, response)

		return err
	}
}

func (h *Handler) PingAll(c telebot.Context) error {
	stands, err := h.repo.Stands()
	if err != nil {
		return err
	}

	usersToPing := make(map[string]struct{})

	var response string

	for _, stand := range stands {
		if stand.OwnerUsername == "" || stand.Name == "" {
			continue
		}
		if !stand.Released {
			for user := range usersToPing {
				response += fmt.Sprintf("@%s would you mind to release the stand?", user)
			}
		}
	}

	return c.Send(response)
}

func (h *Handler) Ping(c telebot.Context) error {
	stands, err := h.repo.Stands()
	if err != nil {
		return err
	}

	if len(stands) == 0 {
		return c.Reply("No environments found")
	}

	usersToPing := make(map[string]struct{})

	for _, stand := range stands {
		if stand.OwnerUsername == "" || stand.Name == "" {
			continue
		}
		if !stand.Released {
			usersToPing[stand.OwnerUsername] = struct{}{}
		}
	}

	var response string

	usernameToPingFromMessage := c.Message().Payload

	if _, found := usersToPing[usernameToPingFromMessage]; found {
		response = fmt.Sprintf("@%s would you mind to release the stand?", usernameToPingFromMessage)
	}

	return c.Send(response)
}

func (h *Handler) ListStands(c telebot.Context) error {
	stands, err := h.repo.Stands()
	if err != nil {
		return err
	}

	if len(stands) == 0 {
		return c.Reply("No environments found")
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
	stands, err := h.repo.Stands()
	if err != nil {
		return err
	}

	standName := c.Message().Payload

	for _, stand := range stands {
		if stand.Name == standName {
			if stand.Released {
				h.repo.ClaimStand(stand, entity.NewOwner(c))
				return c.Send(
					fmt.Sprintf(
						"@%s has claimed %s",
						c.Message().Sender.Username,
						standName,
					),
				)
			}

			return c.Reply("stand is busy, choose another free one")
		}
	}

	return c.Reply("stand not found")
}

func (h *Handler) Release(c telebot.Context) error {
	return nil
}

func (h *Handler) Greetings(c telebot.Context) error {
	joined := c.Message().UserJoined

	greeting := fmt.Sprintf(
		"Hello @%s, I'm a StandClaimer bot, I will help you to manage environments across the team. "+
			"Tap `/` on the group menu to see commands",
		joined.Username,
	)

	return c.Send(greeting)
}

func (h *Handler) Bot() *Bot {
	return h.bot
}

func (h *Handler) Repo() *repo.Repo {
	return h.repo
}

func (h *Handler) Handlers() map[string]telebot.HandlerFunc {
	return map[string]telebot.HandlerFunc{
		"/list":    h.ListStands,
		"/claim":   h.Claim,
		"/release": h.Release,
		"/ping":    h.Ping,
	}
}
