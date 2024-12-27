package workers

import (
	"github.com/tibeahx/claimer/app/internal/repo"
	"github.com/tibeahx/claimer/pkg/entity"
)

// todo надо сделать userNamager чтобы он фоном работал и ждал пока сработает
// событие на вхождение юзера в чат
// как только вошел в чат, дернуть метод базы для добавления юзера в таблице
// как только вышел из чата, дернуть метод базы для удаления юзера из таблицы
type UserManager struct {
	repo     *repo.Repo
	manageFn func(groupInfo entity.ChatInfo) error
}

func NewUserManager(repo *repo.Repo, fn func(groupInfo entity.ChatInfo) error) *UserManager {
	return &UserManager{
		repo:     repo,
		manageFn: fn,
	}
}

func (m *UserManager) Start() {}
