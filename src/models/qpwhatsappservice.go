package models

import (
	"log"
	"sync"
)

// Serviço que controla os servidores / bots individuais do whatsapp
type QPWhatsAppService struct {
	Servers map[string]*QPWhatsAppServer
	Sync    *sync.Mutex // Objeto de sinaleiro para evitar chamadas simultâneas a este objeto
}

var WhatsAppService *QPWhatsAppService

func QPWhatsAppStart() {
	log.Println("Starting WhatsApp Service ....")

	servers := make(map[string]*QPWhatsAppServer)
	sync := &sync.Mutex{}
	WhatsAppService = &QPWhatsAppService{servers, sync}

	// iniciando servidores e cada bot individualmente
	err := WhatsAppService.initService()
	if err != nil {
		log.Printf("Problema ao instanciar bots .... %s", err)
	}
}

// Inclui um novo servidor em um serviço já em andamento
// *Usado quando se passa pela verificação do QRCode
// *Usado quando se inicializa o sistema
func (service *QPWhatsAppService) AppendNewServer(bot QPBot) {
	// Trava simultaneos
	service.Sync.Lock()

	// Cria um novo servidor
	server := CreateWhatsAppServer(bot)

	// Adiciona na lista de servidores
	service.Servers[bot.ID] = &server

	// Destrava simultaneos
	service.Sync.Unlock()

	// Inicializa o servidor
	go server.Initialize()
}

// Função privada que irá iniciar todos os servidores apartir do banco de dados
func (service *QPWhatsAppService) initService() error {
	bots, err := FindAllBots(GetDB())
	if err != nil {
		return err
	}

	for _, bot := range bots {
		// Somente será iniciado o bot/servidor que estiver verificado
		if bot.Verified {
			service.AppendNewServer(bot)
		}
	}

	return nil
}

func GetServer(botID string) (server *QPWhatsAppServer, ok bool) {
	server, ok = WhatsAppService.Servers[botID]
	return
}
