package library

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"log"
	"strings"

	"github.com/Rhymen/go-whatsapp"
	"github.com/sufficit/sufficit-quepasa-fork/models"
)

func SendValidate(botID string, recipient string) error {
	recipient = strings.TrimLeft(recipient, "+")
	allowedSuffix := map[string]bool{
		"g.us":           true, // Mensagem para um grupo
		"s.whatsapp.net": true, // Mensagem direta a um usuário
	}

	if strings.ContainsAny(recipient, "@") {
		suffix := strings.Split(recipient, "@")
		if !allowedSuffix[suffix[1]] {
			return fmt.Errorf("invalid recipient %s", recipient)
		}
	} else {
		return fmt.Errorf("incomplete recipient (@ .uuu) %s", recipient)
	}
	return nil
}

func SendTextMessage(botID string, recipient string, text string) (response models.QPSendResponseV2, err error) {
	err = SendValidate(botID, recipient)
	if err != nil {
		return
	}

	server, ok := models.GetServer(botID)
	if !ok {
		err = fmt.Errorf("server not found or not ready")
		return
	}

	response.Chat.ID = recipient
	response.Chat.UserName = recipient
	response.From.ID = server.Bot.ID
	response.From.UserName = server.Bot.GetNumber()

	// Informações basicas para todo tipo de mensagens
	info := whatsapp.MessageInfo{
		RemoteJid: recipient,
	}

	if server.IsDevelopment() {
		log.Printf("(%s)(DEV) Sending msg from bot :: %s :: %s", server.Bot.GetNumber(), recipient, text)
	}

	// log.Printf("sending message from bot: %s :: to recipient: %s", botID, recipient)
	if len(text) > 0 {
		msg := whatsapp.TextMessage{
			Info: info,
			Text: text,
		}
		response.ID, err = server.SendMessage(msg)
	} else {
		err = fmt.Errorf("invalid text length")
	}

	if err != nil {
		log.Printf("(%s) recipient: %s :: error sending text message", server.Bot.GetNumber(), recipient)
	}

	return
}

func SendDocumentMessage(botID string, recipient string, attachment models.QPAttachment) (response models.QPSendResponseV2, err error) {
	err = SendValidate(botID, recipient)
	if err != nil {
		return
	}

	server, ok := models.GetServer(botID)
	if !ok {
		err = fmt.Errorf("server not found or not ready")
		return
	}

	response.Chat.ID = recipient
	response.Chat.UserName = recipient
	response.From.ID = server.Bot.ID
	response.From.UserName = server.Bot.GetNumber()

	// Informações basicas para todo tipo de mensagens
	info := whatsapp.MessageInfo{
		RemoteJid: recipient,
	}

	if server.IsDevelopment() {
		log.Printf("(%s)(DEV) Sending document from bot :: %s", server.Bot.GetNumber(), recipient)
	}

	// log.Printf("sending message from bot: %s :: to recipient: %s", botID, recipient)
	if attachment.Length > 0 {
		var data []byte
		data, err = base64.StdEncoding.DecodeString(attachment.Base64)
		if err != nil {
			return
		}

		// Definindo leitor de bytes do arquivo
		// Futuramente fazer download de uma URL para permitir tamanhos maiores
		reader := bytes.NewReader(data)

		caption := attachment.FileName
		if idx := strings.IndexByte(caption, '.'); idx >= 0 {
			caption = caption[:idx]
		}

		switch attachment.WAMediaType() {
		case whatsapp.MediaAudio:
			{
				ptt := strings.HasPrefix(attachment.MIME, "audio/ogg")
				msg := whatsapp.AudioMessage{
					Info:    info,
					Length:  uint32(attachment.Length),
					Type:    attachment.MIME,
					Ptt:     ptt,
					Content: reader,
				}
				response.ID, err = server.SendMessage(msg)
			}
		case whatsapp.MediaImage:
			{
				msg := whatsapp.ImageMessage{
					Info:    info,
					Caption: caption,
					Type:    attachment.MIME,
					Content: reader,
				}
				response.ID, err = server.SendMessage(msg)
			}
		default:
			{
				msg := whatsapp.DocumentMessage{
					Info:     info,
					Title:    caption,
					FileName: attachment.FileName,
					Type:     attachment.MIME,
					Content:  reader,
				}
				response.ID, err = server.SendMessage(msg)
			}
		}

	} else {
		err = fmt.Errorf("invalid document length")
	}

	if err != nil {
		log.Printf("(%s) recipient: %s :: error sending message, attachment: %s :: %s", server.Bot.GetNumber(), recipient, attachment.MIME, attachment.FileName)
	}

	return
}
