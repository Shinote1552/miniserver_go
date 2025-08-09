package dto

import "time"

type (
	UserDB struct {
		ID        int64     `db:"id"`
		UUID      string    `db:"uuid"`
		CreatedAt time.Time `db:"created_at"`
	}
)

func UserDBToDomain(u UserDB) UserDB {
	return UserDB{
		ID:        u.ID,
		CreatedAt: u.CreatedAt,
	}
}

func UserDBFromDomain(u UserDB) UserDB {
	return UserDB{
		ID:        u.ID,
		CreatedAt: u.CreatedAt,
	}
}
