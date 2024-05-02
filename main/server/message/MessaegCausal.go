package msg

import (
	"main/common"
)

// ----- Consistenza causale ----- //

type MessageC struct {
	Common      MessageCommon
	VectorClock [common.Replicas]int // Orologio vettoriale
}

func NewMessageCau(idSender int, typeOfMessage string, args common.Args, vectorClock [common.Replicas]int) *MessageC {
	msg := &MessageC{}

	msg.setIdMessage()
	msg.SetIdSender(idSender)
	msg.SetTypeOfMessage(typeOfMessage)
	msg.SetKey(args.GetKey())
	msg.SetValue(args.GetValue())
	msg.SetTimestampClient(args.GetTimestamp())
	msg.SetClock(vectorClock)
	return msg
}

// ----- Metodi SET ----- //

func (msg *MessageC) setIdMessage() {
	msg.Common.Id = common.GenerateUniqueID()
}

func (msg *MessageC) SetIdSender(idSender int) {
	msg.Common.IdSender = idSender
}

func (msg *MessageC) SetTypeOfMessage(typeOfMessage string) {
	msg.Common.TypeOfMessage = typeOfMessage
}

func (msg *MessageC) SetClock(vectorClock [common.Replicas]int) {
	msg.VectorClock = vectorClock
}

func (msg *MessageC) SetValueClock(index int, value int) {
	msg.VectorClock[index] = value
}

func (msg *MessageC) SetKey(key string) {
	msg.Common.Args.SetKey(key)
}

func (msg *MessageC) SetValue(value string) {
	msg.Common.Args.SetValue(value)
}

func (msg *MessageC) SetTimestampClient(timestamp int) {
	msg.Common.Args.SetTimestamp(timestamp)
}

// ----- Metodi GET ----- //

func (msg *MessageC) GetIdSender() int {
	return msg.Common.IdSender
}

func (msg *MessageC) GetIdMessage() string {
	return msg.Common.Id
}

func (msg *MessageC) GetTypeOfMessage() string {
	return msg.Common.TypeOfMessage
}

func (msg *MessageC) GetKey() string {
	return msg.Common.Args.GetKey()
}

func (msg *MessageC) GetValue() string {
	return msg.Common.Args.GetValue()
}

func (msg *MessageC) GetOrderClient() int {
	return msg.Common.Args.GetTimestamp()
}

func (msg *MessageC) GetClock() [common.Replicas]int {
	return msg.VectorClock
}
