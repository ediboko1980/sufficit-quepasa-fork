package models

import (
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type QPStorePostgres struct {
	db *sqlx.DB
}

func (source QPStorePostgres) Exists(wid string) (bool, error) {
	var count int
	err := source.db.Get(&count, "SELECT count(*) FROM sessionstore WHERE bot_id = $1", wid)
	return count > 0, err
}

func (source QPStorePostgres) GetOrCreate(wid string) (QPStore, error) {
	var store QPStore
	exists, err := source.Exists(wid)
	if err != nil {
		return store, err
	}

	if exists {
		store, err := source.Get(wid)
		if err != nil {
			return store, err
		}
	} else {
		store, err = source.Create(wid)
	}
	return store, err
}

func (source QPStorePostgres) Create(wid string) (QPStore, error) {
	var user QPStore
	now := time.Now().Format(time.RFC3339)
	query := `INSERT INTO sessionstore
    (bot_id, created_at, updated_at)
    VALUES ($1, $2, $3)`
	if _, err := source.db.Exec(query, wid, now, now); err != nil {
		return user, err
	}

	return source.Get(wid)
}

func (source QPStorePostgres) Get(wid string) (QPStore, error) {
	var store QPStore
	err := source.db.Get(&store, "SELECT * FROM sessionstore WHERE bot_id = $1", wid)
	return store, err
}

func (source QPStorePostgres) Update(wid string, data []byte) ([]byte, error) {
	now := time.Now().Format(time.RFC3339)
	query := "UPDATE sessionstore SET data = ($1::bytea), updated_at = $2 WHERE bot_id = $3"
	_, err := source.db.Exec(query, data, now, wid)
	return data, err
}

func (source QPStorePostgres) Delete(wid string) error {
	query := "DELETE FROM sessionstore WHERE bot_id = $1"
	_, err := source.db.Exec(query, wid)
	return err
}
