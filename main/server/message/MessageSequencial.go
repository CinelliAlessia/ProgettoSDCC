package msg

import (
	"main/common"
)

// ----- Consistenza Sequenziale ----- //

type MessageS struct {
	Common       MessageCommon
	LogicalClock int
	NumberAck    int
}

// NewMessageSeq crea un nuovo messaggio sequenziale, fa le veci di un costruttore
func NewMessageSeq(idSender int, typeOfMessage string, args common.Args, logicalClock int, numberAck int) *MessageS {
	msg := &MessageS{}

	msg.setIdMessage()
	msg.SetIdSender(idSender)
	msg.SetTypeOfMessage(typeOfMessage)
	msg.SetKey(args.GetKey())
	msg.SetValue(args.GetValue())
	msg.SetTimestampClient(args.GetTimestamp())
	msg.SetClock(logicalClock)
	msg.SetNumberAck(numberAck)

	return msg
}

// ----- Metodi SET ----- //

func (msg *MessageS) setIdMessage() {
	msg.Common.Id = common.GenerateUniqueID()
}

func (msg *MessageS) SetIdSender(idSender int) {
	msg.Common.IdSender = idSender
}

func (msg *MessageS) SetTypeOfMessage(typeOfMessage string) {
	msg.Common.TypeOfMessage = typeOfMessage
}

func (msg *MessageS) SetNumberAck(numberAck int) {
	msg.NumberAck = numberAck
}

func (msg *MessageS) SetClock(clock int) {
	msg.LogicalClock = clock
}

func (msg *MessageS) SetKey(key string) {
	msg.Common.Args.SetKey(key)
}

func (msg *MessageS) SetValue(value string) {
	msg.Common.Args.SetValue(value)
}

func (msg *MessageS) SetTimestampClient(timestamp int) {
	msg.Common.Args.SetTimestamp(timestamp)
}

/* METODI GET */

func (msg *MessageS) GetNumberAck() int {
	return msg.NumberAck
}

func (msg *MessageS) GetIdSender() int {
	return msg.Common.IdSender
}

func (msg *MessageS) GetIdMessage() string {
	return msg.Common.Id
}

func (msg *MessageS) GetTypeOfMessage() string {
	return msg.Common.TypeOfMessage
}

func (msg *MessageS) GetKey() string {
	return msg.Common.Args.GetKey()
}

func (msg *MessageS) GetValue() string {
	return msg.Common.Args.GetValue()
}

func (msg *MessageS) GetClock() int {
	return msg.LogicalClock
}

func (msg *MessageS) GetOrderClient() int {
	return msg.Common.Args.GetTimestamp()
}
