package models

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	whatsapp "github.com/Rhymen/go-whatsapp"
)

type QPMessageHandler struct {
	Bot         *QPBot
	Synchronous bool
	Server      *QPWhatsAppServer
}

// Essencial
// Unico item realmente necessario para o sistema do whatsapp funcionar
// Trata qualquer erro que influêncie no recebimento de msgs
func (h *QPMessageHandler) HandleError(publicError error) {
	if e, ok := publicError.(*whatsapp.ErrConnectionFailed); ok {
		// Erros comuns de desconexão por qualquer motivo aleatório
		if strings.Contains(e.Err.Error(), "close 1006") {
			// 1006 falha no websocket, informações inválidas, provavelmente baixa qualidade de internet no celular
			log.Printf("(%s) Websocket corrupted (probably poor connection quality), should restart ...", h.Server.Bot.GetNumber())
			go h.Server.Restart()
		} else if strings.Contains(e.Err.Error(), "connection reset by peer") {
			// Possível falha no keep alive, muito tempo sem tráfego
			log.Printf("(%s) Websocket corrupted (probably inactive), should restart ...", h.Server.Bot.GetNumber())
			go h.Server.Restart()
		} else {
			log.Printf("(%s) SUFF ERROR B :: %v", h.Server.Bot.GetNumber(), e.Err)
		}
		return
	} else if strings.Contains(publicError.Error(), "code: 1000") {
		// Desconexão forçado é algum evento iniciado pelo whatsapp
		log.Printf("(%s) Desconexão forçada pelo whatsapp, code: 1000", h.Server.Bot.GetNumber())
		// Se houve desconexão, reseta
		go h.Server.Restart()
		return
	} else if strings.Contains(publicError.Error(), "close 1006") {
		// Desconexão forçado é algum evento iniciado pelo whatsapp
		log.Printf("(%s) Desconexão por falha no websocket, code: 1006, iremos reiniciar automaticamente", h.Server.Bot.GetNumber())
		// Se houve desconexão, reseta
		go h.Server.Restart()
		return
	} else if strings.Contains(publicError.Error(), "keepAlive failed") {
		if *h.Server.Status == "ready" {
			*h.Server.Status = "unreachable"
		}

		// Se houve desconexão, reseta
		log.Printf("(%s) Keep alive failed, waiting ...", h.Server.Bot.GetNumber())
		// go h.Server.Restart()
		return
	} else if strings.Contains(publicError.Error(), "server closed connection") {
		// Se houve desconexão, reseta
		log.Printf("(%s) Server closed connection, restarting ...", h.Server.Bot.GetNumber())
		go h.Server.Restart()
		return
	} else if strings.Contains(publicError.Error(), "message type not implemented") {
		// Ignorando, nova implementação com Handlers não criados ainda
		return
	} else {
		log.Printf("(%s) SUFF ERROR D :: %s", h.Server.Bot.GetNumber(), publicError)
	}
}

func (h *QPMessageHandler) HandleJsonMessage(msgString string) {
	var waJsonMessage WhatsAppJsonMessage
	err := json.Unmarshal([]byte(msgString), &waJsonMessage)
	if err == nil {
		if waJsonMessage.Cmd.Type == "disconnect" {
			// Restarting because an order of whatsapp
			log.Printf("(%s) Restart Order by: %s", h.Bot.GetNumber(), waJsonMessage.Cmd.Kind)
			go h.Server.Restart()
		} else {
			if h.Server.IsDevelopment() {
				log.Printf("(%s)(DEV) JSON Unmarshal string :: %s", h.Server.Bot.GetNumber(), msgString)
				log.Printf("(%s)(DEV) JSON Unmarshal :: %s", h.Server.Bot.GetNumber(), waJsonMessage)
			}
		}
	} else {
		if h.Server.IsDevelopment() && ENV.DEBUGJsonMessages() {
			log.Printf("(%s)(DEV) JSON :: %s", h.Server.Bot.GetNumber(), msgString)
		}
	}
}

/// Atualizando informações sobre a bateria
func (h *QPMessageHandler) HandleBatteryMessage(msg whatsapp.BatteryMessage) {
	h.Server.Battery.Timestamp = time.Now()
	h.Server.Battery.Plugged = msg.Plugged
	h.Server.Battery.Percentage = msg.Percentage
	h.Server.Battery.Powersave = msg.Powersave
}

func (h *QPMessageHandler) HandleNewContact(contact whatsapp.Contact) {
	if h.Server.IsDevelopment() {
		log.Printf("(%s)(DEV) NEWCONTACT :: %#v\n", h.Bot.GetNumber(), contact)
	}
}

func (h *QPMessageHandler) HandleInfoMessage(msg whatsapp.MessageInfo) {
	if h.Server.IsDevelopment() {
		b, err := json.Marshal(msg)
		if err != nil {
			fmt.Println(err)
			return
		}

		log.Printf("(%s)(DEV) INFO :: %#v\n", h.Bot.GetNumber(), string(b))
	}
}

func (h *QPMessageHandler) HandleImageMessage(msg whatsapp.ImageMessage) {
	message := CreateQPMessage(msg.Info)
	message.FillHeader(msg.Info, h.Server)

	//  --> Personalizado para esta seção
	message.Text = "Imagem recebida: " + msg.Type
	message.FillImageAttachment(msg, h.Server.Connection)
	//  <--

	h.Server.AppenMsgToCache(message)
}

func (h *QPMessageHandler) HandleLocationMessage(msg whatsapp.LocationMessage) {
	message := CreateQPMessage(msg.Info)
	message.FillHeader(msg.Info, h.Server)

	//  --> Personalizado para esta seção
	message.Text = "Localização recebida ... "
	//  <--

	h.Server.AppenMsgToCache(message)
}

func (h *QPMessageHandler) HandleLiveLocationMessage(msg whatsapp.LiveLocationMessage) {
	message := CreateQPMessage(msg.Info)
	message.FillHeader(msg.Info, h.Server)

	//  --> Personalizado para esta seção
	message.Text = "Localização em tempo real recebida ... "
	//  <--

	h.Server.AppenMsgToCache(message)
}

func (h *QPMessageHandler) HandleDocumentMessage(msg whatsapp.DocumentMessage) {
	message := CreateQPMessage(msg.Info)
	message.FillHeader(msg.Info, h.Server)

	//  --> Personalizado para esta seção
	innerMSG := msg.Info.Source.Message.DocumentMessage
	message.Text = "Documento recebido: " + msg.Type + " :: " + *innerMSG.Mimetype + " :: " + msg.FileName

	message.FillDocumentAttachment(msg, h.Server.Connection)
	//  <--

	h.Server.AppenMsgToCache(message)
}

func (h *QPMessageHandler) HandleContactMessage(msg whatsapp.ContactMessage) {
	message := CreateQPMessage(msg.Info)
	message.FillHeader(msg.Info, h.Server)

	//  --> Personalizado para esta seção
	message.Text = "Contato VCARD recebido ... "
	//  <--

	h.Server.AppenMsgToCache(message)
}

func (h *QPMessageHandler) HandleAudioMessage(msg whatsapp.AudioMessage) {
	message := CreateQPMessage(msg.Info)
	message.FillHeader(msg.Info, h.Server)

	//  --> Personalizado para esta seção
	message.Text = "Audio recebido: " + msg.Type
	message.FillAudioAttachment(msg, h.Server.Connection)
	//  <--

	h.Server.AppenMsgToCache(message)
}

func (h *QPMessageHandler) HandleTextMessage(msg whatsapp.TextMessage) {
	message := CreateQPMessage(msg.Info)
	message.FillHeader(msg.Info, h.Server)

	//  --> Personalizado para esta seção
	message.Text = msg.Text
	//  <--

	h.Server.AppenMsgToCache(message)
}

// Se chamado em modo síncrono, mantém os handlers ativos
func (h *QPMessageHandler) ShouldCallSynchronously() bool {
	return h.Synchronous
}
