package models

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
	"time"

	wa "github.com/Rhymen/go-whatsapp"
)

// Ler uma seção já logada e salva no banco de dados
// Pronta para uso
// wid = uinque id do whatsapp, não id do bot
func ReadSession(wid string) (wa.Session, error) {
	var session wa.Session
	store, err := GetStore(GetDB(), wid)
	if err != nil {
		return session, err
	}

	r := bytes.NewReader(store.Data)
	decoder := gob.NewDecoder(r)
	err = decoder.Decode(&session)
	if err != nil {
		return session, err
	}

	return session, nil
}

func WriteSession(wid string, session wa.Session) error {
	_, err := GetOrCreateStore(GetDB(), wid)
	if err != nil {
		return err
	}

	var buff bytes.Buffer
	encoder := gob.NewEncoder(&buff)
	if err = encoder.Encode(session); err != nil {
		return err
	}

	_, err = UpdateStore(GetDB(), wid, buff.Bytes())
	if err != nil {
		return err
	}

	return nil
}

// Cria uma instancia básica de conexão com whatsapp
func CreateConnection() (*wa.Conn, error) {
	con, err := wa.NewConn(30 * time.Second)
	if err != nil {
		return con, err
	}

	con.SetClientName("QuePasa for Link", "QuePasa", "0.4")
	con.SetClientVersion(0, 4, 2088)

	return con, err
}

func SendMessageFromBOT(botID string, recipient string, text string, attachment QPAttachment) (messageID string, err error) {
	recipient = strings.TrimLeft(recipient, "+")

	allowedSuffix := map[string]bool{
		"g.us":           true, // Mensagem para um grupo
		"s.whatsapp.net": true, // Mensagem direta a um usuário
	}

	if strings.ContainsAny(recipient, "@") {
		suffix := strings.Split(recipient, "@")
		if !allowedSuffix[suffix[1]] {
			return messageID, fmt.Errorf("invalid recipient %s", recipient)
		}
	} else {
		return messageID, fmt.Errorf("incomplete recipient (@ .uuu) %s", recipient)
	}

	server, ok := GetServer(botID)
	if !ok {
		err = fmt.Errorf("server not found or not ready")
		return
	}

	// Informações basicas para todo tipo de mensagens
	info := wa.MessageInfo{
		RemoteJid: recipient,
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

		switch attachment.MIME {
		case "audio/ogg", "audio/mpeg":
			{
				ptt := attachment.MIME == "audio/ogg"
				msg := wa.AudioMessage{
					Info:    info,
					Length:  uint32(attachment.Length),
					Type:    attachment.MIME,
					Ptt:     ptt,
					Content: reader,
				}
				messageID, err = server.SendMessage(msg)
			}
		case "image/png", "image/jpeg":
			{
				msg := wa.ImageMessage{
					Info:    info,
					Caption: caption,
					Type:    attachment.MIME,
					Content: reader,
				}
				messageID, err = server.SendMessage(msg)
			}
		default:
			{
				msg := wa.DocumentMessage{
					Info:     info,
					Title:    caption,
					FileName: attachment.FileName,
					Type:     attachment.MIME,
					Content:  reader,
				}
				messageID, err = server.SendMessage(msg)
			}
		}

	} else if len(text) > 0 {
		msg := wa.TextMessage{
			Info: info,
			Text: text,
		}
		messageID, err = server.SendMessage(msg)
	}

	if err != nil {
		log.Printf("(%s) recipient: %s :: error sending message, attachment: %s :: %s", botID, recipient, attachment.MIME, attachment.FileName)
	}

	return
}

// Retrieve messages from the controller, external
func RetrieveMessages(botID string, timestamp string) (messages []QPMessage, err error) {
	searchTimestamp, _ := strconv.ParseUint(timestamp, 10, 64)
	if searchTimestamp == 0 {
		searchTimestamp = 1000000
	}

	server, ok := WhatsAppService.Servers[botID]
	if !ok {
		err = fmt.Errorf("handlers not read yet, please wait")
		return
	}

	messages, err = server.GetMessages(searchTimestamp)
	if err != nil {
		err = fmt.Errorf("msgs not read yet, please wait")
		return
	}

	//mutex.Lock() // travando multi threading
	sort.Sort(ByTimestamp(messages))
	//mutex.Unlock() // destravando multi threading

	return
}

func GetConnection(botID string) (con *wa.Conn, err error) {
	server, ok := WhatsAppService.Servers[botID]
	if !ok {
		err = fmt.Errorf("handlers not read yet, please wait")
		return
	}

	con = server.Connection
	return
}
