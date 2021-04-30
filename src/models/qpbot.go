package models

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type QPBot struct {
	ID        string `db:"id" json:"id"`
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

func FindBotByID(db *sqlx.DB, botID string) (QPBot, error) {
	var bot QPBot
	err := db.Get(&bot, "SELECT * FROM bots WHERE id = $1", botID)
	return bot, err
}

func GetOrCreateBot(db *sqlx.DB, botID string, userID string) (bot QPBot, err error) {
	bot, err = FindBotByID(db, botID)
	if err != nil {
		if strings.Contains(err.Error(), "no rows in result set") {
			bot, err = CreateBot(db, botID, userID)
		}
	}
	return
}

// botID = Wid of whatsapp connection
func CreateBot(db *sqlx.DB, botID string, userID string) (QPBot, error) {
	var bot QPBot
	token := uuid.New().String()
	now := time.Now().Format(time.RFC3339)
	query := `INSERT INTO bots
    (id, is_verified, token, user_id, created_at, updated_at, webhook)
    VALUES ($1, $2, $3, $4, $5, $6, $7)`
	if _, err := db.Exec(query, botID, false, token, userID, now, now, ""); err != nil {
		return bot, err
	}

	return FindBotForUser(db, userID, botID)
}

func (bot *QPBot) MarkVerified(db *sqlx.DB, ok bool) error {
	now := time.Now().Format(time.RFC3339)
	query := "UPDATE bots SET is_verified = $3, updated_at = $1 WHERE id = $2"
	_, err := db.Exec(query, now, bot.ID, ok)
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

// Traduz o Wid para um número de telefone em formato E164
func (bot *QPBot) GetNumber() string {
	phoneNumber, err := GetPhoneByID(bot.ID)
	if err != nil {
		return ""
	}
	return "+" + phoneNumber
}

func (bot *QPBot) WebHookUpdate(db *sqlx.DB) error {
	now := time.Now().Format(time.RFC3339)
	query := "UPDATE bots SET webhook = $1, updated_at = $2 WHERE id = $3"
	_, err := db.Exec(query, bot.WebHook, now, bot.ID)
	return err
}

func (bot *QPBot) WebHookSincronize(db *sqlx.DB) {
	db.Get(&bot.WebHook, "SELECT webhook FROM bots WHERE id = $1", bot.ID)
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
