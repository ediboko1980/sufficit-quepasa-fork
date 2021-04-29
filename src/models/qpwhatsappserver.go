package models

import (
	"log"
	"sync"
	"time"

	wa "github.com/Rhymen/go-whatsapp"
)

type QPWhatsAppServer struct {
	Bot            QPBot
	Connection     *wa.Conn
	Handlers       *QPMessageHandler
	Recipients     map[string]bool
	Messages       map[string]QPMessage
	SyncConnection *sync.Mutex // Objeto de sinaleiro para evitar chamadas simultâneas a este objeto
	SyncMessages   *sync.Mutex // Objeto de sinaleiro para evitar chamadas simultâneas a este objeto
}

// Inicializa um repetidor eterno que confere o estado da conexão e tenta novamente a cada 10 segundos
func (server *QPWhatsAppServer) Initialize() (err error) {
	log.Printf("(%s) Initializing WhatsApp Server ...", server.Bot.Number)
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

func (server *QPWhatsAppServer) Start() (err error) {
	log.Printf("(%s) Starting WhatsApp Server ...", server.Bot.Number)

	server.SyncConnection.Lock() // Travando
	// ------

	// Inicializando conexões e handlers
	err = server.startHandlers()
	if err != nil {
		log.Printf("(%s) SUFF ERROR F :: Starting Handlers error ... %s :", server.Bot.Number, err)
	}

	// ------
	server.SyncConnection.Unlock() // Destravando

	return
}

func (server *QPWhatsAppServer) Restart() {
	log.Printf("(%s) Restarting WhatsApp Server ...", server.Bot.Number)

	server.SyncConnection.Lock() // Travando
	// ------

	server.Connection.RemoveHandlers()
	server.Connection.Disconnect()

	// ------
	server.SyncConnection.Unlock() // Destravando

	// Inicia novamente o servidor e os Handlers(alças)
	server.Start()
}

// Salva em cache e inicia gatilhos assíncronos
func (server *QPWhatsAppServer) AppenMsgToCache(msg QPMessage) error {

	server.SyncMessages.Lock() // Sinal vermelho para atividades simultâneas
	// Apartir deste ponto só se executa um por vez

	//server.Recipients[msg.ReplyTo.ID] = true
	server.Messages[msg.ID] = msg

	server.SyncMessages.Unlock() // Sinal verde !

	// Executando WebHook de forma assincrona
	go server.Bot.PostToWebHook(msg)

	return nil
}

func (server *QPWhatsAppServer) GetMessages(timestamp uint64) (messages []QPMessage, err error) {
	server.SyncMessages.Lock() // Sinal vermelho para atividades simultâneas
	for _, item := range server.Messages {
		if item.Timestamp >= timestamp {
			messages = append(messages, item)
		}
	}
	server.SyncMessages.Unlock() // Sinal verde !
	return
}

func (server *QPWhatsAppServer) startHandlers() (err error) {
	con, err := CreateConnection()
	if err != nil {
		return err
	}

	// Definindo conexão
	server.Connection = con

	// Definindo handlers para mensagens assincronas
	startupHandler := &QPMessageHandler{&server.Bot, true, server}
	con.AddHandler(startupHandler)

	// Consultando banco de dados e buscando dados de alguma seção salva
	session, err := ReadSession(server.Bot.ID)
	if err != nil {
		return
	}

	// Agora sim, restaura a conexão com o whatsapp apartir de uma seção salva
	session, err = con.RestoreWithSession(session)
	if err != nil {
		return
	}

	// Aguarda 3 segundos
	<-time.After(3 * time.Second)

	// Atualiza o banco de dados com os novos dados
	if err = writeSession(server.Bot.ID, session); err != nil {
		return
	}

	con.RemoveHandlers()

	log.Printf("(%s) Fetching initial messages", server.Bot.Number)
	err = server.fetchMessages(con, server.Bot, server.Recipients)
	if err != nil {
		return err
	}

	log.Printf("(%s) Setting up long-running message handler", server.Bot.Number)
	asyncMessageHandler := &QPMessageHandler{&server.Bot, true, server}
	server.Handlers = asyncMessageHandler
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

// importante para não derrubar as conexões
func (server *QPWhatsAppServer) SendMessage(msg interface{}) (string, error) {
	server.SyncMessages.Lock()
	messageID, err := server.Connection.Send(msg)
	server.SyncMessages.Unlock()
	return messageID, err
}
