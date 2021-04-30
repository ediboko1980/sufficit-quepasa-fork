package models

// Destino de msg whatsapp
type QPEndPoint struct {
	ID     string `json:"id"`
	Phone  string `json:"phone"`
	Title  string `json:"title,omitempty"`
	Status string `json:"status,omitempty"`
}
