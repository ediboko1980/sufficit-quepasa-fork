package controllers

import "github.com/sufficit/sufficit-quepasa-fork/models"

type QPSendRequest struct {
	Recipient  string              `json:"recipient,omitempty"`
	Message    string              `json:"message,omitempty"`
	Attachment models.QPAttachment `json:"attachment,omitempty"`
}
