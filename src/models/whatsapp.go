package models

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"log"
	"sort"
	"strconv"
	"time"

	wa "github.com/Rhymen/go-whatsapp"
	"github.com/Rhymen/go-whatsapp/binary/proto"
	qrcode "github.com/skip2/go-qrcode"
)

func SignIn(botID string, out chan<- []byte) error {
	con, err := wa.NewConn(5 * time.Second)
	if err != nil {
		return err
	}
	con.SetClientName("QuePasa for Link", "QuePasa")

	qr := make(chan string)
	go func() {
		var png []byte
		png, err := qrcode.Encode(<-qr, qrcode.Medium, 256)
		if err != nil {
			log.Println(err)
		}
		encodedPNG := base64.StdEncoding.EncodeToString(png)
		out <- []byte(encodedPNG)
	}()

	session, err := con.Login(qr)
	if err != nil {
		return err
	}

	return writeSession(botID, session)
}

func writeSession(botID string, session wa.Session) error {
	_, err := GetOrCreateStore(GetDB(), botID)
	if err != nil {
		return err
	}

	var buff bytes.Buffer
	encoder := gob.NewEncoder(&buff)
	if err = encoder.Encode(session); err != nil {
		return err
	}

	_, err = UpdateStore(GetDB(), botID, buff.Bytes())
	if err != nil {
		return err
	}

	return nil
}

func readSession(botID string) (wa.Session, error) {
	var session wa.Session
	store, err := GetStore(GetDB(), botID)
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

func SendMessage(botID string, recipient string, message string) (string, error) {
	var messageID string
	con, err := wa.NewConn(10 * time.Second)
	if err != nil {
		return messageID, err
	}

	session, err := readSession(botID)
	if err != nil {
		return messageID, err
	}

	session, err = con.RestoreWithSession(session)
	if err != nil {
		return messageID, err
	}

	if err := writeSession(botID, session); err != nil {
		return messageID, err
	}

	<-time.After(3 * time.Second)

	formattedRecipient := CleanPhoneNumber(recipient)

	msg := wa.TextMessage{
		Info: wa.MessageInfo{
			RemoteJid: formattedRecipient + "@s.whatsapp.net",
		},
		Text: message,
	}

	messageID, err = con.Send(msg)
	if err != nil {
		return messageID, err
	}

	if err := writeSession(botID, session); err != nil {
		return messageID, err
	}

	return messageID, nil
}

//
// ReceiveMessages
//

func ReceiveMessages(botID string, timestamp string) ([]Message, error) {
	messages := []Message{}
	con, err := wa.NewConn(10 * time.Second)
	if err != nil {
		return messages, err
	}

	allUserIDs := make(map[string]bool)
	chatHandler := &chatHandler{con, allUserIDs}
	con.AddHandler(chatHandler)

	session, err := readSession(botID)
	if err != nil {
		return messages, err
	}

	session, err = con.RestoreWithSession(session)
	if err != nil {
		return messages, err
	}

	<-time.After(3 * time.Second)

	con.RemoveHandlers()

	messages, err = fetchMessages(con, chatHandler.userIDs, timestamp)
	if err != nil {
		return messages, err
	}

	session, err = con.Disconnect()
	if err != nil {
		return messages, err
	}

	if err := writeSession(botID, session); err != nil {
		return messages, err
	}

	return messages, nil
}

func loadMessages(con *wa.Conn, userID string, count int) ([]Message, error) {
	messages := []Message{}
	handler := &messageHandler{con, messages}
	con.LoadFullChatHistory(userID, count, time.Millisecond*300, handler)
	con.RemoveHandlers()
	return handler.messages, nil
}

func fetchUserMessages(con *wa.Conn, userID string, timestamp int64) ([]Message, error) {
	messages := []Message{}
	count := 50

	allMessages, err := loadMessages(con, userID, count)
	if err != nil {
		return messages, err
	}

	for _, message := range allMessages {
		if message.Timestamp >= uint64(timestamp) {
			messages = append(messages, message)
		}
	}

	return messages, nil
}

func fetchMessages(con *wa.Conn, userIDs map[string]bool, timestamp string) ([]Message, error) {
	messages := []Message{}
	var t int64
	var err error
	if timestamp != "" {
		t, err = strconv.ParseInt(timestamp, 10, 64)
		if err != nil {
			return messages, err
		}
	}

	for userID, _ := range userIDs {
		if string(userID[0]) == "+" {
			continue
		}
		userMsgs, err := fetchUserMessages(con, userID, t)
		if err != nil {
			return messages, err
		}

		messages = append(messages, userMsgs...)
	}

	sort.Sort(ByTimestamp(messages))

	return messages, nil
}

type chatHandler struct {
	con     *wa.Conn
	userIDs map[string]bool
}

func (h *chatHandler) ShouldCallSynchronously() bool {
	return true
}

func (h *chatHandler) HandleError(err error) {
	if _, ok := err.(*wa.ErrConnectionFailed); ok {
		<-time.After(30 * time.Second)
		err := h.con.Restore()
		if err != nil {
			log.Printf("Restore failed: %v", err)
		}
	} else {
		log.Printf("Chat handler error: %v\n", err)
	}
}

func (h *chatHandler) HandleRawMessage(message *proto.WebMessageInfo) {
	if message != nil && message.Key.RemoteJid != nil {
		userID := *message.Key.RemoteJid
		h.userIDs[userID] = true
	}
}

type messageHandler struct {
	con      *wa.Conn
	messages []Message
}

func (h *messageHandler) ShouldCallSynchronously() bool {
	return true
}

func (h *messageHandler) HandleError(err error) {
	if _, ok := err.(*wa.ErrConnectionFailed); ok {
		<-time.After(30 * time.Second)
		err := h.con.Restore()
		if err != nil {
			log.Printf("Restore failed: %v", err)
		}
	} else {
		log.Printf("Message handler error: %v\n", err)
	}
}

func (h *messageHandler) HandleTextMessage(msg wa.TextMessage) {
	currentUserID := CleanPhoneNumber(h.con.Info.Wid) + "@s.whatsapp.net"
	message := Message{}
	message.ID = msg.Info.Id
	message.Timestamp = msg.Info.Timestamp
	message.Body = msg.Text
	if msg.Info.FromMe {
		message.Source = currentUserID
		message.Recipient = msg.Info.RemoteJid
	} else {
		message.Source = msg.Info.RemoteJid
		message.Recipient = currentUserID
	}

	h.messages = append(h.messages, message)
}
