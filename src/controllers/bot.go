package controllers

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/Rhymen/go-whatsapp"
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

// CycleHandler renders route POST "/bot/{botID}/cycle"
func CycleHandler(w http.ResponseWriter, r *http.Request) {
	user, err := models.GetUser(r)
	if err != nil {
		redirectToLogin(w, r)
		return
	}

	r.ParseForm()
	botID := r.Form.Get("botID")
	bot, err := models.FindBotForUser(models.GetDB(), user.ID, botID)
	if err != nil {
		return
	}

	err = bot.CycleToken(models.GetDB())
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
	err = bot.MarkVerified(models.GetDB(), true)
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

type sentMessage struct {
	Source    string `json:"source"`
	Recipient string `json:"recipient"`
	MessageId string `json:"messageId"`
}

type sendResponse struct {
	Result *sentMessage `json:"result"`
}

// SendAPIHandler renders route "/v1/bot/{token}/send"
func SendAPIHandler(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	bot, err := models.FindBotByToken(models.GetDB(), token)
	if err != nil {
		respondNotFound(w, fmt.Errorf("Token '%s' not found", token))
		return
	}

	// Declare a new Person struct.
	var request models.QPSendRequest

	// Try to decode the request body into the struct. If there is an error,
	// respond to the client with the error message and a 400 status code.
	err = json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		respondServerError(bot, w, err)
		return
	}

	messageID, err := models.SendMessageFromBOT(bot.ID, request.Recipient, request.Message, request.Attachment)
	if err != nil {
		messageSendErrors.Inc()
		respondServerError(bot, w, err)
		return
	}

	messagesSent.Inc()

	res := &sendResponse{
		Result: &sentMessage{
			Source:    bot.GetNumber(),
			Recipient: request.Recipient,
			MessageId: messageID,
		},
	}

	respondSuccess(w, res)
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

// ReceiveAPIHandler renders route GET "/v1/bot/{token}/receive"
func ReceiveAPIHandler(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	bot, err := models.FindBotByToken(models.GetDB(), token)
	if err != nil {
		respondNotFound(w, fmt.Errorf("Token '%s' not found", token))
		return
	}

	queryValues := r.URL.Query()
	timestamp := queryValues.Get("timestamp")

	messages, err := models.RetrieveMessages(bot.ID, timestamp)
	if err != nil {
		messageReceiveErrors.Inc()
		respondServerError(bot, w, err)
		return
	}

	messagesReceived.Add(float64(len(messages)))

	out := receiveResponse{
		Bot:      bot,
		Messages: messages,
	}

	respondSuccess(w, out)
}

//
// Info
//

// InfoAPIHandler renders route GET "/v1/bot/{token}"
func InfoAPIHandler(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	bot, err := models.FindBotByToken(models.GetDB(), token)
	if err != nil {
		respondNotFound(w, fmt.Errorf("Token '%s' not found", token))
		return
	}

	var ep models.QPEndPoint
	ep.ID = bot.ID
	ep.Phone = bot.GetNumber()
	if bot.Verified {
		ep.Status = "verified"
	} else {
		ep.Status = "unverified"
	}

	respondSuccess(w, ep)
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

	bot, err := models.FindBotForUser(models.GetDB(), user.ID, botID)
	if err != nil {
		return
	}

	if err := models.DeleteStore(models.GetDB(), bot.ID); err != nil {
		return
	}

	if err := bot.Delete(models.GetDB()); err != nil {
		return
	}

	http.Redirect(w, r, "/account", http.StatusFound)
}

// SUFFICIT -------------------

func WebHookHandler(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	bot, err := models.FindBotByToken(models.GetDB(), token)
	if err != nil {
		respondNotFound(w, fmt.Errorf("Token '%s' not found on WebHookHandler", token))
		return
	}

	// Declare a new Person struct.
	var p models.QPReqWebHook

	// Try to decode the request body into the struct. If there is an error,
	// respond to the client with the error message and a 400 status code.
	err = json.NewDecoder(r.Body).Decode(&p)
	if err != nil {
		respondServerError(bot, w, err)
	}

	log.Printf("UPDATING WEBHOOK :: %#v :: %s\n", bot.ID, p.Url)

	bot.WebHook = p.Url
	if err := bot.WebHookUpdate(models.GetDB()); err != nil {
		return
	}

	respondSuccess(w, bot)
}

// AttachmentHandler renders route POST "/v1/bot/{token}/attachment"
func AttachmentHandler(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	bot, err := models.FindBotByToken(models.GetDB(), token)
	if err != nil {
		respondNotFound(w, fmt.Errorf("Token '%s' not found", token))
		return
	}

	// Declare a new Person struct.
	var p models.QPAttachment

	// Try to decode the request body into the struct. If there is an error,
	// respond to the client with the error message and a 400 status code.
	err = json.NewDecoder(r.Body).Decode(&p)
	if err != nil {
		respondServerError(bot, w, err)
	}

	mediaKey, err := p.MediaKey()
	if err != nil {
		respondServerError(bot, w, err)
		return
	}

	data, err := whatsapp.Download(p.Url, mediaKey, p.WAMediaType(), p.Length)
	if err != nil {
		// se for  "invalid media hmac" é bem provavel que seja de outra conexão
		// só é possivel baixar pela url sendo da mesma conexão
		respondServerError(bot, w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", p.MIME)
	w.Write(data)
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

	return models.FindBotForUser(models.GetDB(), user.ID, botID)
}
