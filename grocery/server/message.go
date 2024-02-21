package server

import (
	"encoding/json"
	"log"
)

type (
	Message struct {
		Code    int         `json:"code,omitempty"`
		Message string      `json:"message,omitempty"`
		Data    interface{} `json:"data,omitempty"`
	}
)

func (msg *Message) Marshal() []byte {
	b, err := json.Marshal(msg)
	if err != nil {
		log.Printf("marshaling message [ERR: %s]", err)
	}

	return b
}
