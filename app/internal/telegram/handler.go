package telegram

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/tibeahx/claimer/app/internal/repo"
	"github.com/tibeahx/claimer/pkg/entity"
	"gopkg.in/telebot.v4"
)

type notifierFunc func(chatID int64, users ...string) error

type Handler struct {
	repo *repo.Repo
	bot  *Bot
}

type inlineButton struct {
	text string
	data string
}

func NewHandler(b *Bot, repo *repo.Repo) *Handler {
	return &Handler{
		repo: repo,
		bot:  b,
	}
}

func (h *Handler) HandleCallbacks(c telebot.Context) error {
	data := strings.Split(c.Callback().Data, ":")
	if len(data) != 2 {
		return c.Respond()
	}

	action, standName := data[0], data[1]

	c.Message().Payload = standName

	handlers := h.CallbackHandlers()

	if h, ok := handlers["/"+action]; ok {
		err := h(c)
		if err != nil {
			return err
		}
		return c.Respond()
	}

	return c.Respond()
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
	stands, err := h.checkStands(c)
	if err != nil {
		return err
	}

	var (
		mentions = make(map[string]string)
		parts    = make([]string, 0, len(stands))
	)

	for _, stand := range stands {
		if stand.Released {
			continue
		}
		if stand.OwnerUsername.String != "" && stand.Name != "" {
			mentions[stand.OwnerUsername.String] = stand.Name
			parts = append(parts, fmt.Sprintf(TplUserStand, stand.OwnerUsername.String, stand.Name))
		}
	}

	if len(mentions) == 0 {
		return c.Reply(ErrNoBusyStands)
	}

	message := fmt.Sprintf(TplPingAllUsers, strings.Join(parts, ", "))
	return c.Send(message)
}

func (h *Handler) Ping(c telebot.Context) error {
	stands, err := h.checkStands(c)
	if err != nil {
		return err
	}

	if c.Callback() != nil {
		username := c.Message().Payload

		for _, stand := range stands {
			if stand.OwnerUsername.String == username && !stand.Released {
				return c.Edit(fmt.Sprintf(TplPingUser, username))
			}
		}
		return c.Edit(ErrNoBusyStands)
	}

	buttons := make([]inlineButton, 0, len(stands))
	usersToPing := make(map[string]struct{}, len(stands))

	for _, stand := range stands {
		if stand.Released || stand.OwnerUsername.String == "" {
			continue
		}

		if _, exists := usersToPing[stand.OwnerUsername.String]; !exists {
			usersToPing[stand.OwnerUsername.String] = struct{}{}
			buttons = append(buttons, inlineButton{
				text: fmt.Sprintf(TplButtonUser, stand.OwnerUsername.String, stand.Name),
				data: fmt.Sprintf("ping:%s", stand.OwnerUsername.String),
			})
		}
	}

	if len(buttons) == 0 {
		return c.Reply(ErrNoBusyStands)
	}

	menu := createInlineKeyboard(buttons)
	return c.Reply(MsgChooseUserToPing, &telebot.ReplyMarkup{
		InlineKeyboard: menu,
	})
}

func (h *Handler) ListStands(c telebot.Context) error {
	stands, err := h.checkStands(c)
	if err != nil {
		return err
	}

	standInfos := make([]string, 0, len(stands))

	for _, stand := range stands {
		if stand.Name == "" {
			continue
		}

		standInfo := fmt.Sprintf(TplStandInfo,
			EmojiComputer,
			stand.Name,
			formatStandStatus(stand),
		)

		standInfos = append(standInfos, standInfo)
	}

	if len(standInfos) == 0 {
		return c.Reply(ErrNoEnvironments)
	}

	message := strings.Join(standInfos, "\n")

	return c.Reply(message)
}

func (h *Handler) Claim(c telebot.Context) error {
	stands, err := h.checkStands(c)
	if err != nil {
		return err
	}

	if c.Callback() != nil {
		standName := c.Message().Payload
		senderUsername := c.Callback().Sender.Username

		if err := h.repo.CreateUser(senderUsername); err != nil {
			return c.Edit(fmt.Sprintf("failed to create user: %v", err))
		}

		for _, stand := range stands {
			if stand.Name == standName {
				if stand.Released {
					standToClaim := entity.Stand{
						Name:          standName,
						OwnerUsername: sql.NullString{String: senderUsername, Valid: true},
					}

					if err := h.repo.ClaimStand(standToClaim); err != nil {
						return c.Edit(fmt.Sprintf(ErrFailedToClaim, err))
					}

					return c.Edit(fmt.Sprintf(TplStandClaimed, senderUsername, standName))
				}
				return c.Edit(ErrStandBusy)
			}
		}
		return c.Edit(ErrStandNotFound)
	}

	buttons := make([]inlineButton, 0, len(stands))

	for _, stand := range stands {
		if !stand.Released || stand.Name == "" {
			continue
		}

		buttons = append(buttons, inlineButton{
			text: fmt.Sprintf(TplButtonStand, EmojiComputer, stand.Name),
			data: fmt.Sprintf("claim:%s", stand.Name),
		})
	}

	if len(buttons) == 0 {
		return c.Reply(ErrNoFreeStands)
	}

	menu := createInlineKeyboard(buttons)
	return c.Reply(MsgChooseStand, &telebot.ReplyMarkup{
		InlineKeyboard: menu,
	})
}

func (h *Handler) Release(c telebot.Context) error {
	stands, err := h.checkStands(c)
	if err != nil {
		return err
	}

	if c.Callback() != nil {
		standName := c.Message().Payload
		senderUsername := c.Callback().Sender.Username

		standToRelease := entity.Stand{
			Name:          standName,
			OwnerUsername: sql.NullString{String: senderUsername},
		}

		if err := h.repo.ReleaseStand(standToRelease); err != nil {
			return c.Edit(fmt.Sprintf(ErrFailedToRelease, err))
		}

		return c.Edit(fmt.Sprintf(TplStandReleased, senderUsername, standName))
	}

	buttons := make([]inlineButton, 0, len(stands))
	senderUsername := c.Sender().Username

	for _, stand := range stands {
		if stand.Released || stand.Name == "" || stand.OwnerUsername.String != senderUsername {
			continue
		}

		buttons = append(buttons, inlineButton{
			text: fmt.Sprintf(TplButtonStand, EmojiComputer, stand.Name),
			data: fmt.Sprintf("release:%s", stand.Name),
		})
	}

	if len(buttons) == 0 {
		return c.Reply(ErrNoStandsToRelease)
	}

	menu := createInlineKeyboard(buttons)
	return c.Reply(MsgChooseToRelease, &telebot.ReplyMarkup{
		InlineKeyboard: menu,
	})
}

func (h *Handler) Greetings(c telebot.Context) error {
	return c.Send(fmt.Sprintf(
		TplGreetings,
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

func (h *Handler) checkStands(c telebot.Context) ([]entity.Stand, error) {
	stands, err := h.repo.Stands()
	if err != nil {
		return nil, err
	}

	if len(stands) == 0 {
		return nil, c.Reply(ErrNoEnvironments)
	}

	return stands, nil
}

func createInlineKeyboard(items []inlineButton) [][]telebot.InlineButton {
	var (
		menu = make([][]telebot.InlineButton, 0, (len(items)+1)/2)
		row  = make([]telebot.InlineButton, 0, 2)
	)

	for _, item := range items {
		btn := telebot.InlineButton{
			Text: item.text,
			Data: item.data,
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

	return menu
}

func formatStandStatus(stand entity.Stand) string {
	if !stand.Released {
		if !stand.TimeClaimed.Valid {
			return fmt.Sprintf(
				TplStandBusyBy,
				stand.OwnerUsername.String,
				0,
				EmojiBusy,
			)
		}

		timeBusy := time.Since(stand.TimeClaimed.Time)

		return fmt.Sprintf(
			TplStandBusyBy,
			stand.OwnerUsername.String,
			int(timeBusy.Hours()),
			EmojiBusy,
		)
	}

	return fmt.Sprintf(TplStandFree, EmojiFree)
}
