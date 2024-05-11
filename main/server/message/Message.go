package commonMsg

import (
	"main/common"
)

type Message interface {
	GetIdMessage() string
	GetSenderID() int

	GetTypeOfMessage() string

	GetKey() string
	GetValue() string

	GetSendingFIFO() int
}

// MessageCommon rappresenta la struttura dati comune a tutti i messaggi
type MessageCommon struct {
	Id       string // Id del messaggio stesso
	SenderID int    // SenderID rappresenta l'indice del server che invia il messaggio

	TypeOfMessage string
	Args          common.Args
}
