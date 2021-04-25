package models

import (
	"encoding/base64"
	"strings"

	wa "github.com/Rhymen/go-whatsapp"
)

// Mensagem no formato QuePasa
// Utilizada na API do QuePasa para troca com outros sistemas
type QPAttachment struct {
	Url         string
	B64MediaKey string
	Length      int
	MIME        string
}

// Traz o MediaType para download do whatsapp
func (m QPAttachment) WAMediaType() wa.MediaType {

	// apaga informações após o ;
	// fica somente o mime mesmo
	mimeOnly := strings.Split(m.MIME, ";")
	switch mimeOnly[0] {
	case "image/jpeg":
		return wa.MediaImage
	case "audio/ogg":
		return wa.MediaAudio
	default:
		return wa.MediaDocument
	}
}

// Traz o MediaKey em []byte apatir de base64
func (m QPAttachment) MediaKey() ([]byte, error) {
	return base64.StdEncoding.DecodeString(m.B64MediaKey)
}
