package models

import (
	"encoding/base64"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	wa "github.com/Rhymen/go-whatsapp"
	"github.com/skip2/go-qrcode"
)

type QPWhatsAppServer struct {
	Bot            QPBot
	Connection     *wa.Conn
	Handlers       QPMessageHandler
	Recipients     map[string]bool
	Messages       map[string]QPMessage
	syncConnection *sync.Mutex // Objeto de sinaleiro para evitar chamadas simultâneas a este objeto
	syncMessages   *sync.Mutex // Objeto de sinaleiro para evitar chamadas simultâneas a este objeto
	Status         *string
	Battery        *WhatsAppBateryStatus
}

// Envia o QRCode para o usuário e aguarda pela resposta
func SignInWithQRCode(user QPUser, out chan<- []byte) (bot QPBot, err error) {
	con, err := CreateConnection()
	if err != nil {
		return
	}

	qr := make(chan string)
	go func() {
		var png []byte
		png, err := qrcode.Encode(<-qr, qrcode.Medium, 256)
		if err != nil {
			log.Printf("(%s)(ERR) SUFF ERROR C :: %#v\n", bot.GetNumber(), err.Error())
		}
		encodedPNG := base64.StdEncoding.EncodeToString(png)
		out <- []byte(encodedPNG)
	}()

	session, err := con.Login(qr)
	if err != nil {
		return
	}

	// Se chegou até aqui é pq o QRCode foi validado e sincronizado
	bot, err = WhatsAppService.DB.Bot.GetOrCreate(con.Info.Wid, user.ID)
	if err != nil {
		return
	}

	err = WriteSession(con.Info.Wid, session)
	return
}

// Instanciando um novo servidor para controle de whatsapp
func CreateWhatsAppServer(bot QPBot) QPWhatsAppServer {

	// Definindo conexão com whatsapp
	connection, _ := CreateConnection()

	handlers := &QPMessageHandler{}
	syncConnetion := &sync.Mutex{}
	syncMessages := &sync.Mutex{}
	recipients := make(map[string]bool)
	messages := make(map[string]QPMessage)
	status := "created"
	batery := WhatsAppBateryStatus{}
	return QPWhatsAppServer{bot, connection, *handlers, recipients, messages, syncConnetion, syncMessages, &status, &batery}
}

// Inicializa um repetidor eterno que confere o estado da conexão e tenta novamente a cada 10 segundos
func (server *QPWhatsAppServer) Initialize() (err error) {
	log.Printf("(%s) Initializing WhatsApp Server ...", server.Bot.GetNumber())
	for {
		err = server.Start()
		if err == nil {
			break
		}

		// Aguardaremos 10 segundos e vamos tentar novamente
		time.Sleep(10 * time.Second)
	}
	return nil
}

// Inicializa um repetidor eterno que confere o estado da conexão e tenta novamente a cada 10 segundos
func (server *QPWhatsAppServer) Shutdown() (err error) {
	//server.syncConnection.Lock() // Travando

	*server.Status = "halting"
	log.Printf("(%s) Shutting Down WhatsApp Server ...", server.Bot.GetNumber())

	server.Connection.RemoveHandlers()

	_, err = server.Connection.Disconnect()

	// caso erro diferente de nulo e não seja pq já esta desconectado
	if err != nil && !strings.Contains(err.Error(), "not connected") {
		log.Printf("(%s)(ERR) Shutting WhatsApp Server : %s", server.Bot.GetNumber(), err.Error())
	} else {
		*server.Status = "stopped"
	}

	//server.syncConnection.Unlock() // Destravando
	return
}

func (server *QPWhatsAppServer) Start() (err error) {
	server.syncConnection.Lock() // Travando

	*server.Status = "starting"
	log.Printf("(%s) Starting WhatsApp Server ...", server.Bot.GetNumber())

	// Inicializando conexões e handlers
	err = server.startHandlers()
	if err != nil {
		*server.Status = "fail"
		switch err.(type) {
		default:
			if strings.Contains(err.Error(), "401") {
				log.Printf("(%s) WhatsApp return a unauthorized state, please verify again", server.Bot.GetNumber())
				err = server.Bot.MarkVerified(false)
			} else if strings.Contains(err.Error(), "restore session connection timed out") {
				log.Printf("(%s) WhatsApp returns after a timeout, trying again in 10 seconds, please wait ...", server.Bot.GetNumber())
			} else {
				log.Printf("(%s)(ERR) SUFF ERROR F :: Starting Handlers error ... %s :", server.Bot.GetNumber(), err)
			}
		case *ServiceUnreachableError:
			log.Println(err)
	   }		

		// Importante para evitar que a conexão em falha continue aberta
		server.Connection.RemoveHandlers()
		server.Connection.Disconnect()	

	} else {
		*server.Status = "ready"
	}

	server.syncConnection.Unlock() // Destravando
	return
}

func (server *QPWhatsAppServer) Restart() {
	// Somente executa caso não esteja em estado de processo de conexão
	// Evita chamadas simultâneas desnecessárias
	if !strings.Contains(*server.Status, "starting") {
		*server.Status = "restarting"
		log.Printf("(%s) Restarting WhatsApp Server ...", server.Bot.GetNumber())

		server.Connection.RemoveHandlers()
		server.Connection.Disconnect()
		*server.Status = "disconnected"

		// Inicia novamente o servidor e os Handlers(alças)
		err := server.Initialize()
		if err != nil {
			*server.Status = "critical"
			log.Printf("(%s)(ERR) Critical error on WhatsApp Server: %s", server.Bot.GetNumber(), err.Error())
		}
	}
}

// Somente usar em caso de não ser permitida a reconxão automática
func (server *QPWhatsAppServer) Disconnect(cause string) {
	log.Printf("(%s) Disconnecting WhatsApp Server: %s", server.Bot.GetNumber(), cause)

	server.syncConnection.Lock() // Travando
	// ------

	server.Connection.RemoveHandlers()
	server.Connection.Disconnect()

	// ------
	server.syncConnection.Unlock() // Destravando
}

// Salva em cache e inicia gatilhos assíncronos
func (server *QPWhatsAppServer) AppenMsgToCache(msg QPMessage) error {

	server.syncConnection.Lock() // Sinal vermelho para atividades simultâneas
	// Apartir deste ponto só se executa um por vez

	//server.Recipients[msg.ReplyTo.ID] = true
	server.Messages[msg.ID] = msg

	server.syncConnection.Unlock() // Sinal verde !

	// Executando WebHook de forma assincrona
	go server.Bot.PostToWebHook(msg)

	return nil
}

func (server *QPWhatsAppServer) GetMessages(timestamp uint64) (messages []QPMessage, err error) {
	server.syncConnection.Lock() // Sinal vermelho para atividades simultâneas
	for _, item := range server.Messages {
		if item.Timestamp >= timestamp {
			messages = append(messages, item)
		}
	}
	server.syncConnection.Unlock() // Sinal verde !
	return
}

func (server *QPWhatsAppServer) startHandlers() (err error) {
	con, err := CreateConnection()
	if err != nil {		
		if strings.Contains(err.Error(), "bad handshake") {
			return &ServiceUnreachableError { 
				Server: server.Bot.GetNumber(),
				Message: "bad handshake",
			}			
		} else {
			log.Printf("(%s)(ERR) SUFF ERROR H :: Starting Handlers error ... %s :", server.Bot.GetNumber(), err)
		}
		return 
	}

	// Definindo conexão
	server.Connection = con

	// Definindo handlers para mensagens assincronas
	startupHandler := &QPMessageHandler{&server.Bot, true, server}
	con.AddHandler(startupHandler)

	// Consultando banco de dados e buscando dados de alguma seção salva
	session, err := ReadSession(server.Bot.ID)
	if err != nil {
		log.Printf("(%s)(ERR) Error on reading session :: %s", server.Bot.GetNumber(), err)
		return
	}

	// Agora sim, restaura a conexão com o whatsapp apartir de uma seção salva
	session, err = con.RestoreWithSession(session)
	if err != nil {
		log.Printf("(%s)(ERR) Error on restore session :: %s", server.Bot.GetNumber(), err)
		return
	}

	// Atualizando informação sobre o estado da conexão e do servidor
	*server.Status = "connected"

	// Aguarda 3 segundos
	<-time.After(3 * time.Second)

	// Atualiza o banco de dados com os novos dados
	if err = WriteSession(server.Bot.ID, session); err != nil {
		log.Printf("(%s)(ERR) Erro on writing session :: %s", server.Bot.GetNumber(), err)
		return
	}

	con.RemoveHandlers()

	*server.Status = "fetching"
	log.Printf("(%s) Fetching initial messages", server.Bot.GetNumber())
	err = server.fetchMessages(con, server.Bot, server.Recipients)
	if err != nil {
		return err
	}

	log.Printf("(%s) Setting up long-running message handler", server.Bot.GetNumber())
	asyncMessageHandler := &QPMessageHandler{&server.Bot, true, server}
	server.Handlers = *asyncMessageHandler
	con.AddHandler(asyncMessageHandler)
	return
}

func (server *QPWhatsAppServer) fetchMessages(con *wa.Conn, bot QPBot, recipients map[string]bool) (err error) {
	for userID := range recipients {
		if string(userID[0]) == "+" {
			continue
		}

		// Busca até 50 msg de cada conversa para colocar no cache
		err = server.loadMessages(con, bot, userID, 50)
		if err != nil {
			return
		}
	}
	return
}

// Carrega as msg do histórico
// Chamado antes de ativar os handlers
// Após carregar, salva no cache automaticamente
func (server *QPWhatsAppServer) loadMessages(con *wa.Conn, bot QPBot, userID string, count int) (err error) {
	handler := &QPMessageHandler{&server.Bot, true, server}
	if con != nil {
		con.LoadFullChatHistory(userID, count, time.Millisecond*300, handler)
		con.RemoveHandlers()
	}
	return
}

func SendWhatsAppMessage(con *wa.Conn, msg interface{}) (msgid string, err error) {
	msgid, err = con.Send(msg)
	return
}

// importante para não derrubar as conexões (ainda não funcionando)
func (server *QPWhatsAppServer) SendMessage(msg interface{}) (string, error) {

	if *server.Status != "ready" {
		return "", fmt.Errorf("server not ready, wait")
	}

	messageID, err := SendWhatsAppMessage(server.Connection, msg)

	return messageID, err
}

func (server *QPWhatsAppServer) IsDevelopment() bool {
	if ENV.IsDevelopment() {
		return server.Bot.Devel
	}
	return false
}

// Retorna o titulo em cache (se houver) do id passado em parametro
func (server *QPWhatsAppServer) GetTitle(Wid string) string {
	return getTitle(server.Connection.Store, Wid)
}
