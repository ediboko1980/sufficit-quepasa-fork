package models

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"errors"
	"fmt"
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

func SendMessage(botID string, recipient string, message string) error {
	con, err := wa.NewConn(10 * time.Second)
	if err != nil {
		return err
	}

	session, err := readSession(botID)
	if err != nil {
		return err
	}

	session, err = con.RestoreWithSession(session)
	if err != nil {
		return err
	}

	if err := writeSession(botID, session); err != nil {
		return err
	}

	<-time.After(3 * time.Second)

	formattedRecipient := CleanPhoneNumber(recipient)

	msg := wa.TextMessage{
		Info: wa.MessageInfo{
			RemoteJid: formattedRecipient + "@s.whatsapp.net",
		},
		Text: message,
	}

	_, err = con.Send(msg)
	if err != nil {
		return err
	}

	if err := writeSession(botID, session); err != nil {
		return err
	}

	return nil
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

	session, err := readSession(botID)
	if err != nil {
		return messages, err
	}

	session, err = con.RestoreWithSession(session)
	if err != nil {
		return messages, err
	}

	userIDs, err := fetchUserIDs(con)
	if err != nil {
		return messages, err
	}

	messages, err = fetchMessages(con, userIDs, timestamp)
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

var ErrServerRespondedWith401 = errors.New("server responded with 401")

func loadMessages(con *wa.Conn, userID string, count int) ([]interface{}, error) {
	var messages []interface{}
	node, err := con.LoadMessages(userID, "", count)

	if err != nil && err == wa.ErrServerRespondedWith404 {
		return messages, nil
	} else if err != nil && err.Error() == ErrServerRespondedWith401.Error() {
		return messages, nil
	} else if err != nil {
		return nil, err
	}

	messages, ok := node.Content.([]interface{})
	if !ok {
		return nil, fmt.Errorf("Unexpected message content")
	}

	return messages, nil
}

func parseWebMessage(rawMessage interface{}) (Message, error) {
	message := Message{}
	webMsg, ok := rawMessage.(*proto.WebMessageInfo)
	if !ok {
		return message, fmt.Errorf("Unexpected message content")
	}

	message.ID = *webMsg.Key.Id
	message.Source = *webMsg.Key.RemoteJid
	message.Timestamp = int64(*webMsg.MessageTimestamp)

	if webMsg.Message != nil {
		if webMsg.Message.Conversation != nil {
			message.Body = *webMsg.Message.Conversation
		} else if webMsg.Message.ExtendedTextMessage != nil {
			message.Body = *webMsg.Message.ExtendedTextMessage.Text
		}
	}

	return message, nil
}

func fetchUserMessages(con *wa.Conn, userID string, timestamp int64) ([]Message, error) {
	messages := []Message{}
	count := 5
	oldestTimestamp := time.Now().Unix()
	foundOldMessage := false
	noOlderMessages := false

	for {
		rawMessages, err := loadMessages(con, userID, count)
		if err != nil {
			return messages, err
		}

		noOlderMessages = count > len(rawMessages)

		for _, rm := range rawMessages {
			message, err := parseWebMessage(rm)
			if err != nil {
				return messages, err
			}

			if message.Timestamp < oldestTimestamp {
				oldestTimestamp = message.Timestamp
			}

			if message.Timestamp >= timestamp {
				messages = append(messages, message)
			}
		}

		count = count * 2

		// break if we found a message older than the timestamp or
		// if the number of messages returned is less than the limit
		// otherwise double the number of messages fetched and try again
		if foundOldMessage || noOlderMessages || count > 40 {
			break
		} else {
			messages = []Message{}
		}
	}

	return messages, nil
}

func fetchUserIDs(con *wa.Conn) (map[string]bool, error) {
	userIDs := make(map[string]bool)
	ch := make(chan string)
	done := make(chan bool)

	con.AddHandler(&waHandler{con, ch})

	go func() {
		for {
			select {
			case userID := <-ch:
				userIDs[userID] = true
			case <-time.After(2 * time.Second):
				done <- true
			}
		}
	}()

	<-done

	con.RemoveHandlers()

	return userIDs, nil
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
		userMsgs, err := fetchUserMessages(con, userID, t)
		if err != nil {
			return messages, err
		}

		messages = append(messages, userMsgs...)
	}

	sort.Sort(ByTimestamp(messages))

	return messages, nil
}

type waHandler struct {
	con *wa.Conn
	ch  chan<- string
}

func (h *waHandler) HandleError(err error) {
	if _, ok := err.(*wa.ErrConnectionFailed); ok {
		<-time.After(30 * time.Second)
		err := h.con.Restore()
		if err != nil {
			log.Fatalf("Restore failed: %v", err)
		}
	} else {
		log.Printf("Handler error: %v\n", err)
	}
}

func (h *waHandler) HandleRawMessage(message *proto.WebMessageInfo) {
	jid := *message.Key.RemoteJid
	h.ch <- jid
}
