package models

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	wa "github.com/Rhymen/go-whatsapp"
	qrcode "github.com/skip2/go-qrcode"
)

type WhatsAppServer struct {
	connections map[string]*wa.Conn
	handlers    map[string]*messageHandler
}

var server *WhatsAppServer

type messageHandler struct {
	botID       string
	userIDs     map[string]bool
	messages    map[string]QPMessage
	synchronous bool
}

//
// Start
//
func StartServer() error {
	log.Println("Starting WhatsApp server")

	connections := make(map[string]*wa.Conn)
	handlers := make(map[string]*messageHandler)
	server = &WhatsAppServer{connections, handlers}

	return startHandlers()
}

func RestartServer() {
	log.Println("Restarting")

	for _, con := range server.connections {
		con.RemoveHandlers()
		con.Disconnect()
	}
	server = nil
	err := StartServer()
	if err != nil {
		log.Printf("SUFF ERROR F :: Starting service after restart ... %s:", err)
	}
}

func startHandlers() error {
	bots, err := FindAllBots(GetDB())
	if err != nil {
		return err
	}

	for _, bot := range bots {
		log.Printf("(%s) :: Adding message handlers for %s", bot.ID, bot.Number)

		err = startHandler(bot.ID, bot.Token)
		if err != nil {
			return err
		}
	}

	return nil
}

func startHandler(botID string, botToken string) error {
	con, err := createConnection()
	if err != nil {
		return err
	}

	logPrefix := botToken
	server.connections[botID] = con

	userIDs := make(map[string]bool)
	messages := make(map[string]QPMessage)
	startupHandler := &messageHandler{botID, userIDs, messages, true}
	con.AddHandler(startupHandler)

	session, err := readSession(botID)
	if err != nil {
		return err
	}

	session, err = con.RestoreWithSession(session)
	if err != nil {
		return err
	}

	<-time.After(3 * time.Second)

	if err := writeSession(botID, session); err != nil {
		return err
	}

	con.RemoveHandlers()

	log.Printf("(%s) :: Fetching initial messages", logPrefix)
	initialMessages, err := fetchMessages(con, botID, startupHandler.userIDs)
	if err != nil {
		return err
	}

	log.Printf("(%s) :: Setting up long-running message handler", logPrefix)
	asyncMessageHandler := &messageHandler{botID, startupHandler.userIDs, initialMessages, false}
	server.handlers[botID] = asyncMessageHandler
	con.AddHandler(asyncMessageHandler)

	return nil
}

func getConnection(botID string) (*wa.Conn, error) {
	connection, ok := server.connections[botID]
	if !ok {
		return nil, fmt.Errorf("connection not found for botID %s", botID)
	}

	return connection, nil
}

func createConnection() (*wa.Conn, error) {
	con, err := wa.NewConn(30 * time.Second)
	if err != nil {
		return con, err
	}

	con.SetClientName("QuePasa for Link", "QuePasa", "0.4")
	con.SetClientVersion(0, 4, 2088)

	return con, err
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

func SignIn(botID string, out chan<- []byte) error {
	con, err := createConnection()
	if err != nil {
		return err
	}

	qr := make(chan string)
	go func() {
		var png []byte
		png, err := qrcode.Encode(<-qr, qrcode.Medium, 256)
		if err != nil {
			log.Printf("SUFF ERROR C :: %#v\n", err.Error())
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

func SendMessage(botID string, recipient string, message string) (string, error) {
	var messageID string
	con, err := getConnection(botID)
	if err != nil {
		return messageID, err
	}

	//formattedRecipient, _ := CleanPhoneNumber(recipient)
	textMessage := wa.TextMessage{
		Info: wa.MessageInfo{
			RemoteJid: recipient, //formattedRecipient + "@s.whatsapp.net",
		},
		Text: message,
	}

	messageID, err = con.Send(textMessage)
	if err != nil {
		return messageID, err
	}

	//go func() {
	//	RestartServer()
	//}()

	return messageID, nil
}

//
// ReceiveMessages
//

func ReceiveMessages(botID string, timestamp string) ([]QPMessage, error) {
	var messages []QPMessage
	searchTimestamp, err := strconv.ParseUint(timestamp, 10, 64)
	if err != nil {
		searchTimestamp = 1000000
	}

	handler, ok := server.handlers[botID]
	if !ok {
		return messages, nil
	}

	for _, msg := range handler.messages {
		if msg.Timestamp >= searchTimestamp {
			messages = append(messages, msg)
		}
	}

	sort.Sort(ByTimestamp(messages))

	return messages, nil
}

func loadMessages(con *wa.Conn, botID string, userID string, count int) (map[string]QPMessage, error) {
	if con != nil {
		return nil, fmt.Errorf("SUFF ERROR I :: connection not found for botID %s", botID)
	}

	userIDs := make(map[string]bool)
	messages := make(map[string]QPMessage)
	handler := &messageHandler{botID, userIDs, messages, true}
	if handler != nil {
		con.LoadFullChatHistory(userID, count, time.Millisecond*300, handler)
	}
	con.RemoveHandlers()
	return messages, nil
}

func fetchMessages(con *wa.Conn, botID string, userIDs map[string]bool) (map[string]QPMessage, error) {
	messages := make(map[string]QPMessage)

	for userID := range userIDs {
		if string(userID[0]) == "+" {
			continue
		}
		userMessages, err := loadMessages(con, botID, userID, 50)
		if err != nil {
			return messages, err
		}

		for messageID, message := range userMessages {
			var mutex = &sync.Mutex{}
			mutex.Lock()

			messages[messageID] = message

			mutex.Unlock()
		}
	}

	return messages, nil
}

// Message handler

func (h *messageHandler) HandleJsonMessage(message string) {
	fmt.Println("JsonMessage: " + message)
}

func (h *messageHandler) HandleImageMessage(msg wa.ImageMessage) {
	con, err := getConnection(h.botID)
	if err != nil {
		return
	}

	currentUserID, _ := CleanPhoneNumber(con.Info.Wid)
	currentUserID = currentUserID + "@s.whatsapp.net"
	message := QPMessage{}
	message.ID = msg.Info.Id
	message.Timestamp = msg.Info.Timestamp
	message.Body = "Imagem recebida: " + msg.Type
	contact, ok := con.Store.Contacts[msg.Info.RemoteJid]
	if ok {
		message.Name = contact.Name
	}
	if msg.Info.FromMe {
		message.Source = currentUserID
		message.Recipient = msg.Info.RemoteJid
	} else {
		message.Source = msg.Info.RemoteJid
		message.Recipient = currentUserID
	}

	AppenMsgToCache(h, message, msg.Info.RemoteJid)
}

func (h *messageHandler) HandleLocationMessage(msg wa.LocationMessage) {
	con, err := getConnection(h.botID)
	if err != nil {
		return
	}

	currentUserID, _ := CleanPhoneNumber(con.Info.Wid)
	currentUserID = currentUserID + "@s.whatsapp.net"
	message := QPMessage{}
	message.ID = msg.Info.Id
	message.Timestamp = msg.Info.Timestamp
	message.Body = "Localização recebida ... "
	contact, ok := con.Store.Contacts[msg.Info.RemoteJid]
	if ok {
		message.Name = contact.Name
	}
	if msg.Info.FromMe {
		message.Source = currentUserID
		message.Recipient = msg.Info.RemoteJid
	} else {
		message.Source = msg.Info.RemoteJid
		message.Recipient = currentUserID
	}

	AppenMsgToCache(h, message, msg.Info.RemoteJid)
}

func (h *messageHandler) HandleLiveLocationMessage(msg wa.LiveLocationMessage) {
	con, err := getConnection(h.botID)
	if err != nil {
		return
	}

	currentUserID, _ := CleanPhoneNumber(con.Info.Wid)
	currentUserID = currentUserID + "@s.whatsapp.net"

	message := QPMessage{}
	message.ID = msg.Info.Id
	message.Timestamp = msg.Info.Timestamp
	message.Body = "Localização em tempo real recebida ... "
	contact, ok := con.Store.Contacts[msg.Info.RemoteJid]
	if ok {
		message.Name = contact.Name
	}
	if msg.Info.FromMe {
		message.Source = currentUserID
		message.Recipient = msg.Info.RemoteJid
	} else {
		message.Source = msg.Info.RemoteJid
		message.Recipient = currentUserID
	}

	AppenMsgToCache(h, message, msg.Info.RemoteJid)
}

func (h *messageHandler) HandleInfoMessage(msg wa.MessageInfo) {
	b, err := json.Marshal(msg)
	if err != nil {
		fmt.Println(err)
		return
	}

	log.Printf("INFO :: %#v\n", string(b))
}

func (h *messageHandler) HandleDocumentMessage(msg wa.DocumentMessage) {
	con, err := getConnection(h.botID)
	if err != nil {
		return
	}

	currentUserID, _ := CleanPhoneNumber(con.Info.Wid)
	currentUserID = currentUserID + "@s.whatsapp.net"

	message := QPMessage{}
	message.ID = msg.Info.Id
	message.Timestamp = msg.Info.Timestamp
	message.Body = "Documento recebido: " + msg.Type + " :: " + msg.FileName
	contact, ok := con.Store.Contacts[msg.Info.RemoteJid]
	if ok {
		message.Name = contact.Name
	}
	if msg.Info.FromMe {
		message.Source = currentUserID
		message.Recipient = msg.Info.RemoteJid
	} else {
		message.Source = msg.Info.RemoteJid
		message.Recipient = currentUserID
	}

	AppenMsgToCache(h, message, msg.Info.RemoteJid)
}

func (h *messageHandler) HandleContactMessage(msg wa.ContactMessage) {
	con, err := getConnection(h.botID)
	if err != nil {
		return
	}

	currentUserID, _ := CleanPhoneNumber(con.Info.Wid)
	currentUserID = currentUserID + "@s.whatsapp.net"

	message := QPMessage{}
	message.ID = msg.Info.Id
	message.Timestamp = msg.Info.Timestamp
	message.Body = "Contato VCARD recebido ... "
	contact, ok := con.Store.Contacts[msg.Info.RemoteJid]
	if ok {
		message.Name = contact.Name
	}
	if msg.Info.FromMe {
		message.Source = currentUserID
		message.Recipient = msg.Info.RemoteJid
	} else {
		message.Source = msg.Info.RemoteJid
		message.Recipient = currentUserID
	}

	AppenMsgToCache(h, message, msg.Info.RemoteJid)
}

func (h *messageHandler) HandleAudioMessage(msg wa.AudioMessage) {
	con, err := getConnection(h.botID)
	if err != nil {
		return
	}

	currentUserID, _ := CleanPhoneNumber(con.Info.Wid)
	currentUserID = currentUserID + "@s.whatsapp.net"

	message := QPMessage{}
	message.ID = msg.Info.Id
	message.Timestamp = msg.Info.Timestamp
	message.Body = "Audio recebido: " + msg.Type
	contact, ok := con.Store.Contacts[msg.Info.RemoteJid]
	if ok {
		message.Name = contact.Name
	}
	if msg.Info.FromMe {
		message.Source = currentUserID
		message.Recipient = msg.Info.RemoteJid
	} else {
		message.Source = msg.Info.RemoteJid
		message.Recipient = currentUserID
	}

	AppenMsgToCache(h, message, msg.Info.RemoteJid)
}

func (h *messageHandler) HandleTextMessage(msg wa.TextMessage) {
	con, err := ReceiveMessagePreProcessing(h, msg.Info)
	if err != nil {
		log.Printf("SUFF ERROR G :: error on pre processing received message: %v", err)
		return
	}

	message := CreateQPMessage(msg.Info)
	message.FillHeader(msg.Info, con)
	message.Body = msg.Text

	AppenMsgToCache(h, message, msg.Info.RemoteJid)
}

func AppenMsgToCache(h *messageHandler, msg QPMessage, RemoteJid string) error {
	var mutex = &sync.Mutex{}
	mutex.Lock()

	if h != nil {
		h.userIDs[RemoteJid] = true
		h.messages[msg.ID] = msg
	}

	mutex.Unlock()
	return nil
}

func (h *messageHandler) HandleError(publicError error) {
	if e, ok := publicError.(*wa.ErrConnectionFailed); ok {
		log.Printf("SUFF ERROR B :: Connection failed, underlying error: %v", e.Err)
		<-time.After(10 * time.Second)
		RestartServer()
	} else if strings.Contains(publicError.Error(), "tag 174") {
		log.Printf("SUFF ERROR D :: Binary decode error, underlying error: %v", publicError)
		<-time.After(10 * time.Second)
		//RestartServer()
	} else if strings.Contains(publicError.Error(), "code: 1000") {
		log.Printf("SUFF ERROR H :: %v\n", publicError)
		<-time.After(10 * time.Second)
		RestartServer()
	} else {
		log.Printf("SUFF ERROR E :: Message handler error: %v\n", publicError)
	}
}

func (h *messageHandler) ShouldCallSynchronously() bool {
	return h.synchronous
}
