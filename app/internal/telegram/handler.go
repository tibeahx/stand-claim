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
		if len(users) == 0 {
			return nil
		}

		mentionsFormatted := make([]string, 0, len(users))
		for _, user := range users {
			mentionsFormatted = append(mentionsFormatted, "@"+user)
		}

		message := fmt.Sprintf(
			"%s, would you mind to release the stand? It's been busy for more than 100 hours",
			strings.Join(mentionsFormatted, ", "),
		)

		_, err := h.bot.Tele().Send(&telebot.Chat{ID: chatID}, message)
		return err
	}
}

func (h *Handler) PingAll(c telebot.Context) error {
	stands, err := h.repo.Stands()
	if err != nil {
		return err
	}

	if len(stands) == 0 {
		return c.Reply("No environments found")
	}

	var (
		mentions = make(map[string]string, 0)
		parts    = make([]string, 0)
	)

	for _, stand := range stands {
		if stand.Released {
			continue
		}
		if stand.OwnerUsername != "" && stand.Name != "" {
			mentions[stand.OwnerUsername] = stand.Name
			parts = append(parts, fmt.Sprintf("@%s: %s", stand.OwnerUsername, stand.Name))
		}
	}

	if len(mentions) == 0 {
		return c.Reply("No busy stands found")
	}

	message := fmt.Sprintf("%s, would you mind releasing your stands?", strings.Join(parts, ", "))

	return c.Send(message)
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
		return c.Send(fmt.Sprintf(
			"@%s would you mind to release the stand?",
			c.Message().Payload),
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

	standInfos := make([]string, 0)

	for _, stand := range stands {
		if stand.OwnerUsername == "" || stand.Name == "" {
			continue
		}

		standInfo := fmt.Sprintf("%s %s %s",
			comp,
			stand.Name,
			formatStandStatus(stand),
		)

		standInfos = append(standInfos, standInfo)
	}

	if len(standInfos) == 0 {
		return c.Reply("No stands found")
	}

	message := strings.Join(standInfos, "\n")

	return c.Reply(message)
}

// для клейма должна всплывать менюшка с доступными стендами
func (h *Handler) Claim(c telebot.Context) error {
	stands, err := h.repo.Stands()
	if err != nil {
		return err
	}

	if len(stands) == 0 {
		return c.Reply("No environments found")
	}

	var (
		standName      = c.Message().Payload
		senderUsername = c.Message().Sender.Username
	)

	for _, stand := range stands {
		if stand.Name == standName {
			if stand.Released {
				h.repo.ClaimStand(stand)
				return c.Send(fmt.Sprintf(
					"@%s has claimed %s",
					senderUsername,
					standName),
				)
			}

			return c.Reply("stand is busy, choose another free one")
		}
	}

	return c.Reply("stand not found")
}

// для релиза должна всплывать менюшка с доступными стендами
func (h *Handler) Release(c telebot.Context) error {
	stands, err := h.repo.Stands()
	if err != nil {
		return err
	}

	if len(stands) == 0 {
		return c.Reply("No environments found")
	}

	var (
		standName      = c.Message().Payload
		senderUsername = c.Message().Sender.Username
	)

	for _, stand := range stands {
		if stand.Name == standName && stand.OwnerUsername == senderUsername {
			if !stand.Released {
				h.repo.ReleaseStand(stand)
				return c.Send(fmt.Sprintf(
					"@%s has released %s",
					senderUsername,
					standName),
				)
			}

			return c.Reply("stand is already free")
		}
	}

	return c.Reply("stand not found")
}

func (h *Handler) Greetings(c telebot.Context) error {
	return c.Send(fmt.Sprintf(
		"Hello @%s, I'm StandClaimer bot, I will help you to manage environments across the team. "+
			"Tap `/` on the group menu to see commands",
		c.Message().UserJoined.Username,
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
			"busy by @%v for %d h. %s",
			stand.OwnerUsername,
			int(timeBusy.Hours()),
			busy,
		)
	}

	return fmt.Sprintf("is finally free %s", free)
}
