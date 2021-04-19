package models

import wa "github.com/Rhymen/go-whatsapp"

// Cria uma mensagem no formato do QuePasa apartir de uma msg do WhatsApp
// Preenche somente as propriedades padrões do cabecalho
func CreateQPMessage(msg wa.TextMessage) QPMessage {
	message := QPMessage{}
	message.ID = msg.Info.Id
	message.Timestamp = msg.Info.Timestamp
	return message
}
