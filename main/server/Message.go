package main

import (
	"main/common"
)

type Message interface {
	GetIdMessage() string
	GetIdSender() int

	GetTypeOfMessage() string

	GetKey() string
	GetValue() string

	GetClock() interface{}
	GetOrderClient() int
}

type MessageC struct {
	Id       string // Id del messaggio stesso
	IdSender int    // IdSender rappresenta l'indice del server che invia il messaggio

	TypeOfMessage string
	Args          common.Args

	VectorClock [common.Replicas]int // Orologio vettoriale
}

type MessageS struct {
	Id       string
	IdSender int // IdSender rappresenta l'indice del server che invia il messaggio

	TypeOfMessage string
	Args          common.Args

	LogicalClock int
	NumberAck    int
}

// ----- Consistenza Causale ----- //

func (msg MessageC) GetIdSender() int {
	return msg.IdSender
}

func (msg MessageC) GetIdMessage() string {
	return msg.Id
}

func (msg MessageC) GetTypeOfMessage() string {
	return msg.TypeOfMessage
}

func (msg MessageC) GetKey() string {
	return msg.Args.Key
}

func (msg MessageC) GetValue() string {
	return msg.Args.Value
}

func (msg MessageC) GetClock() interface{} {
	return msg.VectorClock
}

func (msg MessageC) GetOrderClient() int {
	return msg.Args.Timestamp
}

// ----- Consistenza Sequenziale ----- //

func (msg MessageS) GetIdSender() int {
	return msg.IdSender
}

func (msg MessageS) GetIdMessage() string {
	return msg.Id
}

func (msg MessageS) GetTypeOfMessage() string {
	return msg.TypeOfMessage
}

func (msg MessageS) GetKey() string {
	return msg.Args.Key
}

func (msg MessageS) GetValue() string {
	return msg.Args.Value
}

func (msg MessageS) GetClock() interface{} {
	return msg.LogicalClock
}

func (msg MessageS) GetOrderClient() int {
	return msg.Args.Timestamp
}
