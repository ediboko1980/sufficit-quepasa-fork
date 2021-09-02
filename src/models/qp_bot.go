package models

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
)

type QPBot struct {
	ID        string `db:"id" json:"id"`
	Verified  bool   `db:"is_verified" json:"is_verified"`
	Token     string `db:"token" json:"token"`
	UserID    string `db:"user_id" json:"user_id"`
	WebHook   string `db:"webhook" json:"webhook"`
	CreatedAt string `db:"created_at" json:"created_at"`
	UpdatedAt string `db:"updated_at" json:"updated_at"`
	Devel     bool   `db:"devel" json:"devel"`
}

type IQPBot interface {
	FindAll() ([]QPBot, error)
	FindAllForUser(userID string) ([]QPBot, error)
	FindByToken(token string) (QPBot, error)
	FindForUser(userID string, ID string) (QPBot, error)
	FindByID(botID string) (QPBot, error)
	GetOrCreate(botID string, userID string) (bot QPBot, err error)
	Create(botID string, userID string) (QPBot, error)

	/// FORWARDING ---
	MarkVerified(id string, ok bool) error
	CycleToken(id string) error
	Delete(id string) error
	WebHookUpdate(webhook string, id string) error
	WebHookSincronize(id string) (result string, err error)
	Devel(id string, status bool) error
}

// Traduz o Wid para um número de telefone em formato E164
func (bot *QPBot) GetNumber() string {
	phoneNumber, err := GetPhoneByID(bot.ID)
	if err != nil {
		return ""
	}
	return "+" + phoneNumber
}

func (bot *QPBot) GetStatus() string {
	server, ok := WhatsAppService.Servers[bot.ID]
	if !ok {
		return "stopped"
	}

	if len(*server.Status) > 0 {
		return *server.Status
	}

	return "running"
}

func (bot *QPBot) GetBatteryInfo() WhatsAppBateryStatus {
	server, ok := WhatsAppService.Servers[bot.ID]
	if !ok {
		return WhatsAppBateryStatus{}
	}
	return *server.Battery
}

// Encaminha msg ao WebHook específicado
func (bot *QPBot) PostToWebHook(message QPMessage) error {
	if len(bot.WebHook) > 0 {
		payloadJson, _ := json.Marshal(message.ToV2())
		requestBody := bytes.NewBuffer(payloadJson)
		resp, _ := http.Post(bot.WebHook, "application/json", requestBody)

		if resp != nil {
			defer resp.Body.Close()
			if resp.StatusCode == 422 {
				body, _ := ioutil.ReadAll(resp.Body)
				if body != nil && strings.Contains(string(body), "invalid callback token") {

					// Sincroniza o token mais novo
					bot.WebHookSincronize()

					// Preenche o body novamente pois foi esvaziado na requisição anterior
					requestBody = bytes.NewBuffer(payloadJson)
					http.Post(bot.WebHook, "application/json", requestBody)
				}
			}
		}
	}
	return nil
}

func (bot *QPBot) Toggle() (err error) {
	server, ok := WhatsAppService.Servers[bot.ID]
	if !ok {
		go WhatsAppService.AppendNewServer(*bot)
	} else {
		if *server.Status == "stopped" || *server.Status == "created" {
			err = server.Start()
		} else {
			err = server.Shutdown()
		}
	}
	return
}

func (bot *QPBot) IsDevelopmentGlobal() bool {
	return ENV.IsDevelopment()
}

func (bot *QPBot) MarkVerified(ok bool) error {
	return WhatsAppService.DB.Bot.MarkVerified(bot.ID, ok)
}

func (bot *QPBot) CycleToken() error {
	return WhatsAppService.DB.Bot.CycleToken(bot.ID)
}

func (bot *QPBot) Delete() error {
	return WhatsAppService.DB.Bot.Delete(bot.ID)
}

func (bot *QPBot) WebHookUpdate() error {
	return WhatsAppService.DB.Bot.WebHookUpdate(bot.WebHook, bot.ID)
}

func (bot *QPBot) WebHookSincronize() error {
	webhook, err := WhatsAppService.DB.Bot.WebHookSincronize(bot.ID)
	bot.WebHook = webhook
	return err
}

func (bot *QPBot) ToggleDevel() (err error) {
	if bot.Devel {
		err = WhatsAppService.DB.Bot.Devel(bot.ID, false)
		bot.Devel = false
	} else {
		err = WhatsAppService.DB.Bot.Devel(bot.ID, true)
		bot.Devel = true
	}
	return err
}
