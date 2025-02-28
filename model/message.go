package model

import "time"

type MessageType string

const (
	NewClient  MessageType = "new_client"
	Chat       MessageType = "chat"
	EndSession MessageType = "end_session"
	Error      MessageType = "error"
)

func (mt *MessageType) IsValid() bool {
	switch *mt {
	case NewClient:
	case Chat:
	case EndSession:
		return true
	}

	return false
}

type MessageError string

const (
	ErrorUserAlreadyExists MessageError = "User already exists!"
)

type Message struct {
	Sender    string      `json:"sender"`
	Receiver  string      `json:"receiver"`
	Text      string      `json:"text"`
	Type      MessageType `json:"type"`
	Timestamp time.Time   `json:"timestamp"`
}
