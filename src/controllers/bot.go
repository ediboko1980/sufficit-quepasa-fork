package controllers

import (
	"errors"
	"html/template"
	"log"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/gorilla/websocket"
	"gitlab.com/digiresilience/link/quepasa/common"
	"gitlab.com/digiresilience/link/quepasa/models"
)

// Register bot

type registerFormData struct {
	PageTitle string
}

type registerResponse struct {
	Result   string `json:"result"`
	NumberID string `json:"numberID"`
	Secret   string `json:"secret"`
}

func RegisterFormHandler(w http.ResponseWriter, r *http.Request) {
	data := registerFormData{
		PageTitle: "Register",
	}
	templates := template.Must(template.ParseFiles("views/layouts/main.tmpl", "views/bot/register.tmpl"))
	templates.ExecuteTemplate(w, "main", data)
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	user, err := common.GetUser(r)
	if err != nil {
		common.RedirectToLogin(w, r)
		return
	}

	r.ParseForm()
	number := r.Form.Get("number")

	if number == "" {
		common.RespondServerError(w, errors.New("Number is required"))
		return
	}

	bot, err := models.CreateBot(common.GetDB(), user.ID, number)
	if err != nil {
		common.RespondServerError(w, errors.New("Bot creation error"))
		return
	}

	http.Redirect(w, r, "/bot/"+bot.ID+"/verify", http.StatusFound)
}

func CycleHandler(w http.ResponseWriter, r *http.Request) {
	user, err := common.GetUser(r)
	if err != nil {
		return
	}

	r.ParseForm()
	botID := r.Form.Get("botID")
	bot, err := models.FindBotForUser(common.GetDB(), user.ID, botID)
	if err != nil {
		return
	}

	err = bot.CycleToken(common.GetDB())
	if err != nil {
		return
	}

	http.Redirect(w, r, "/account", http.StatusFound)
}

// Verify bot

type verifyFormData struct {
	PageTitle string
	Error     string
	Bot       models.Bot
}

func VerifyFormHandler(w http.ResponseWriter, r *http.Request) {
	bot, err := findBot(r)
	if err != nil {
		return
	}

	data := verifyFormData{
		PageTitle: "Verify",
		Error:     "",
		Bot:       bot,
	}

	templates := template.Must(template.ParseFiles(
		"views/layouts/main.tmpl",
		"views/bot/verify.tmpl",
	))
	templates.ExecuteTemplate(w, "main", data)
}

var upgrader = websocket.Upgrader{}

func VerifyHandler(w http.ResponseWriter, r *http.Request) {
	bot, err := findBot(r)
	if err != nil {
		return
	}

	con, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("Connection upgrade error: ", err)
		return
	}

	defer con.Close()

	out := make(chan []byte)
	go func() {
		err = con.WriteMessage(websocket.TextMessage, <-out)
		if err != nil {
			log.Println("Write message error: ", err)
		}
	}()

	if err = common.SignIn(bot, out); err != nil {
		err = bot.MarkVerified(common.GetDB())
		if err != nil {
			log.Println(err)
		}
		err = con.WriteMessage(websocket.TextMessage, []byte("Complete"))
		if err != nil {
			log.Println("Write message error: ", err)
		}
	}
}

// Send message
type sendFormData struct {
	PageTitle string
	Error     string
	Bot       models.Bot
}

func renderSendForm(w http.ResponseWriter, data sendFormData) {
	templates := template.Must(template.ParseFiles("views/layouts/main.tmpl", "views/bot/send.tmpl"))
	templates.ExecuteTemplate(w, "main", data)
}

func SendFormHandler(w http.ResponseWriter, r *http.Request) {
	var displayError string
	bot, err := findBot(r)
	if err != nil {
		displayError = err.Error()
		return
	}

	data := sendFormData{
		PageTitle: "Send",
		Error:     displayError,
		Bot:       bot,
	}

	renderSendForm(w, data)
}

func SendHandler(w http.ResponseWriter, r *http.Request) {
	var displayError string
	bot, err := findBot(r)
	if err != nil {
		displayError = err.Error()
		return
	}

	r.ParseForm()
	recipient := r.Form.Get("recipient")
	message := r.Form.Get("message")

	if err = common.SendMessage(bot, recipient, message); err != nil {
		common.RespondServerError(w, err)
		return
	}

	data := sendFormData{
		PageTitle: "Send",
		Error:     displayError,
		Bot:       bot,
	}

	renderSendForm(w, data)
}

type sendResponse struct {
	Result string `json:"result"`
}

func SendAPIHandler(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	bot, err := models.FindBotByToken(common.GetDB(), token)
	if err != nil {
		common.RespondBadRequest(w, err)
	}

	postParams, err := common.ParseJSONBody(r)
	if err != nil {
		common.RespondBadRequest(w, err)
	}

	number, numberOk := postParams["number"].(string)
	if !numberOk {
		err = errors.New("'number' parameter is required")
		common.RespondBadRequest(w, err)
	}

	message, messageOk := postParams["message"].(string)
	if !messageOk {
		err = errors.New("'message' parameter is required")
		common.RespondBadRequest(w, err)
	}

	if err = common.SendMessage(bot, number, message); err != nil {
		common.RespondServerError(w, err)
	}

	res := &sendResponse{
		Result: "ok",
	}

	common.RespondSuccess(w, res)
}

// Receive messages
type message struct {
	Source    string
	Timestamp string
	Body      string
}

type receiveFormData struct {
	PageTitle string
	Error     string
	Number    string
	Messages  []message
}

func ReceiveFormHandler(w http.ResponseWriter, r *http.Request) {
	bot, err := findBot(r)
	if err != nil {
		return
	}

	data := receiveFormData{
		PageTitle: "Receive",
		Error:     "",
		Number:    bot.Number,
		Messages:  []message{},
	}

	templates := template.Must(template.ParseFiles("views/layouts/main.tmpl", "views/bot/receive.tmpl"))
	templates.ExecuteTemplate(w, "main", data)
}

func ReceiveHandler(w http.ResponseWriter, r *http.Request) {
	bot, err := findBot(r)
	if err != nil {
		return
	}

	err = common.ReceiveMessages(bot)
	if err != nil {
		return
	}
}

// Delete bot
type deleteResponse struct {
	Result string `json:"result"`
}

func DeleteHandler(w http.ResponseWriter, r *http.Request) {
	user, err := common.GetUser(r)
	if err != nil {
		return
	}

	r.ParseForm()
	botID := r.Form.Get("botID")

	bot, err := models.FindBotForUser(common.GetDB(), user.ID, botID)
	if err != nil {
		return
	}

	if err := models.DeleteStore(common.GetDB(), bot.ID); err != nil {
		return
	}

	if err := bot.Delete(common.GetDB()); err != nil {
		return
	}

	http.Redirect(w, r, "/account", http.StatusFound)
}

func findBot(r *http.Request) (models.Bot, error) {
	var bot models.Bot
	user, err := common.GetUser(r)
	if err != nil {
		return bot, err
	}

	botID := chi.URLParam(r, "botID")

	return models.FindBotForUser(common.GetDB(), user.ID, botID)
}
