package models

import (
	"fmt"

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

	message.ReplyTo = Info.RemoteJid
	//log.Printf("con.Info.Wid: %s :: contact.Name: %s :: RemoteJid: %s", con.Info.Wid, contact.Name, Info.RemoteJid)

	//currentUserID, _ := CleanPhoneNumber(con.Info.Wid)
	//currentUserID = currentUserID + "@s.whatsapp.net"
	if Info.FromMe {
		message.Source = con.Info.Wid
		message.Recipient = Info.RemoteJid
	} else {
		message.Source = Info.RemoteJid
		message.Recipient = con.Info.Wid
	}
}
