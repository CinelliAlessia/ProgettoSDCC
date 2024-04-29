package msg

import (
	"main/common"
)

type Message interface {
	GetIdMessage() string
	GetIdSender() int

	GetTypeOfMessage() string

	GetKey() string
	GetValue() string

	GetOrderClient() int
}

type MessageCommon struct {
	Id       string // Id del messaggio stesso
	IdSender int    // IdSender rappresenta l'indice del server che invia il messaggio

	TypeOfMessage string
	Args          common.Args
}
