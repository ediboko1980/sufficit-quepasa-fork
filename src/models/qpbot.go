package models

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type QPBot struct {
	ID        string `db:"id" json:"id"`
	Number    string `db:"number" json:"number"`
	Verified  bool   `db:"is_verified" json:"is_verified"`
	Token     string `db:"token" json:"token"`
	UserID    string `db:"user_id" json:"user_id"`
	WebHook   string `db:"webhook" json:"webhook"`
	CreatedAt string `db:"created_at" json:"created_at"`
	UpdatedAt string `db:"updated_at" json:"updated_at"`
}

func FindAllBots(db *sqlx.DB) ([]QPBot, error) {
	bots := []QPBot{}
	err := db.Select(&bots, "SELECT * FROM bots")
	return bots, err
}

func FindAllBotsForUser(db *sqlx.DB, userID string) ([]QPBot, error) {
	bots := []QPBot{}
	err := db.Select(&bots, "SELECT * FROM bots WHERE user_id = $1", userID)
	return bots, err
}

func FindBotByToken(db *sqlx.DB, token string) (QPBot, error) {
	var bot QPBot
	err := db.Get(&bot, "SELECT * FROM bots WHERE token = $1", token)
	return bot, err
}

func FindBotForUser(db *sqlx.DB, userID string, ID string) (QPBot, error) {
	var bot QPBot
	err := db.Get(&bot, "SELECT * FROM bots WHERE user_id = $1 AND id = $2", userID, ID)
	return bot, err
}

func FindBotByNumber(db *sqlx.DB, number string) (QPBot, error) {
	var bot QPBot
	err := db.Get(&bot, "SELECT * FROM bots WHERE number = $1", number)
	return bot, err
}

func CreateBot(db *sqlx.DB, userID string, number string) (QPBot, error) {
	var bot QPBot
	botID := uuid.New().String()
	token := uuid.New().String()
	now := time.Now().Format(time.RFC3339)
	query := `INSERT INTO bots
    (id, number, is_verified, token, user_id, created_at, updated_at, webhook)
    VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	if _, err := db.Exec(query, botID, number, false, token, userID, now, now, ""); err != nil {
		return bot, err
	}

	return FindBotForUser(db, userID, botID)
}

func (bot *QPBot) MarkVerified(db *sqlx.DB) error {
	now := time.Now().Format(time.RFC3339)
	query := "UPDATE bots SET is_verified = true, updated_at = $1 WHERE id = $2"
	_, err := db.Exec(query, now, bot.ID)
	return err
}

func (bot *QPBot) CycleToken(db *sqlx.DB) error {
	token := uuid.New().String()
	now := time.Now().Format(time.RFC3339)
	query := "UPDATE bots SET token = $1, updated_at = $2 WHERE id = $3"
	_, err := db.Exec(query, token, now, bot.ID)
	return err
}

func (bot *QPBot) Delete(db *sqlx.DB) error {
	query := "DELETE FROM bots WHERE id = $1"
	_, err := db.Exec(query, bot.ID)
	return err
}

func (bot *QPBot) FormattedNumber() string {
	phoneNumber, err := CleanPhoneNumber(bot.Number)
	if err != nil {
		log.Printf("SUFF ERROR G :: error on regex: %v\n", err)
	}
	return phoneNumber
}

func (bot *QPBot) WebHookUpdate(db *sqlx.DB) error {
	now := time.Now().Format(time.RFC3339)
	query := "UPDATE bots SET webhook = $1, updated_at = $2 WHERE id = $3"
	_, err := db.Exec(query, bot.WebHook, now, bot.ID)
	return err
}

func (bot *QPBot) WebHookSincronize(db *sqlx.DB) {
	db.Get(&bot.WebHook, "SELECT webhook FROM bots WHERE number = $1", bot.Number)
}

// Encaminha msg ao WebHook específicado
func (bot *QPBot) PostToWebHook(message QPMessage) error {
	if len(bot.WebHook) > 0 {
		payloadJson, _ := json.Marshal(&struct {
			Message QPMessage `json:"message"`
		}{Message: message})
		requestBody := bytes.NewBuffer(payloadJson)
		resp, _ := http.Post(bot.WebHook, "application/json", requestBody)

		if resp != nil {
			defer resp.Body.Close()
			if resp.StatusCode == 422 {
				body, _ := ioutil.ReadAll(resp.Body)
				if body != nil && strings.Contains(string(body), "invalid callback token") {

					// Sincroniza o token mais novo
					bot.WebHookSincronize(GetDB())

					// Preenche o body novamente pois foi esvaziado na requisição anterior
					requestBody = bytes.NewBuffer(payloadJson)
					http.Post(bot.WebHook, "application/json", requestBody)
				}
			}
		}
	}
	return nil
}
