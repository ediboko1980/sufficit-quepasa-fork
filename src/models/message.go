package models

type Message struct {
	ID        string `json:"id"`
	Source    string `json:"source"`
	Timestamp string `json:"timestamp"`
	Body      string `json:"body"`
}
