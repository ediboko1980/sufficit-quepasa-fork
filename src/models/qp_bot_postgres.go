package models

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type QPBotPostgres struct {
	db *sqlx.DB
}

func (source QPBotPostgres) FindAll() ([]QPBot, error) {
	bots := []QPBot{}
	err := source.db.Select(&bots, "SELECT * FROM bots")
	return bots, err
}

func (source QPBotPostgres) FindAllForUser(userID string) ([]QPBot, error) {
	bots := []QPBot{}
	err := source.db.Select(&bots, "SELECT * FROM bots WHERE user_id = $1", userID)
	return bots, err
}

func (source QPBotPostgres) FindByToken(token string) (QPBot, error) {
	var bot QPBot
	err := source.db.Get(&bot, "SELECT * FROM bots WHERE token = $1", token)
	return bot, err
}

func (source QPBotPostgres) FindForUser(userID string, ID string) (QPBot, error) {
	var bot QPBot
	err := source.db.Get(&bot, "SELECT * FROM bots WHERE user_id = $1 AND id = $2", userID, ID)
	return bot, err
}

func (source QPBotPostgres) FindByID(botID string) (QPBot, error) {
	var bot QPBot
	err := source.db.Get(&bot, "SELECT * FROM bots WHERE id = $1", botID)
	return bot, err
}

func (source QPBotPostgres) GetOrCreate(botID string, userID string) (bot QPBot, err error) {
	bot, err = source.FindByID(botID)
	if err != nil {
		if strings.Contains(err.Error(), "no rows in result set") {
			bot, err = source.Create(botID, userID)
		}
	}
	return
}

// botID = Wid of whatsapp connection
func (source QPBotPostgres) Create(botID string, userID string) (QPBot, error) {
	var bot QPBot
	token := uuid.New().String()
	now := time.Now()
	query := `INSERT INTO bots
    (id, is_verified, token, user_id, created_at, updated_at, webhook, devel)
    VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	if _, err := source.db.Exec(query, botID, false, token, userID, now, now, "", false); err != nil {
		return bot, err
	}

	return source.FindForUser(userID, botID)
}

func (source QPBotPostgres) MarkVerified(id string, ok bool) error {
	now := time.Now()
	query := "UPDATE bots SET is_verified = $1, updated_at = $2 WHERE id = $3"
	_, err := source.db.Exec(query, ok, now, id)
	return err
}

func (source QPBotPostgres) CycleToken(id string) error {
	token := uuid.New().String()
	now := time.Now()
	query := "UPDATE bots SET token = $1, updated_at = $2 WHERE id = $3"
	_, err := source.db.Exec(query, token, now, id)
	return err
}

func (source QPBotPostgres) Delete(id string) error {
	query := "DELETE FROM bots WHERE id = $1"
	_, err := source.db.Exec(query, id)
	return err
}

func (source QPBotPostgres) WebHookUpdate(webhook string, id string) error {
	now := time.Now()
	query := "UPDATE bots SET webhook = $1, updated_at = $2 WHERE id = $3"
	_, err := source.db.Exec(query, webhook, now, id)
	return err
}

func (source QPBotPostgres) WebHookSincronize(id string) (result string, err error) {
	err = source.db.Get(&result, "SELECT webhook FROM bots WHERE id = $1", id)
	return result, err
}

func (source QPBotPostgres) Devel(id string, status bool) (err error) {
	now := time.Now()
	query := "UPDATE bots SET devel = $1, updated_at = $2 WHERE id = $3"
	_, err = source.db.Exec(query, status, now, id)
	return err
}
