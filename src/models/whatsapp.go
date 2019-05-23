package models

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"log"
	"time"

	wa "github.com/Rhymen/go-whatsapp"
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

func ReceiveMessages(botID string, limit int) ([]Message, error) {
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

	ch := make(chan Message)
	done := make(chan bool)
	con.AddHandler(&waHandler{con, ch})

	go func() {
		for {
			message := <-ch
			messages = append(messages, message)
			if len(messages) >= limit {
				done <- true
			}
		}
	}()

	select {
	case <-done:
		log.Println("Done")
	case <-time.After(10 * time.Second):
		log.Println("Timeout after 10 seconds")
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

type waHandler struct {
	con *wa.Conn
	ch  chan<- Message
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

func (h *waHandler) HandleTextMessage(message wa.TextMessage) {
	msg := Message{
		ID:        message.Info.Id,
		Source:    message.Info.RemoteJid,
		Timestamp: time.Unix(int64(message.Info.Timestamp), 0).Format(time.RFC3339),
		Body:      message.Text,
	}

	h.ch <- msg
}
