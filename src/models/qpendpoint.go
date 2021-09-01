package models

// Destino de msg whatsapp
type QPEndPoint struct {
	ID     string `json:"id"`
	Phone  string `json:"phone"`
	Title  string `json:"title,omitempty"`
	Status string `json:"status,omitempty"`
}

func (source QPEndPoint) GetQPEndPointV2() QPEndpointV2 {
	ob2 := QPEndpointV2{ID: source.ID, UserName: source.Phone, FirstName: source.Title, LastName: source.Status}
	return ob2
}

func (source QPEndPoint) ToQPUserV2() QPUserV2 {
	result := QPUserV2{
		ID: source.ID,
	}
	return result
}

func (source QPEndPoint) ToQPChatV2() QPChatV2 {
	result := QPChatV2{
		ID: source.ID,
	}
	return result
}
