package models

import (
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type QPStoreMysql struct {
	db *sqlx.DB
}

func (source QPStoreMysql) Exists(wid string) (bool, error) {
	var count int
	err := source.db.Get(&count, "SELECT count(*) FROM sessionstore WHERE bot_id = ?", wid)
	return count > 0, err
}

func (source QPStoreMysql) GetOrCreate(wid string) (QPStore, error) {
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

func (source QPStoreMysql) Create(wid string) (QPStore, error) {
	var user QPStore
	now := time.Now()
	query := `INSERT INTO sessionstore (bot_id, created_at, updated_at) VALUES (?, ?, ?)`
	if _, err := source.db.Exec(query, wid, now, now); err != nil {
		return user, err
	}

	return source.Get(wid)
}

func (source QPStoreMysql) Get(wid string) (QPStore, error) {
	var store QPStore
	err := source.db.Get(&store, "SELECT * FROM sessionstore WHERE `bot_id` = ?", wid)
	return store, err
}

func (source QPStoreMysql) Update(wid string, data []byte) ([]byte, error) {
	now := time.Now()
	query := "UPDATE sessionstore SET data = ?, updated_at = ? WHERE bot_id = ?"
	_, err := source.db.Exec(query, data, now, wid)
	return data, err
}

func (source QPStoreMysql) Delete(wid string) error {
	query := "DELETE FROM sessionstore WHERE bot_id = ?"
	_, err := source.db.Exec(query, wid)
	return err
}
