package main

import "main/common"

type MessageS struct {
	Common       MessageCommon
	LogicalClock int
	NumberAck    int
}

// newMessageSeq crea un nuovo messaggio sequenziale, fa le veci di un costruttore
func newMessageSeq(IdSender int, TypeOfMessage string, Args common.Args, LogicalClock int, NumberAck int) MessageS {
	msg := &MessageS{}

	setIdMessage(msg)
	SetIdSender(msg, IdSender)
	SetTypeOfMessage(msg, TypeOfMessage)
	SetKey(msg, Args.Key)
	SetValue(msg, Args.Value)
	SetTimestampClient(msg, Args.Timestamp)
	SetClock(msg, LogicalClock)
	SetNumberAck(msg, NumberAck)

	return *msg
}

// ----- Consistenza Sequenziale ----- //

func setIdMessage(msg *MessageS) {
	msg.Common.Id = common.GenerateUniqueID()
}

func SetIdSender(msg *MessageS, idSender int) {
	msg.Common.IdSender = idSender
}

func SetTypeOfMessage(msg *MessageS, typeOfMessage string) {
	msg.Common.TypeOfMessage = typeOfMessage
}

func SetNumberAck(msg *MessageS, numberAck int) {
	msg.NumberAck = numberAck
}

func SetClock(msg *MessageS, clock int) {
	msg.LogicalClock = clock
}

func SetKey(msg *MessageS, key string) {
	msg.Common.Args.Key = key
}

func SetValue(msg *MessageS, value string) {
	msg.Common.Args.Value = value
}

func SetTimestampClient(msg *MessageS, timestamp int) {
	msg.Common.Args.Timestamp = timestamp
}

func (msg MessageS) GetNumberAck() int {
	return msg.NumberAck
}

func (msg MessageS) GetIdSender() int {
	return msg.Common.IdSender
}

func (msg MessageS) GetIdMessage() string {
	return msg.Common.Id
}

func (msg MessageS) GetTypeOfMessage() string {
	return msg.Common.TypeOfMessage
}

func (msg MessageS) GetKey() string {
	return msg.Common.Args.Key
}

func (msg MessageS) GetValue() string {
	return msg.Common.Args.Value
}

func (msg MessageS) GetClock() int {
	return msg.LogicalClock
}

func (msg MessageS) GetOrderClient() int {
	return msg.Common.Args.Timestamp
}
