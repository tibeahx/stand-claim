package repo

import (
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/tibeahx/claimer/app/internal/entity"
	"github.com/tibeahx/claimer/pkg/dbutils"
)

const schemaStands = `
create table
	if not exists stands (
		id uuid primary key not null,
		name text unique,
		owner_id int,
		released bool,
		owner_username text,
		time_claimed timestamp,
		time_released timestamp,
	);
`

type Repo struct {
	db *sqlx.DB
}

func NewRepo(db *sqlx.DB) (*Repo, error) {
	r := &Repo{db: db}

	if err := r.migration(); err != nil {
		return nil, err
	}
	if err := r.prefill(); err != nil {
		return nil, err
	}

	return r, nil
}

func (r *Repo) migration() error {
	_, err := r.db.Exec(schemaStands)
	if err != nil {
		return fmt.Errorf("failed to init schema: %w", err)
	}

	return nil
}

var defaultStands = []string{"dev1", "dev2", "dev3", "dev4"}

func (r *Repo) prefill() error {
	const q = `insert into
	stands (id, name)
values
	(:id, :name);
	`

	for _, stand := range defaultStands {
		_, err := r.db.NamedExec(q, map[string]any{
			"name": stand,
			"id":   uuid.New().String(),
		})
		if err != nil {
			return fmt.Errorf("failed to exec prefill due to: %w", err)
		}
	}

	return nil
}

func (r *Repo) Stands() ([]entity.Stand, error) {
	const q = `select
	id,
	name,
	owner_id,
	released,
	owner_username,
	time_claimed,
	time_released
from
	stands`

	var stands []entity.Stand
	err := dbutils.NamedSelect(
		r.db,
		q,
		&stands,
		map[string]any{},
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no stands found")
		}
	}

	return stands, nil
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
