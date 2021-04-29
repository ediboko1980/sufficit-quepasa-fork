package models

import (
	"encoding/base64"
	"fmt"
	"log"

	"github.com/Rhymen/go-whatsapp"
)

// Cria uma mensagem no formato do QuePasa apartir de uma msg do WhatsApp
// Preenche somente as propriedades padrões e comuns a todas as msgs
func CreateQPMessage(Info whatsapp.MessageInfo) (message QPMessage) {
	message = QPMessage{}
	message.ID = Info.Id
	message.Timestamp = Info.Timestamp
	return
}

func (message *QPMessage) FillHeader(Info whatsapp.MessageInfo, con *whatsapp.Conn) (err error) {

	if con == nil {
		err = fmt.Errorf("nil connection")
		log.Print(err)
		return
	}

	if con.Info == nil {
		err = fmt.Errorf("nil connection information")
		log.Print(err)
		return
	}

	// Fui eu quem enviou a msg ?
	message.FromMe = Info.FromMe

	// Controlador, whatsapp gerenciador
	message.Controller.ID = con.Info.Wid
	message.Controller.Phone = getPhone(con.Info.Wid)
	message.Controller.Title = getTitle(con.Store, con.Info.Wid)

	// Endereço correto para onde deve ser devolvida a msg
	message.ReplyTo.ID = Info.RemoteJid
	message.ReplyTo.Phone = getPhone(Info.RemoteJid)
	message.ReplyTo.Title = getTitle(con.Store, Info.RemoteJid)

	// Pessoa que enviou a msg dentro de um grupo
	if Info.Source.Participant != nil {
		message.Participant.ID = *Info.Source.Participant
		message.Participant.Phone = getPhone(*Info.Source.Participant)
		message.Participant.Title = getTitle(con.Store, *Info.Source.Participant)
	}

	return
}

func (message *QPMessage) FillAudioAttachment(msg whatsapp.AudioMessage, con *whatsapp.Conn) {
	getKey := msg.Info.Source.Message.AudioMessage.MediaKey
	getUrl := *msg.Info.Source.Message.AudioMessage.Url
	getLength := *msg.Info.Source.Message.AudioMessage.FileLength
	getMIME := *msg.Info.Source.Message.AudioMessage.Mimetype

	message.Attachment = QPAttachment{
		B64MediaKey: base64.StdEncoding.EncodeToString(getKey),
		Url:         getUrl,
		Length:      int(getLength),
		MIME:        getMIME,
	}
}

func (message *QPMessage) FillDocumentAttachment(msg whatsapp.DocumentMessage, con *whatsapp.Conn) {
	innerMSG := msg.Info.Source.Message.DocumentMessage
	//filename := &innerMSG.FileName
	message.Attachment = QPAttachment{
		B64MediaKey: base64.StdEncoding.EncodeToString(innerMSG.MediaKey),
		Url:         *innerMSG.Url,
		Length:      int(*innerMSG.FileLength),
		MIME:        *innerMSG.Mimetype,
		//FileName:    &filename,
	}
}

func (message *QPMessage) FillImageAttachment(msg whatsapp.ImageMessage, con *whatsapp.Conn) {
	if msg.Info.Source.Message.ImageMessage.Url == nil {
		// Aconteceu na primeira vez, quando cadastrei o número de whatsapp errado
		log.Println("erro on filling image attachement, url not avail")
		return
	}

	getKey := msg.Info.Source.Message.ImageMessage.MediaKey
	getUrl := *msg.Info.Source.Message.ImageMessage.Url
	getLength := *msg.Info.Source.Message.ImageMessage.FileLength
	getMIME := *msg.Info.Source.Message.ImageMessage.Mimetype

	message.Attachment = QPAttachment{
		B64MediaKey: base64.StdEncoding.EncodeToString(getKey),
		Url:         getUrl,
		Length:      int(getLength),
		MIME:        getMIME,
	}
}

func getPhone(textPhone string) string {
	var result string
	phone, err := CleanPhoneNumber(textPhone)
	if err == nil {
		result = "+" + phone
	}
	return result
}

// Retorna algum titulo válido apartir de um jid
func getTitle(store *whatsapp.Store, jid string) string {
	var result string
	contact, ok := store.Contacts[jid]
	if ok {
		result = getContactTitle(contact)
	}
	return result
}

// Retorna algum titulo válido apartir de um contato do whatsapp
func getContactTitle(contact whatsapp.Contact) string {
	var result string
	result = contact.Name
	if len(result) == 0 {
		result = contact.Notify
		if len(result) == 0 {
			result = contact.Short
		}
	}
	return result
}
