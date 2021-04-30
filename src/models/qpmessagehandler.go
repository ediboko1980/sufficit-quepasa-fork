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
		log.Printf("(%s) SUFF ERROR B :: %v", h.Server.Bot.GetNumber(), e.Err)
	} else if strings.Contains(publicError.Error(), "code: 1000") {
		// Desconexão forçado é algum evento iniciado pelo whatsapp
		log.Printf("(%s) Desconexão forçada pelo whatsapp, code: 1000", h.Server.Bot.GetNumber())
		// Se houve desconexão, reseta
		h.Server.Restart()
		return
	} else {
		log.Printf("(%s) SUFF ERROR D :: %s", h.Server.Bot.GetNumber(), publicError)
	}

	// Tratando erros individualmente
	if strings.Contains(publicError.Error(), "keepAlive failed") {
		// Se houve desconexão, reseta
		h.Server.Restart()
		return
	}

	if strings.Contains(publicError.Error(), "server closed connection") {
		// Se houve desconexão, reseta
		h.Server.Restart()
		return
	}
}

// Message handler

func (h *QPMessageHandler) HandleJsonMessage(msgString string) {

	// mensagem de desconexão, o número de whatsapp for removido da lista de permitidos para whatsapp web
	//JsonMessage: ["Cmd",{"type":"disconnect","kind":"replaced"}]

	var waJsonMessage WhatsAppJsonMessage
	err := json.Unmarshal([]byte(msgString), &waJsonMessage)
	if err != nil {
		if waJsonMessage.Cmd.Type == "disconnect" {
			// Restarting because an order of whatsapp
			log.Printf("(%s) Restart Order by: %s", h.Bot.GetNumber(), waJsonMessage.Cmd.Kind)
			h.Server.Restart()
		}
	} else {
		if isDevelopment() {
			fmt.Println("JsonMessage: " + msgString)
		}
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

	if h.Server.Connection.Info == nil {
		log.Print("nil connection information on text msg")
		return
	}

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
