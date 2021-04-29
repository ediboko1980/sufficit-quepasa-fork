package models

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	wa "github.com/Rhymen/go-whatsapp"
)

type QPMessageHandler struct {
	Bot         *QPBot
	Synchronous bool
	Server      *QPWhatsAppServer
}

// Essencial
// Unico item realmente necessario para o sistema do whatsapp funcionar
func (h *QPMessageHandler) HandleError(publicError error) {
	if e, ok := publicError.(*wa.ErrConnectionFailed); ok {
		log.Printf("(%s) SUFF ERROR B :: Connection failed, underlying error: %v", h.Server.Bot.Number, e.Err)
		//<-time.After(10 * time.Second)
		//h.Server.Restart()
	} else if strings.Contains(publicError.Error(), "received invalid data") {
		return // ignorando erro conhecido
	} else if strings.Contains(publicError.Error(), "tag 174") {
		log.Printf("(%s) SUFF ERROR D :: Binary decode error, underlying error: %v", h.Server.Bot.Number, publicError)
		//<-time.After(10 * time.Second)
		//h.Server.Restart()
	} else if strings.Contains(publicError.Error(), "code: 1000") {
		log.Printf("(%s) SUFF ERROR H :: %v\n", h.Server.Bot.Number, publicError)
		//<-time.After(10 * time.Second)
		//h.Server.Restart()
	} else {
		log.Printf("SUFF ERROR E :: Message handler error: %v\n", publicError.Error())
	}
}

// Message handler

func (h *QPMessageHandler) HandleJsonMessage(message string) {
	if isDevelopment() {
		fmt.Println("JsonMessage: " + message)
	}
}

func (h *QPMessageHandler) HandleInfoMessage(msg wa.MessageInfo) {
	b, err := json.Marshal(msg)
	if err != nil {
		fmt.Println(err)
		return
	}

	log.Printf("INFO :: %#v\n", string(b))
}

func (h *QPMessageHandler) HandleImageMessage(msg wa.ImageMessage) {
	//con, err := ReceiveMessagePreProcessing(h, msg.Info)
	//if err != nil {
	//	log.Printf("SUFF ERROR G :: ImageMessage error on pre processing received message: %v", err)
	//	return
	//}

	message := CreateQPMessage(msg.Info)
	message.FillHeader(msg.Info, h.Server.Connection)

	//  --> Personalizado para esta seção
	message.Text = "Imagem recebida: " + msg.Type
	message.FillImageAttachment(msg, h.Server.Connection)
	//  <--

	h.Server.AppenMsgToCache(message)
}

func (h *QPMessageHandler) HandleLocationMessage(msg wa.LocationMessage) {
	//con, err := ReceiveMessagePreProcessing(h, msg.Info)
	//if err != nil {
	//	log.Printf("SUFF ERROR G :: LocationMessage error on pre processing received message: %v", err)
	//	return
	//}

	message := CreateQPMessage(msg.Info)
	message.FillHeader(msg.Info, h.Server.Connection)

	//  --> Personalizado para esta seção
	message.Text = "Localização recebida ... "
	//  <--

	h.Server.AppenMsgToCache(message)
}

func (h *QPMessageHandler) HandleLiveLocationMessage(msg wa.LiveLocationMessage) {
	//con, err := ReceiveMessagePreProcessing(h, msg.Info)
	//if err != nil {
	//	log.Printf("SUFF ERROR G :: LiveLocationMessage error on pre processing received message: %v", err)
	//	return
	//}

	message := CreateQPMessage(msg.Info)
	message.FillHeader(msg.Info, h.Server.Connection)

	//  --> Personalizado para esta seção
	message.Text = "Localização em tempo real recebida ... "
	//  <--

	h.Server.AppenMsgToCache(message)
}

func (h *QPMessageHandler) HandleDocumentMessage(msg wa.DocumentMessage) {
	//con, err := ReceiveMessagePreProcessing(h, msg.Info)
	//if err != nil {
	//	log.Printf("SUFF ERROR G :: DocumentMessage error on pre processing received message: %v", err)
	//	return
	//}

	message := CreateQPMessage(msg.Info)
	message.FillHeader(msg.Info, h.Server.Connection)

	//  --> Personalizado para esta seção
	message.Text = "Documento recebido: " + msg.Type + " :: " + msg.FileName
	message.FillDocumentAttachment(msg, h.Server.Connection)
	//  <--

	h.Server.AppenMsgToCache(message)
}

func (h *QPMessageHandler) HandleContactMessage(msg wa.ContactMessage) {
	//con, err := ReceiveMessagePreProcessing(h, msg.Info)
	//if err != nil {
	//	log.Printf("SUFF ERROR G :: ContactMessage error on pre processing received message: %v", err)
	//	return
	//}

	message := CreateQPMessage(msg.Info)
	message.FillHeader(msg.Info, h.Server.Connection)

	//  --> Personalizado para esta seção
	message.Text = "Contato VCARD recebido ... "
	//  <--

	h.Server.AppenMsgToCache(message)
}

func (h *QPMessageHandler) HandleAudioMessage(msg wa.AudioMessage) {
	//con, err := ReceiveMessagePreProcessing(h, msg.Info)
	//if err != nil {
	//	log.Printf("SUFF ERROR G :: AudioMessage error on pre processing received message: %v", err)
	//	return
	//}

	message := CreateQPMessage(msg.Info)
	message.FillHeader(msg.Info, h.Server.Connection)

	//  --> Personalizado para esta seção
	message.Text = "Audio recebido: " + msg.Type
	message.FillAudioAttachment(msg, h.Server.Connection)
	//  <--

	h.Server.AppenMsgToCache(message)
}

func (h *QPMessageHandler) HandleTextMessage(msg wa.TextMessage) {
	//con, err := ReceiveMessagePreProcessing(h, msg.Info)
	//if err != nil {
	//	log.Printf("SUFF ERROR G :: TextMessage error on pre processing received message: %v", err)
	//	return
	//}

	message := CreateQPMessage(msg.Info)
	message.FillHeader(msg.Info, h.Server.Connection)

	//  --> Personalizado para esta seção
	message.Text = msg.Text
	//  <--

	h.Server.AppenMsgToCache(message)
}

// Não sei pra que é utilizado
func (h *QPMessageHandler) ShouldCallSynchronously() bool {
	return h.Synchronous
}
