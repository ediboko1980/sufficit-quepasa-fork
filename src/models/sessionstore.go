package models

import (
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type Store struct {
	BotID     string `db:"bot_id"`
	Data      []byte `db:"data"`
	CreatedAt string `db:"created_at"`
	UpdatedAt string `db:"updated_at"`
}

func storeExists(db *sqlx.DB, wid string) (bool, error) {
	var count int
	err := db.Get(&count, "SELECT count(*) FROM sessionstore WHERE bot_id = ?", wid)
	return count > 0, err
}

func GetOrCreateStore(db *sqlx.DB, wid string) (Store, error) {
	var store Store
	exists, err := storeExists(db, wid)
	if err != nil {
		return store, err
	}

	if exists {
		store, err := GetStore(db, wid)
		if err != nil {
			return store, err
		}
	} else {
		store, err = CreateStore(db, wid)
	}
	return store, err
}

func CreateStore(db *sqlx.DB, wid string) (Store, error) {
	var user Store
	now := time.Now()
	query := `INSERT INTO sessionstore (bot_id, created_at, updated_at) VALUES (?, ?, ?)`
	if _, err := db.Exec(query, wid, now, now); err != nil {
		return user, err
	}

	return GetStore(db, wid)
}

func GetStore(db *sqlx.DB, wid string) (Store, error) {
	var store Store
	err := db.Get(&store, "SELECT * FROM sessionstore WHERE `bot_id` = ?", wid)
	return store, err
}

func UpdateStore(db *sqlx.DB, wid string, data []byte) ([]byte, error) {
	now := time.Now()
	query := "UPDATE sessionstore SET data = ?, updated_at = ? WHERE bot_id = ?"
	_, err := db.Exec(query, data, now, wid)
	return data, err
}

func DeleteStore(db *sqlx.DB, wid string) error {
	query := "DELETE FROM sessionstore WHERE bot_id = ?"
	_, err := db.Exec(query, wid)
	return err
}
