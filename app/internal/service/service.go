package service

import (
	"github.com/tibeahx/claimer/app/internal/entity"
	"github.com/tibeahx/claimer/app/internal/repo"
	"gopkg.in/telebot.v4"
)

type Service struct {
	repo *repo.Repo
}

func NewService(repo *repo.Repo) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) Ping(c telebot.Context, owner entity.Owner) error {
	return nil
}

func (s *Service) ListStands(c telebot.Context) ([]entity.Stand, error) {
	return s.repo.Stands()
}

func (s *Service) ListFree(c telebot.Context) ([]entity.Stand, error) {
	return s.repo.FreeStands()
}

func (s *Service) Claim(c telebot.Context, stand entity.Stand) error {
	return s.repo.ClaimStand(stand)
}

func (s *Service) Release(c telebot.Context, stand entity.Stand) (string, error) {

	return s.repo.ReleaseStand(stand)
}
