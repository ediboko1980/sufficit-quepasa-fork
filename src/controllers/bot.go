package controllers

import (
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/gorilla/websocket"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/sufficit/sufficit-quepasa-fork/models"
)

//
// Metrics
//

var messagesSent = promauto.NewCounter(prometheus.CounterOpts{
	Name: "quepasa_sent_messages_total",
	Help: "Total sent messages",
})

var messageSendErrors = promauto.NewCounter(prometheus.CounterOpts{
	Name: "quepasa_send_message_errors_total",
	Help: "Total message send errors",
})

var messagesReceived = promauto.NewCounter(prometheus.CounterOpts{
	Name: "quepasa_received_messages_total",
	Help: "Total messages received",
})

var messageReceiveErrors = promauto.NewCounter(prometheus.CounterOpts{
	Name: "quepasa_receive_message_errors_total",
	Help: "Total message receive errors",
})

//
// Cycle
//

// CycleHandler renders route POST "/bot/cycle"
func CycleHandler(w http.ResponseWriter, r *http.Request) {
	user, err := models.GetUser(r)
	if err != nil {
		redirectToLogin(w, r)
		return
	}

	r.ParseForm()
	botID := r.Form.Get("botID")
	bot, err := models.WhatsAppService.DB.Bot.FindForUser(user.ID, botID)
	if err != nil {
		return
	}

	err = bot.CycleToken()
	if err != nil {
		return
	}

	http.Redirect(w, r, "/account", http.StatusFound)
}

// DebugHandler renders route POST "/bot/debug"
func DebugHandler(w http.ResponseWriter, r *http.Request) {
	user, err := models.GetUser(r)
	if err != nil {
		redirectToLogin(w, r)
		return
	}

	r.ParseForm()
	botID := r.Form.Get("botID")
	bot, err := models.WhatsAppService.DB.Bot.FindForUser(user.ID, botID)
	if err != nil {
		return
	}

	err = bot.ToggleDevel()
	if err != nil {
		log.Print("Error on toggle devel: ", err)
		return
	}

	http.Redirect(w, r, "/account", http.StatusFound)
}

// ToggleHandler renders route POST "/bot/toggle"
func ToggleHandler(w http.ResponseWriter, r *http.Request) {
	user, err := models.GetUser(r)
	if err != nil {
		redirectToLogin(w, r)
		return
	}

	r.ParseForm()
	botID := r.Form.Get("botID")
	bot, err := models.WhatsAppService.DB.Bot.FindForUser(user.ID, botID)
	if err != nil {
		return
	}

	err = bot.Toggle()
	if err != nil {
		log.Print("error on toggle: ", err)
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
	Bot          models.QPBot
	Protocol     string
	Host         string
}

// VerifyFormHandler renders route GET "/bot/verify"
func VerifyFormHandler(w http.ResponseWriter, r *http.Request) {
	data := verifyFormData{
		PageTitle: "Verify To Add or Update",
		Protocol:  webSocketProtocol(),
		Host:      r.Host,
	}

	templates := template.Must(template.ParseFiles(
		"views/layouts/main.tmpl",
		"views/bot/verify.tmpl",
	))
	templates.ExecuteTemplate(w, "main", data)
}

var upgrader = websocket.Upgrader{}

// VerifyHandler renders route GET "/bot/verify/ws"
func VerifyHandler(w http.ResponseWriter, r *http.Request) {
	user, err := models.GetUser(r)
	if err != nil {
		log.Print("Connection upgrade error (not logged): ", err)
		redirectToLogin(w, r)
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

	// Exibindo código QR
	bot, err := models.SignInWithQRCode(user, out)
	if err != nil {
		err = con.WriteMessage(websocket.TextMessage, []byte("Complete"))
		// Se for timeout não me interessa e volta para tela de contas
		if err != nil {
			log.Printf("error on read qr code: %s", err)
		}

		w.WriteHeader(http.StatusOK)
		return
	}

	log.Printf("(%s) Verificação QRCode confirmada ...", bot.GetNumber())
	err = bot.MarkVerified(true)
	if err != nil {
		log.Println(err)
	}

	go models.WhatsAppService.AppendNewServer(bot)

	err = con.WriteMessage(websocket.TextMessage, []byte("Complete"))
	if err != nil {
		log.Println("Write message error: ", err)
	}

	w.WriteHeader(http.StatusOK)
}

//
// Send
//

type sendFormData struct {
	PageTitle    string
	MessageId    string
	ErrorMessage string
	Bot          models.QPBot
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
// Vindo do formulário de testes
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

	messageID, err := models.SendMessageFromBOT(bot.ID, recipient, message, models.QPAttachment{})
	if err != nil {
		messageSendErrors.Inc()
		data.ErrorMessage = err.Error()
		renderSendForm(w, data)
		return
	}

	data.MessageId = messageID

	messagesSent.Inc()

	renderSendForm(w, data)
}

//
// Receive
//

type receiveResponse struct {
	Messages []models.QPMessage `json:"messages"`
	Bot      models.QPBot       `json:"bot"`
}

type receiveFormData struct {
	PageTitle    string
	ErrorMessage string
	Number       string
	Messages     []models.QPMessage
}

// ReceiveFormHandler renders route GET "/bot/{botID}/receive"
func ReceiveFormHandler(w http.ResponseWriter, r *http.Request) {
	data := receiveFormData{
		PageTitle: "Receive",
	}

	bot, err := findBot(r)
	if err != nil {
		data.ErrorMessage = err.Error()
	} else {
		data.Number = bot.GetNumber()
	}

	// Evitando tentativa de download de anexos sem o bot estar devidamente sincronizado
	if bot.GetStatus() != "ready" {
		respondNotReady(w, fmt.Errorf("bot not ready yet ! try later."))
		return
	}

	queryValues := r.URL.Query()
	timestamp := queryValues.Get("timestamp")

	messages, err := models.RetrieveMessages(bot.ID, timestamp)
	if err != nil {
		messageReceiveErrors.Inc()
		data.ErrorMessage = err.Error()
	}

	data.Messages = messages

	messagesReceived.Add(float64(len(messages)))

	templates := template.Must(template.ParseFiles(
		"views/layouts/main.tmpl",
		"views/bot/receive.tmpl"))
	templates.ExecuteTemplate(w, "main", data)
}

//
// Delete
//

// DeleteHandler renders route POST "/bot/{botID}/delete"
func DeleteHandler(w http.ResponseWriter, r *http.Request) {
	user, err := models.GetUser(r)
	if err != nil {
		return
	}

	r.ParseForm()
	botID := r.Form.Get("botID")

	bot, err := models.WhatsAppService.DB.Bot.FindForUser(user.ID, botID)
	if err != nil {
		return
	}

	if err := models.WhatsAppService.DB.Store.Delete(bot.ID); err != nil {
		return
	}

	if err := bot.Delete(); err != nil {
		return
	}

	http.Redirect(w, r, "/account", http.StatusFound)
}

//
// Helpers
//

func findBot(r *http.Request) (models.QPBot, error) {
	var bot models.QPBot
	user, err := models.GetUser(r)
	if err != nil {
		return bot, err
	}

	botID := chi.URLParam(r, "botID")

	return models.WhatsAppService.DB.Bot.FindForUser(user.ID, botID)
}
