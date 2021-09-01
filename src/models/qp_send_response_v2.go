package models

type QPSendResponseV2 struct {
	ID   string       `json:"message_id"`
	Date int          `json:"date,omitempty"`
	From QPEndPointV2 `json:"from,omitempty"`
	Chat QPEndPointV2 `json:"chat,omitempty"`

	// Para compatibilidade apenas
	PreviusV1 QPSendResult `json:"result,omitempty"`
}
