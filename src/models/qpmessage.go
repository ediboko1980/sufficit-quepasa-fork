package models

// Mensagem no formato QuePasa
// Utilizada na API do QuePasa para troca com outros sistemas
type QPMessage struct {
	ID        string `json:"id"`
	Timestamp uint64 `json:"timestamp"`

	// Whatsapp que gerencia a bagaça toda
	Controller QPEndPoint `json:"controller"`

	// Endereço garantido que deve receber uma resposta
	ReplyTo QPEndPoint `json:"replyto"`

	// Se a msg foi postado em algum grupo ? quem postou !
	Participant QPEndPoint `json:"participant"`

	// Fui eu quem enviou a msg ?
	FromMe bool `json:"fromme"`

	// Forma (Content Type ou MIME) que se apresenta o corpo da msg
	Type string `json:"type"`

	// Conteudo da msg
	Body string `json:"body"`
}

type ByTimestamp []QPMessage

func (m ByTimestamp) Len() int           { return len(m) }
func (m ByTimestamp) Less(i, j int) bool { return m[i].Timestamp > m[j].Timestamp }
func (m ByTimestamp) Swap(i, j int)      { m[i], m[j] = m[j], m[i] }
