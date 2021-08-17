package models_v2

// Mensagem no formato QuePasa
// Utilizada na API do QuePasa para troca com outros sistemas
type QPMessageV2 struct {
	MessageID    string   `json:"message_id"`
	From         QPUserV2 `json:"from"`
	Date         int      `json:"date"`
	EditDate     int      `json:"edit_date,omitempty"`
	Chat         QPChatV2 `json:"chat"`
	NewChatTitle string   `json:"new_chat_title,omitempty"`
	Caption      string   `json:"caption,omitempty"`
	Text         string   `json:"text,omitempty"`

	//Photo      Photo   `json:"photo,omitempty"`
	//sticker
	//voice
	//video
	//document
}
