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
	owner_username,
	time_claimed
from
	stands
where
	name is not null
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
	time_claimed = now (),
	released = false
where
	name = :name
	and released = true
	`

	return dbutils.NamedExec(
		r.db,
		q,
		map[string]any{
			"owner_username": stand.OwnerUsername.String,
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
			"owner_username": stand.OwnerUsername.String,
			"name":           stand.Name,
		},
	)
}

func (r *Repo) FindUser(username string) (bool, error) {
	const q = `
with
	user_check as (
		select
			exists (
				select
					1
				from
					users
				where
					username = :username
			) as user_exists
	),
	claimed_stands as (
		select
			count(*) as claimed_count
		from
			stands
		where
			owner_username = :username
			and released = false
	)
select
	case
		when user_check.user_exists = true
		and claimed_stands.claimed_count = 0 then true
		else false
	end as can_claim
from
	user_check,
	claimed_stands;
	`

	var canClaim bool
	err := dbutils.NamedGet(
		r.db,
		q,
		&canClaim,
		map[string]any{
			"username": username,
		},
	)

	if err == sql.ErrNoRows {
		return false, nil
	}

	if err != nil {
		return false, fmt.Errorf("failed to check user status: %w", err)
	}

	return canClaim, nil
}

func (r *Repo) DeleteUser(username string) error {
	const q = `
delete from users
where
	username = :username
	`

	return dbutils.NamedExec(
		r.db,
		q,
		map[string]any{
			"username": username,
		},
	)
}
