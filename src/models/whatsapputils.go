package models

import (
	"fmt"
	"strings"

	wa "github.com/Rhymen/go-whatsapp"
)

func ReceiveMessagePreProcessing(h *messageHandler, Info wa.MessageInfo) (con *wa.Conn, err error) {
	con, err = getConnection(h.botID)
	if err != nil {
		return
	}

	mutex.Lock()
	_, exists := h.messages[Info.Id]
	mutex.Unlock()

	if exists {
		err = fmt.Errorf("message (%s) already exists on this handler", Info.Id)
	}
	return
}

// Cria uma mensagem no formato do QuePasa apartir de uma msg do WhatsApp
// Preenche somente as propriedades padrões e comuns a todas as msgs
func CreateQPMessage(Info wa.MessageInfo) (message QPMessage) {
	message = QPMessage{}
	message.ID = Info.Id
	message.Timestamp = Info.Timestamp
	return
}

func (message *QPMessage) FillHeader(Info wa.MessageInfo, con *wa.Conn) {
	contact, ok := con.Store.Contacts[Info.RemoteJid]
	if ok {
		message.Name = contact.Name
	}

	// Enderço correto para onde deve ser devolvida a msg
	message.ReplyTo = Info.RemoteJid

	// Extremidade (pessoa que enviou a msg)
	remoteEndPoint := Info.RemoteJid

	// Mensagem vinda de um grupo
	if strings.HasSuffix(Info.RemoteJid, "@g.us") {
		remoteEndPoint = *Info.Source.Participant
	}

	// con.Info.Wid = Whatsapp que esta processando a msg
	currentWhatsAppBot, _ := CleanPhoneNumber(con.Info.Wid)
	currentWhatsAppBot = "+" + currentWhatsAppBot

	// Destino, indo ou vindo
	remoteEndPoint, _ = CleanPhoneNumber(remoteEndPoint)
	remoteEndPoint = "+" + remoteEndPoint

	if Info.FromMe {
		message.Source = currentWhatsAppBot
		message.Recipient = remoteEndPoint
	} else {
		message.Source = remoteEndPoint
		message.Recipient = currentWhatsAppBot
	}
}
