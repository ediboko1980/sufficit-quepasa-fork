package models

// Mensagem no formato QuePasa
// Utilizada na API do QuePasa para troca com outros sistemas
type QPMessageV2 struct {
	ID        string `json:"message_id"`
	Timestamp int    `json:"timestamp"`

	// Whatsapp que gerencia a bagaça toda
	Controller QPEndPoint `json:"controller"`

	// Endereço garantido que deve receber uma resposta
	ReplyTo QPEndPoint `json:"replyto"`

	// Se a msg foi postado em algum grupo ? quem postou !
	Participant QPEndPoint `json:"participant,omitempty"`

	// Fui eu quem enviou a msg ?
	FromMe bool `json:"fromme"`

	// Texto da msg
	Text string `json:"text"`

	Attachment QPAttachment `json:"attachment,omitempty"`

	Chat QPChatV2 `json:"chat"`
}
