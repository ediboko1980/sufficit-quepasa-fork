package common

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	wa "github.com/Rhymen/go-whatsapp"
	qrcode "github.com/skip2/go-qrcode"
	"gitlab.com/digiresilience/link/quepasa/models"
)

func SignIn(bot models.Bot, out chan<- []byte) error {
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

	return writeSession(bot.ID, session)
}

func writeSession(botID string, session wa.Session) error {
	_, err := models.GetOrCreateStore(GetDB(), botID)
	if err != nil {
		return err
	}

	var buff bytes.Buffer
	encoder := gob.NewEncoder(&buff)
	if err = encoder.Encode(session); err != nil {
		return err
	}

	_, err = models.UpdateStore(GetDB(), botID, buff.Bytes())
	if err != nil {
		return err
	}

	return nil
}

func readSession(botID string) (wa.Session, error) {
	var session wa.Session
	store, err := models.GetStore(GetDB(), botID)
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

func SendMessage(bot models.Bot, recipient string, message string) error {
	con, err := wa.NewConn(10 * time.Second)
	if err != nil {
		return err
	}

	session, err := readSession(bot.ID)
	if err != nil {
		return err
	}

	session, err = con.RestoreWithSession(session)
	if err != nil {
		return err
	}

	if err := writeSession(bot.ID, session); err != nil {
		return err
	}

	<-time.After(3 * time.Second)

	msg := wa.TextMessage{
		Info: wa.MessageInfo{
			RemoteJid: recipient + "@s.whatsapp.net",
		},
		Text: message,
	}

	_, err = con.Send(msg)
	if err != nil {
		return err
	}

	if err := writeSession(bot.ID, session); err != nil {
		return err
	}

	return nil
}

func ReceiveMessages(bot models.Bot) error {
	con, err := wa.NewConn(10 * time.Second)
	if err != nil {
		return err
	}

	con.AddHandler(&waHandler{con})

	session, err := readSession(bot.ID)
	if err != nil {
		return err
	}

	session, err = con.RestoreWithSession(session)
	if err != nil {
		return err
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	session, err = con.Disconnect()
	if err != nil {
		return err
	}

	if err := writeSession(bot.ID, session); err != nil {
		return err
	}

	return nil
}

type waHandler struct {
	con *wa.Conn
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

func (*waHandler) HandleTextMessage(message wa.TextMessage) {
	log.Printf("%v\n%v\n%v\n%v\n%v\n\n", message.Info.Timestamp, message.Info.Id, message.Info.RemoteJid, message.Info.QuotedMessageID, message.Text)
}
