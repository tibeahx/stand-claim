package service

import (
	"github.com/tibeahx/claimer/app/internal/entity"
	"github.com/tibeahx/claimer/app/internal/repo"
)

type Service struct {
	repo *repo.Repo
}

func NewService(repo *repo.Repo) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) Ping(owner entity.Owner) error {
	return nil
}

func (s *Service) ListStands() ([]entity.Stand, error) {
	return s.repo.Stands()
}

func (s *Service) ListFree() ([]entity.Stand, error) {
	return s.repo.FreeStands()
}

func (s *Service) Claim(stand entity.Stand) error {
	return s.repo.ClaimStand(stand)
}

func (s *Service) Release(stand entity.Stand) (string, error) {
	return s.repo.ReleaseStand(stand)
}
