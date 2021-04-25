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
var mutex = &sync.Mutex{}

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
		log.Printf("(%s) :: Adding message handlers for %s with token: %s", bot.ID, bot.Number, bot.Token)

		err = startHandler(bot.ID)
		if err != nil {
			return err
		}
	}

	return nil
}

func startHandler(botID string) error {
	con, err := createConnection()
	if err != nil {
		return err
	}

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

	log.Printf("(%s) :: Fetching initial messages", botID)
	initialMessages, err := fetchMessages(con, botID, startupHandler.userIDs)
	if err != nil {
		return err
	}

	log.Printf("(%s) :: Setting up long-running message handler", botID)
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

func SendMessage(botID string, recipient string, message string) (messageID string, err error) {
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

	con, err := getConnection(botID)
	if err != nil {
		return
	}

	log.Printf("sending message from bot: %s :: to recipient: %s", botID, recipient)
	//formattedRecipient, _ := CleanPhoneNumber(recipient)
	textMessage := wa.TextMessage{
		Info: wa.MessageInfo{
			RemoteJid: recipient, //formattedRecipient + "@s.whatsapp.net",
		},
		Text: message,
	}

	messageID, err = con.Send(textMessage)
	return
}

// Receive messages from the controller, external
func ReceiveMessages(botID string, timestamp string) (messages []QPMessage, err error) {
	searchTimestamp, _ := strconv.ParseUint(timestamp, 10, 64)
	if searchTimestamp == 0 {
		searchTimestamp = 1000000
	}

	handler, ok := server.handlers[botID]
	if !ok {
		err = fmt.Errorf("handlers not read yet, please wait")
		return
	}

	for _, msg := range handler.messages {
		if msg.Timestamp >= searchTimestamp {
			mutex.Lock() // travando multi threading

			// Incluindo mensagem na lista de retorno
			messages = append(messages, msg)

			mutex.Unlock() // destravando multi threading
		}
	}

	mutex.Lock() // travando multi threading
	sort.Sort(ByTimestamp(messages))
	mutex.Unlock() // destravando multi threading

	return
}

func loadMessages(con *wa.Conn, botID string, userID string, count int) (map[string]QPMessage, error) {

	userIDs := make(map[string]bool)
	messages := make(map[string]QPMessage)
	handler := &messageHandler{botID, userIDs, messages, true}

	if con != nil {
		con.LoadFullChatHistory(userID, count, time.Millisecond*300, handler)
		con.RemoveHandlers()
	}

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
			mutex.Lock()

			messages[messageID] = message

			mutex.Unlock()
		}
	}

	return messages, nil
}

// Message handler

func (h *messageHandler) HandleJsonMessage(message string) {
	if isDevelopment() {
		fmt.Println("JsonMessage: " + message)
	}
}

func (h *messageHandler) HandleInfoMessage(msg wa.MessageInfo) {
	b, err := json.Marshal(msg)
	if err != nil {
		fmt.Println(err)
		return
	}

	log.Printf("INFO :: %#v\n", string(b))
}

func (h *messageHandler) HandleImageMessage(msg wa.ImageMessage) {
	con, err := ReceiveMessagePreProcessing(h, msg.Info)
	if err != nil {
		log.Printf("SUFF ERROR G :: ImageMessage error on pre processing received message: %v", err)
		return
	}

	message := CreateQPMessage(msg.Info)
	message.FillHeader(msg.Info, con)

	//  --> Personalizado para esta seção
	message.Text = "Imagem recebida: " + msg.Type
	message.FillImageAttachment(msg, con)
	//  <--

	AppenMsgToCache(h, message, msg.Info.RemoteJid)
}

func (h *messageHandler) HandleLocationMessage(msg wa.LocationMessage) {
	con, err := ReceiveMessagePreProcessing(h, msg.Info)
	if err != nil {
		log.Printf("SUFF ERROR G :: LocationMessage error on pre processing received message: %v", err)
		return
	}

	message := CreateQPMessage(msg.Info)
	message.FillHeader(msg.Info, con)

	//  --> Personalizado para esta seção
	message.Text = "Localização recebida ... "
	//  <--

	AppenMsgToCache(h, message, msg.Info.RemoteJid)
}

func (h *messageHandler) HandleLiveLocationMessage(msg wa.LiveLocationMessage) {
	con, err := ReceiveMessagePreProcessing(h, msg.Info)
	if err != nil {
		log.Printf("SUFF ERROR G :: LiveLocationMessage error on pre processing received message: %v", err)
		return
	}

	message := CreateQPMessage(msg.Info)
	message.FillHeader(msg.Info, con)

	//  --> Personalizado para esta seção
	message.Text = "Localização em tempo real recebida ... "
	//  <--

	AppenMsgToCache(h, message, msg.Info.RemoteJid)
}

func (h *messageHandler) HandleDocumentMessage(msg wa.DocumentMessage) {
	con, err := ReceiveMessagePreProcessing(h, msg.Info)
	if err != nil {
		log.Printf("SUFF ERROR G :: DocumentMessage error on pre processing received message: %v", err)
		return
	}

	message := CreateQPMessage(msg.Info)
	message.FillHeader(msg.Info, con)
	//message.FillDocumentAttachment(msg, con)

	//  --> Personalizado para esta seção
	message.Text = "Documento recebido: " + msg.Type + " :: " + msg.FileName
	//  <--

	AppenMsgToCache(h, message, msg.Info.RemoteJid)
}

func (h *messageHandler) HandleContactMessage(msg wa.ContactMessage) {
	con, err := ReceiveMessagePreProcessing(h, msg.Info)
	if err != nil {
		log.Printf("SUFF ERROR G :: ContactMessage error on pre processing received message: %v", err)
		return
	}

	message := CreateQPMessage(msg.Info)
	message.FillHeader(msg.Info, con)

	//  --> Personalizado para esta seção
	message.Text = "Contato VCARD recebido ... "
	//  <--

	AppenMsgToCache(h, message, msg.Info.RemoteJid)
}

func (h *messageHandler) HandleAudioMessage(msg wa.AudioMessage) {
	con, err := ReceiveMessagePreProcessing(h, msg.Info)
	if err != nil {
		log.Printf("SUFF ERROR G :: AudioMessage error on pre processing received message: %v", err)
		return
	}

	message := CreateQPMessage(msg.Info)
	message.FillHeader(msg.Info, con)

	//  --> Personalizado para esta seção
	message.Text = "Audio recebido: " + msg.Type
	//  <--

	AppenMsgToCache(h, message, msg.Info.RemoteJid)
}

func (h *messageHandler) HandleTextMessage(msg wa.TextMessage) {
	con, err := ReceiveMessagePreProcessing(h, msg.Info)
	if err != nil {
		log.Printf("SUFF ERROR G :: TextMessage error on pre processing received message: %v", err)
		return
	}

	message := CreateQPMessage(msg.Info)
	message.FillHeader(msg.Info, con)

	//  --> Personalizado para esta seção
	message.Text = msg.Text
	//  <--

	AppenMsgToCache(h, message, msg.Info.RemoteJid)
}

func AppenMsgToCache(h *messageHandler, msg QPMessage, RemoteJid string) error {
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
	} else if strings.Contains(publicError.Error(), "received invalid data") {
		return // ignorando erro conhecido
	} else if strings.Contains(publicError.Error(), "tag 174") {
		log.Printf("SUFF ERROR D :: Binary decode error, underlying error: %v", publicError)
		<-time.After(10 * time.Second)
		//RestartServer()
	} else if strings.Contains(publicError.Error(), "code: 1000") {
		log.Printf("SUFF ERROR H :: %v\n", publicError)
		<-time.After(10 * time.Second)
		RestartServer()
	} else {
		log.Printf("SUFF ERROR E :: Message handler error: %v\n", publicError.Error())
	}
}

func (h *messageHandler) ShouldCallSynchronously() bool {
	return h.synchronous
}
