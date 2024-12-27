package telegram

import (
	"fmt"
	"strings"
	"time"

	"github.com/tibeahx/claimer/app/internal/repo"
	"github.com/tibeahx/claimer/pkg/entity"
	"gopkg.in/telebot.v4"
)

type notifierFunc func(chatID int64, users ...string) error

const (
	free = "✅"
	busy = "❌"
	comp = "🖥️"
)

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
		mentions := make([]string, len(users))
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

	mentions := make([]string, 0)

	for _, stand := range stands {
		if !stand.Released {
			continue
		}
		if stand.OwnerUsername == "" || stand.Name == "" {
			mentions = append(mentions, stand.OwnerUsername)
		}
	}

	return c.Send(fmt.Sprintf(
		"@%s would you mind to release the stand?",
		strings.Join(mentions, ", "),
	))
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

	if _, found := usersToPing[c.Message().Payload]; found {
		return c.Send(
			fmt.Sprintf(
				"@%s would you mind to release the stand?",
				c.Message().Payload,
			),
		)
	}

	return nil
}

func (h *Handler) ListStands(c telebot.Context) error {
	stands, err := h.repo.Stands()
	if err != nil {
		return err
	}

	if len(stands) == 0 {
		return c.Reply("No environments found")
	}

	for _, stand := range stands {
		if stand.OwnerUsername == "" {
			continue
		}

		return c.Reply(fmt.Sprintf(
			"%s %s %s\n",
			comp,
			stand.Name,
			formatStandStatus(stand),
		))
	}

	return nil
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
	stands, err := h.repo.Stands()
	if err != nil {
		return err
	}

	standName := c.Message().Payload

	for _, stand := range stands {
		if stand.Name == standName {
			if !stand.Released {
				h.repo.ReleaseStand(stand, entity.NewOwner(c))
				return c.Send(
					fmt.Sprintf(
						"@%s has released %s",
						c.Message().Sender.Username,
						standName,
					),
				)
			}

			return c.Reply("stand is already released, choose another busy one")
		}
	}
	return c.Reply("stand not found")
}

func (h *Handler) Greetings(c telebot.Context) error {
	joined := c.Message().UserJoined

	return c.Send(fmt.Sprintf(
		"Hello @%s, I'm StandClaimer bot, I will help you to manage environments across the team. "+
			"Tap `/` on the group menu to see commands",
		joined.Username,
	))
}

func (h *Handler) Handlers() map[string]telebot.HandlerFunc {
	return map[string]telebot.HandlerFunc{
		"/list":     h.ListStands,
		"/claim":    h.Claim,
		"/release":  h.Release,
		"/ping":     h.Ping,
		"/ping_all": h.PingAll,
	}
}

func (h *Handler) Bot() *Bot {
	return h.bot
}

func (h *Handler) Repo() *repo.Repo {
	return h.repo
}

func formatStandStatus(stand entity.Stand) string {
	if !stand.Released {
		timeBusy := time.Since(stand.TimeClaimed)

		return fmt.Sprintf(
			"busy by @%s for %d h. %s",
			stand.OwnerUsername,
			int(timeBusy.Hours()),
			busy,
		)
	}

	return fmt.Sprintf("is finally free %s", free)
}
