package models

import "github.com/sufficit/sufficit-quepasa-fork/models_v2"

// Destino de msg whatsapp
type QPEndPoint struct {
	ID     string `json:"id"`
	Phone  string `json:"phone"`
	Title  string `json:"title,omitempty"`
	Status string `json:"status,omitempty"`
}

type QPEndPointV2 struct {
	ID        string `json:"id"`
	UserName  string `json:"username,omitempty"`
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
}

func (source QPEndPoint) GetQPEndPointV2() QPEndPointV2 {
	ob2 := QPEndPointV2{ID: source.ID, UserName: source.Phone, FirstName: source.Title, LastName: source.Status}
	return ob2
}

func (source QPEndPoint) ToQPUserV2() models_v2.QPUserV2 {
	result := models_v2.QPUserV2{
		ID: source.ID,
	}
	return result
}

func (source QPEndPoint) ToQPChatV2() models_v2.QPChatV2 {
	result := models_v2.QPChatV2{
		ID: source.ID,
	}
	return result
}
