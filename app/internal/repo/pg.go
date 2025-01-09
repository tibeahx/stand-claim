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
	name,
	released,
	coalesce(owner_username, '') as owner_username,
	time_claimed
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

func (r *Repo) CreateUser(username string) error {
	const q = `
	insert into
	users (username, created)
values
	(:username, now ()) on conflict (username) do nothing
	`
	return dbutils.NamedExec(
		r.db,
		q,
		map[string]any{
			"username": username,
		},
	)
}

func (r *Repo) ClaimStand(stand entity.Stand) error {
	const q = `
	update stands
	set
		owner_username = :owner_username,
		time_claimed = now(),
		released = false
	where
		name = :name
		and released = true
	`

	return dbutils.NamedExec(
		r.db,
		q,
		map[string]any{
			"owner_username": stand.OwnerUsername,
			"name":           stand.Name,
		},
	)
}

func (r *Repo) ReleaseStand(stand entity.Stand) error {
	const q = `
	update stands
	set
		owner_username = null,
		released = true
	where
		name = :name
		and released = false
		and owner_username = :owner_username
	`

	return dbutils.NamedExec(
		r.db,
		q,
		map[string]any{
			"owner_username": stand.OwnerUsername,
			"name":           stand.Name,
		},
	)
}

func (r *Repo) FindUser(username string) (bool, error) {
	const q = `select
	u.username
from
	users u
	left join stands s on u.username = s.owner_username
	and s.released = false
where
	u.username = :username
	and
	u.username is not null
group by
	u.username
having
	count(s.owner_username) = 0;`

	var usernameFound string

	err := dbutils.NamedGet(
		r.db,
		q,
		&usernameFound,
		map[string]any{
			"username": username,
		},
	)

	return len(usernameFound) != 0, err
}

func (r *Repo) DeleteUser(username string) error {
	const q = `
	delete from users
	where username = :username
	`

	return dbutils.NamedExec(
		r.db,
		q,
		map[string]any{
			"username": username,
		},
	)
}
