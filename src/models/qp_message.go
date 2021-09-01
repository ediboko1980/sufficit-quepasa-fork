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
	Participant QPEndPoint `json:"participant,omitempty"`

	// Fui eu quem enviou a msg ?
	FromMe bool `json:"fromme"`

	// Texto da msg
	Text string `json:"text"`

	Attachment QPAttachment `json:"attachment,omitempty"`
}

type ByTimestamp []QPMessage

func (m ByTimestamp) Len() int           { return len(m) }
func (m ByTimestamp) Less(i, j int) bool { return m[i].Timestamp > m[j].Timestamp }
func (m ByTimestamp) Swap(i, j int)      { m[i], m[j] = m[j], m[i] }

func (source QPMessage) ToV2() QPMessageV2 {
	message := QPMessageV2{
		ID:          source.ID,
		Timestamp:   int(source.Timestamp),
		Controller:  source.Controller,
		ReplyTo:     source.ReplyTo,
		Participant: source.Participant,
		FromMe:      source.FromMe,
		Text:        source.Text,
		Attachment:  source.Attachment,
		Chat:        source.ReplyTo.ToQPChatV2(),
	}
	return message
}
