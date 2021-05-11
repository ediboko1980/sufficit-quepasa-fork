package models

import (
	"encoding/base64"
	"strings"

	wa "github.com/Rhymen/go-whatsapp"
)

// Mensagem no formato QuePasa
// Utilizada na API do QuePasa para troca com outros sistemas
type QPAttachment struct {
	Url         string `json:"url,omitempty"`
	B64MediaKey string `json:"b64mediakey,omitempty"`
	Length      int    `json:"length,omitempty"`
	MIME        string `json:"mime,omitempty"`
	Base64      string `json:"base64,omitempty"`
	FileName    string `json:"filename,omitempty"`
}

// Traz o MediaType para download do whatsapp
func (m QPAttachment) WAMediaType() wa.MediaType {

	if strings.Contains(m.MIME, "document") {
		return wa.MediaDocument
	}

	// apaga informações após o ;
	// fica somente o mime mesmo
	mimeOnly := strings.Split(m.MIME, ";")
	switch mimeOnly[0] {
	case "image/jpeg":
		return wa.MediaImage
	case "audio/ogg", "audio/mpeg", "audio/mp4":
		return wa.MediaAudio
	default:
		return wa.MediaDocument
	}
}

// Traz o MediaKey em []byte apatir de base64
func (m QPAttachment) MediaKey() ([]byte, error) {
	return base64.StdEncoding.DecodeString(m.B64MediaKey)
}
