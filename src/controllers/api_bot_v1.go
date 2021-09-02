package controllers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/Rhymen/go-whatsapp"
	"github.com/go-chi/chi"
	"github.com/sufficit/sufficit-quepasa-fork/models"
)

// SendAPIHandler renders route "/v1/bot/{token}/send"
func SendAPIHandlerV1(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	bot, err := models.WhatsAppService.DB.Bot.FindByToken(token)
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

	res := &models.QPSendResponse{
		Result: &models.QPSendResult{
			Source:    bot.GetNumber(),
			Recipient: request.Recipient,
			MessageId: messageID,
		},
	}

	respondSuccess(w, res)
}

// ReceiveAPIHandler renders route GET "/v1/bot/{token}/receive"
func ReceiveAPIHandlerV1(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	bot, err := models.WhatsAppService.DB.Bot.FindByToken(token)
	if err != nil {
		respondNotFound(w, fmt.Errorf("Token '%s' not found", token))
		return
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

// InfoAPIHandler renders route GET "/v1/bot/{token}"
func InfoAPIHandlerV1(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	bot, err := models.WhatsAppService.DB.Bot.FindByToken(token)
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

func WebHookAPIHandlerV1(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	bot, err := models.WhatsAppService.DB.Bot.FindByToken(token)
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

	server, ok := models.WhatsAppService.Servers[bot.ID]
	if ok {
		bot = server.Bot
	}

	// Já tratei os parametros
	if models.ENV.IsDevelopment() {
		log.Printf("(%s) Updating Webhook: %s", server.Bot.GetNumber(), p.Url)
	}

	bot.WebHook = p.Url
	// Atualizando banco de dados
	if err := bot.WebHookUpdate(); err != nil {
		return
	}

	respondSuccess(w, bot)
}

// AttachmentHandler renders route POST "/v1/bot/{token}/attachment"
func AttachmentAPIHandlerV1(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	bot, err := models.WhatsAppService.DB.Bot.FindByToken(token)
	if err != nil {
		respondNotFound(w, fmt.Errorf("Token '%s' not found", token))
		return
	}

	// Evitando tentativa de download de anexos sem o bot estar devidamente sincronizado
	if bot.GetStatus() != "ready" {
		respondNotReady(w, fmt.Errorf("bot not ready yet ! try later."))
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
