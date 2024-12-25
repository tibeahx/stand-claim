package telegram

import (
	"fmt"

	"github.com/tibeahx/claimer/app/internal/repo"
	"github.com/tibeahx/claimer/pkg/entity"
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

// у меня вообще по идее все данные в базе уже prefill
// мне нужно решить, какие именно данные оставить в базе, а какие добавлять из запросов

// получаем список всех стендов для ownerInfo.ownerID, ownerInfo.ownerUserName
// затем мапим [ownerInfo]Stand для всех стендов где !released
func (h *Handler) Ping(c telebot.Context) error {
	stands, err := h.repo.Stands()
	if err != nil {
		return err
	}

	if len(stands) == 0 {
		return c.Reply("No environments found")
	}

	userStandsMap := make(map[string]entity.Stand)

	for _, stand := range stands {
		if !stand.Released {
			userStandsMap[stand.OwnerUsername] = stand
		}
	}
	// пропинговать всех юзеров которые имеют стенды за собой активные указав при этом на стенд который надо бы освободить
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

// получаем список всех стендов
// проходимся по стендам, проверяем если стенд занят, отправляем сообщение об ошибка
// если не занят, то мы можем его заклеймить
func (h *Handler) Claim(c telebot.Context) error {
	stands, err := h.repo.Stands()
	if err != nil {
		return err
	}

	freeStands := make(map[string]struct{})

	for _, stand := range stands {
		if !stand.Released {
			return c.Reply("stand is busy, choose another free one")
		}
		freeStands[stand.Name] = struct{}{}
	}

	standName := c.Sender().Username
	if _, found := freeStands[standName]; found {
		h.repo.ClaimStand(standName, entity.NewOwner(c))
		return c.Send("@%s has claimed %s", c.Message().Sender.Username)
	}

	return c.Reply("stand not found")
}

func (h *Handler) Release(c telebot.Context) error {
	return nil
}

func (h *Handler) Greetings(c telebot.Context) error {
	joined := c.Message().UserJoined

	greeting := fmt.Sprintf(
		"Hello @%s, I'm a StandClaimer bot, I will help you to manage enviroments accross the team. Tap `/` on the group menu to see commands",
		joined.Username,
	)

	return c.Reply(greeting)
}

func (h *Handler) Handlers() map[string]telebot.HandlerFunc {
	return map[string]telebot.HandlerFunc{
		"/list":    h.ListStands,
		"/claim":   h.Claim,
		"/release": h.Release,
		"/ping":    h.Ping,
	}
}
