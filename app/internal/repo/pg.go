package repo

import (
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/tibeahx/claimer/pkg/dbutils"
	"github.com/tibeahx/claimer/pkg/entity"
)

type Repo struct {
	db *sqlx.DB
}

func NewRepo(db *sqlx.DB) *Repo {
	return &Repo{
		db: db,
	}
}

func (r *Repo) Stands() ([]entity.Stand, error) {
	const q = `
	select
	id,
	name,
	released,
	owner_id,
	owner_username,
	time_claimed,
	time_released
from
	stands
order by
	name asc
	`

	var stands []entity.Stand
	err := dbutils.NamedSelect(
		r.db,
		q,
		&stands,
		nil,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no stands found")
		}
		return nil, fmt.Errorf("failed to get stands: %w", err)
	}

	return stands, nil
}

// func (r *Repo) StandsByOwner(owner entity.Owner) ([]entity.Stand, error) {
// 	const q = `select
// 	id,
// 	name,
// 	released,
// 	owner_id,
// 	owner_username,
// 	time_claimed,
// 	time_released
// from
// 	stands
// where
// 	owner_id = :owner_id,
// 	`
// }

func (r *Repo) FreeStands() ([]entity.Stand, error) {
	return nil, nil
}

func (r *Repo) ClaimStand(standName string, owner entity.Owner) error {
	const q = `insert into
	stands (owner_id, owner_username, time_claimed)
values
	(:owner_id, :owner_username, :now ())
where
	name = :name;
	`

	return dbutils.NamedExec(
		r.db,
		q,
		map[string]any{
			"owner_id":       owner.ID,
			"owner_username": owner.Username,
			"name":           standName,
		},
	)
}

func (r *Repo) ReleaseStand(stand entity.Stand) (string, error) {
	return "", nil
}
