package models

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

type QPBotMysql struct {
	db *sqlx.DB
}

func (source QPBotMysql) FindAll() ([]QPBot, error) {
	bots := []QPBot{}
	err := source.db.Select(&bots, "SELECT * FROM bots")
	return bots, err
}

func (source QPBotMysql) FindAllForUser(userID string) ([]QPBot, error) {
	bots := []QPBot{}
	err := source.db.Select(&bots, "SELECT * FROM bots WHERE user_id = ?", userID)
	return bots, err
}

func (source QPBotMysql) FindByToken(token string) (QPBot, error) {
	var bot QPBot
	err := source.db.Get(&bot, "SELECT * FROM bots WHERE token = ?", token)
	return bot, err
}

func (source QPBotMysql) FindForUser(userID string, ID string) (QPBot, error) {
	var bot QPBot
	err := source.db.Get(&bot, "SELECT * FROM bots WHERE user_id = ? AND id = ?", userID, ID)
	return bot, err
}

func (source QPBotMysql) FindByID(botID string) (QPBot, error) {
	var bot QPBot
	err := source.db.Get(&bot, "SELECT * FROM bots WHERE id = ?", botID)
	return bot, err
}

func (source QPBotMysql) GetOrCreate(botID string, userID string) (bot QPBot, err error) {
	bot, err = source.FindByID(botID)
	if err != nil {
		if strings.Contains(err.Error(), "no rows in result set") {
			bot, err = source.Create(botID, userID)
		}
	}
	return
}

// botID = Wid of whatsapp connection
func (source QPBotMysql) Create(botID string, userID string) (QPBot, error) {
	var bot QPBot
	token := uuid.New().String()
	now := time.Now()
	query := `INSERT INTO bots
    (id, is_verified, token, user_id, created_at, updated_at, webhook, devel)
    VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	if _, err := source.db.Exec(query, botID, false, token, userID, now, now, "", false); err != nil {
		return bot, err
	}

	return source.FindForUser(userID, botID)
}

func (source QPBotMysql) MarkVerified(id string, ok bool) error {
	now := time.Now()
	query := "UPDATE bots SET is_verified = ?, updated_at = ? WHERE id = ?"
	_, err := source.db.Exec(query, ok, now, id)
	return err
}

func (source QPBotMysql) CycleToken(id string) error {
	token := uuid.New().String()
	now := time.Now()
	query := "UPDATE bots SET token = ?, updated_at = ? WHERE id = ?"
	_, err := source.db.Exec(query, token, now, id)
	return err
}

func (source QPBotMysql) Delete(id string) error {
	query := "DELETE FROM bots WHERE id = ?"
	_, err := source.db.Exec(query, id)
	return err
}

func (source QPBotMysql) WebHookUpdate(webhook string, id string) error {
	now := time.Now()
	query := "UPDATE bots SET webhook = ?, updated_at = ? WHERE id = ?"
	_, err := source.db.Exec(query, webhook, now, id)
	return err
}

func (source QPBotMysql) WebHookSincronize(id string) (result string, err error) {
	err = source.db.Get(&result, "SELECT webhook FROM bots WHERE id = ?", id)
	return result, err
}

func (source QPBotMysql) Devel(id string, status bool) (err error) {
	now := time.Now()
	query := "UPDATE bots SET devel = ?, updated_at = ? WHERE id = ?"
	_, err = source.db.Exec(query, status, now, id)
	return err
}
