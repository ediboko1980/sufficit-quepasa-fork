package models

type WhatsAppJsonMessage struct {
	Cmd WhatsAppCmdMessage `json:"cmd"`
}

type WhatsAppCmdMessage struct {
	Type string `json:"type"`
	Kind string `json:"kind"`
}
