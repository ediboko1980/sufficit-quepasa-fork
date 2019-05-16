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

//
// Register
//

type registerFormData struct {
	PageTitle    string
	ErrorMessage string
}

func renderRegisterForm(w http.ResponseWriter, data registerFormData) {
	templates := template.Must(template.ParseFiles(
		"views/layouts/main.tmpl",
		"views/bot/register.tmpl"))
	templates.ExecuteTemplate(w, "main", data)
}

// RegisterFormHandler renders route GET "/bot/register"
func RegisterFormHandler(w http.ResponseWriter, r *http.Request) {
	data := registerFormData{
		PageTitle: "Register",
	}
	renderRegisterForm(w, data)
}

// RegisterHandler renders route POST "/bot/register"
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	user, err := common.GetUser(r)
	if err != nil {
		common.RedirectToLogin(w, r)
		return
	}

	data := registerFormData{
		PageTitle: "Register",
	}

	r.ParseForm()
	number := r.Form.Get("number")

	if number == "" {
		data.ErrorMessage = "Number is required"
		renderRegisterForm(w, data)
		return
	}

	bot, err := models.CreateBot(common.GetDB(), user.ID, number)
	if err != nil {
		data.ErrorMessage = err.Error()
		renderRegisterForm(w, data)
		return
	}

	http.Redirect(w, r, "/bot/"+bot.ID+"/verify", http.StatusFound)
}

//
// Cycle
//

// CycleHandler renders route POST "/bot/{botID}/cycle"
func CycleHandler(w http.ResponseWriter, r *http.Request) {
	user, err := common.GetUser(r)
	if err != nil {
		common.RedirectToLogin(w, r)
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

//
// Verify
//

type verifyFormData struct {
	PageTitle    string
	ErrorMessage string
	Bot          models.Bot
	Protocol     string
	Host         string
}

// VerifyFormHandler renders route GET "/bot/{botID}/verify"
func VerifyFormHandler(w http.ResponseWriter, r *http.Request) {
	data := verifyFormData{
		PageTitle: "Verify",
		Protocol:  common.WebSocketProtocol(),
		Host:      r.Host,
	}

	bot, err := findBot(r)
	if err != nil {
		data.ErrorMessage = err.Error()
	} else {
		data.Bot = bot
	}

	templates := template.Must(template.ParseFiles(
		"views/layouts/main.tmpl",
		"views/bot/verify.tmpl",
	))
	templates.ExecuteTemplate(w, "main", data)
}

var upgrader = websocket.Upgrader{}

// VerifyHandler renders route GET "/bot/{botID}/verify/ws"
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
		log.Printf("Sign in error: %v", err)
		err = con.WriteMessage(websocket.TextMessage, []byte("Complete"))
		if err != nil {
			log.Println("Write message error: ", err)
		}
		return
	}

	err = bot.MarkVerified(common.GetDB())
	if err != nil {
		log.Println(err)
	}
	err = con.WriteMessage(websocket.TextMessage, []byte("Complete"))
	if err != nil {
		log.Println("Write message error: ", err)
	}
}

//
// Send
//

type sendFormData struct {
	PageTitle    string
	ErrorMessage string
	Bot          models.Bot
}

func renderSendForm(w http.ResponseWriter, data sendFormData) {
	templates := template.Must(template.ParseFiles("views/layouts/main.tmpl", "views/bot/send.tmpl"))
	templates.ExecuteTemplate(w, "main", data)
}

// SendFormHandler renders route GET "/bot/{botID}/send"
func SendFormHandler(w http.ResponseWriter, r *http.Request) {
	data := sendFormData{
		PageTitle: "Send",
	}

	bot, err := findBot(r)
	if err != nil {
		data.ErrorMessage = err.Error()
		renderSendForm(w, data)
		return
	}

	data.Bot = bot
	renderSendForm(w, data)
}

// SendHandler renders route POST "/bot/{botID}/send"
func SendHandler(w http.ResponseWriter, r *http.Request) {
	data := sendFormData{
		PageTitle: "Send",
	}
	bot, err := findBot(r)
	if err != nil {
		data.ErrorMessage = err.Error()
		renderSendForm(w, data)
		return
	}

	r.ParseForm()
	recipient := r.Form.Get("recipient")
	message := r.Form.Get("message")

	if err = common.SendMessage(bot, recipient, message); err != nil {
		data.ErrorMessage = err.Error()
		renderSendForm(w, data)
		return
	}

	renderSendForm(w, data)
}

type sendResponse struct {
	Result string `json:"result"`
}

// SendAPIHandler renders route "/v1/bot/{token}/send"
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

//
// Receive
//

type message struct {
	Source    string
	Timestamp string
	Body      string
}

type receiveResponse struct {
	Messages []message  `json:"messages"`
	Bot      models.Bot `json:"bot"`
}

type receiveFormData struct {
	PageTitle    string
	ErrorMessage string
	Number       string
	Messages     []message
}

// ReceiveFormHandler renders route GET "/bot/{botID}/receive"
func ReceiveFormHandler(w http.ResponseWriter, r *http.Request) {
	data := receiveFormData{
		PageTitle: "Receive",
		Messages:  []message{},
	}

	bot, err := findBot(r)
	if err != nil {
		data.ErrorMessage = err.Error()
	} else {
		data.Number = bot.Number
	}

	templates := template.Must(template.ParseFiles(
		"views/layouts/main.tmpl",
		"views/bot/receive.tmpl"))
	templates.ExecuteTemplate(w, "main", data)
}

// ReceiveHandler renders route GET "/bot/{botID}/receive/ws"
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

// ReceiveAPIHandler renders route GET "/v1/bot/{token}/receive"
func ReceiveAPIHandler(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	bot, err := models.FindBotByToken(common.GetDB(), token)
	if err != nil {
		common.RespondBadRequest(w, err)
	}

	err = common.ReceiveMessages(bot)
	if err != nil {
		return
	}

	out := receiveResponse{
		Bot: bot,
	}

	common.RespondSuccess(w, out)
}

//
// Info
//

// InfoAPIHandler renders route GET "/v1/bot/{token}"
func InfoAPIHandler(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	bot, err := models.FindBotByToken(common.GetDB(), token)
	if err != nil {
		common.RespondBadRequest(w, err)
	}

	common.RespondSuccess(w, bot)
}

//
// Delete
//

// DeleteHandler renders route POST "/bot/{botID}/delete"
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

//
// Helpers
//

func findBot(r *http.Request) (models.Bot, error) {
	var bot models.Bot
	user, err := common.GetUser(r)
	if err != nil {
		return bot, err
	}

	botID := chi.URLParam(r, "botID")

	return models.FindBotForUser(common.GetDB(), user.ID, botID)
}