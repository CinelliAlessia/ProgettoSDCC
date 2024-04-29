package main

import "main/common"

// ----- Consistenza Causale ----- //

type MessageC struct {
	Common      MessageCommon
	VectorClock [common.Replicas]int // Orologio vettoriale
}

func newMessageCau(IdSender int, TypeOfMessage string, Args common.Args, vectorClock [common.Replicas]int) MessageC {
	message := MessageC{}

	message.Common.Id = common.GenerateUniqueID()
	message.Common.IdSender = IdSender
	message.Common.TypeOfMessage = TypeOfMessage
	message.Common.Args = Args
	message.VectorClock = vectorClock

	return message
}

// ----- Metodi GET ----- //

func (msg MessageC) GetIdSender() int {
	return msg.Common.IdSender
}

func (msg MessageC) GetIdMessage() string {
	return msg.Common.Id
}

func (msg MessageC) GetTypeOfMessage() string {
	return msg.Common.TypeOfMessage
}

func (msg MessageC) GetKey() string {
	return msg.Common.Args.Key
}

func (msg MessageC) GetValue() string {
	return msg.Common.Args.Value
}

func (msg MessageC) GetClock() [common.Replicas]int {
	return msg.VectorClock
}

func (msg MessageC) GetOrderClient() int {
	return msg.Common.Args.Timestamp
}

/* ----- Metodi SET ----- //

func SetTypeOfMessage(msg *MessageC, typeOfMessage string) {
	msg.Common.TypeOfMessage = typeOfMessage
}

func SetKey(msg *MessageC, key string) {
	msg.Common.Args.Key = key
}
*/
