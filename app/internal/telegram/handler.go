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

	if c.Callback() != nil {
		username := c.Message().Payload

		for _, stand := range stands {
			if stand.OwnerUsername == username && !stand.Released {
				return c.Edit(fmt.Sprintf(
					"@%s would you mind to release the stand?",
					username),
				)
			}
		}
		return c.Edit("User has no busy stands")
	}

	var (
		menu        = make([][]telebot.InlineButton, 0)
		row         = make([]telebot.InlineButton, 0)
		usersToPing = make(map[string]struct{})
	)

	for _, stand := range stands {
		if stand.Released || stand.OwnerUsername == "" {
			continue
		}

		if _, exists := usersToPing[stand.OwnerUsername]; !exists {
			usersToPing[stand.OwnerUsername] = struct{}{}

			btn := telebot.InlineButton{
				Text: fmt.Sprintf("@%s (%s)", stand.OwnerUsername, stand.Name),
				Data: fmt.Sprintf("ping:%s", stand.OwnerUsername),
			}
			row = append(row, btn)

			if len(row) == 2 {
				menu = append(menu, row)
				row = []telebot.InlineButton{}
			}
		}
	}

	if len(row) > 0 {
		menu = append(menu, row)
	}

	if len(menu) == 0 {
		return c.Reply("No busy stands found")
	}

	return c.Reply("Choose user to ping:", &telebot.ReplyMarkup{
		InlineKeyboard: menu,
	})
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

func (h *Handler) Claim(c telebot.Context) error {
	stands, err := h.repo.Stands()
	if err != nil {
		return err
	}

	if len(stands) == 0 {
		return c.Reply("No environments found")
	}

	if c.Callback() != nil {
		standName := c.Message().Payload
		senderUsername := c.Callback().Sender.Username

		for _, stand := range stands {
			if stand.Name == standName {
				if stand.Released {
					standToClaim := entity.Stand{
						Name:          standName,
						OwnerUsername: senderUsername,
					}

					if err := h.repo.ClaimStand(standToClaim); err != nil {
						return c.Edit(fmt.Sprintf("Failed to claim stand: %v", err))
					}

					return c.Edit(fmt.Sprintf(
						"@%s has claimed %s",
						senderUsername,
						standName),
					)
				}
				return c.Edit("stand is busy, choose another free one")
			}
		}
		return c.Edit("stand not found")
	}

	var (
		menu = make([][]telebot.InlineButton, 0)
		row  = make([]telebot.InlineButton, 0)
	)

	for _, stand := range stands {
		if !stand.Released || stand.Name == "" {
			continue
		}

		btn := telebot.InlineButton{
			Text: fmt.Sprintf("%s %s", comp, stand.Name),
			Data: fmt.Sprintf("claim:%s", stand.Name),
		}
		row = append(row, btn)

		if len(row) == 2 {
			menu = append(menu, row)
			row = []telebot.InlineButton{}
		}
	}

	if len(row) > 0 {
		menu = append(menu, row)
	}

	if len(menu) == 0 {
		return c.Reply("No free stands available")
	}

	return c.Reply("Choose stand to claim:", &telebot.ReplyMarkup{
		InlineKeyboard: menu,
	})
}

func (h *Handler) Release(c telebot.Context) error {
	stands, err := h.repo.Stands()
	if err != nil {
		return err
	}

	if len(stands) == 0 {
		return c.Reply("No environments found")
	}

	if c.Callback() != nil {
		standName := c.Message().Payload
		senderUsername := c.Callback().Sender.Username

		standToRelease := entity.Stand{
			Name:          standName,
			OwnerUsername: senderUsername,
		}

		if err := h.repo.ReleaseStand(standToRelease); err != nil {
			return c.Edit(fmt.Sprintf("Failed to release stand: %v", err))
		}

		return c.Edit(fmt.Sprintf(
			"@%s has released %s",
			senderUsername,
			standName),
		)
	}

	var (
		menu           = make([][]telebot.InlineButton, 0)
		row            []telebot.InlineButton
		senderUsername = c.Sender().Username
	)

	for _, stand := range stands {
		if stand.Released || stand.Name == "" || stand.OwnerUsername != senderUsername {
			continue
		}

		btn := telebot.InlineButton{
			Text: fmt.Sprintf("%s %s", comp, stand.Name),
			Data: fmt.Sprintf("release:%s", stand.Name),
		}
		row = append(row, btn)

		if len(row) == 2 {
			menu = append(menu, row)
			row = []telebot.InlineButton{}
		}
	}

	if len(row) > 0 {
		menu = append(menu, row)
	}

	if len(menu) == 0 {
		return c.Reply("You have no stands to release")
	}

	return c.Reply("Choose stand to release:", &telebot.ReplyMarkup{
		InlineKeyboard: menu,
	})
}

func (h *Handler) Greetings(c telebot.Context) error {
	return c.Send(fmt.Sprintf(
		"Hello @%s, I'm StandClaimer bot, I will help you to manage environments across the team. "+
			"Tap `/` on the group menu to see commands",
		c.Message().UserJoined.Username,
	))
}

func (h *Handler) CommandHandlers() map[string]telebot.HandlerFunc {
	return map[string]telebot.HandlerFunc{
		"/list":     h.ListStands,
		"/ping":     h.Ping,
		"/ping_all": h.PingAll,
	}
}

func (h *Handler) CallbackHandlers() map[string]telebot.HandlerFunc {
	return map[string]telebot.HandlerFunc{
		"/claim":   h.Claim,
		"/release": h.Release,
		"/ping":    h.Ping,
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
