package repo

import (
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/tibeahx/claimer/app/internal/entity"
	"github.com/tibeahx/claimer/pkg/dbutils"
	"gopkg.in/telebot.v4"
)

type Repo struct {
	db *sqlx.DB
}

func NewRepo(db *sqlx.DB) *Repo {
	return &Repo{
		db: db,
	}
}

func (r *Repo) Stands(c telebot.Context) ([]string, error) {
	const q = `select
	name
from
	stands
order by
	name desc
	`

	var names []string
	err := dbutils.NamedSelect(
		r.db,
		q,
		&names,
		nil,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no stands found")
		}
	}

	return names, nil
}

func (r *Repo) FreeStands() ([]entity.Stand, error) {
	return nil, nil
}

func (r *Repo) ClaimStand(stand entity.Stand) error {
	return nil
}

func (r *Repo) ReleaseStand(stand entity.Stand) (string, error) {
	return "", nil
}
