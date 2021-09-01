package models

type QPSendDocumentRequestV2 struct {
	Recipient  string       `json:"recipient,omitempty"`
	Message    string       `json:"message,omitempty"`
	Attachment QPAttachment `json:"attachment,omitempty"`
}
