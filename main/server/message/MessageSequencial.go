package commonMsg

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
	msg.SetClientID(args.GetIDClient())

	msg.SetTypeOfMessage(typeOfMessage)
	msg.SetKey(args.GetKey())
	msg.SetValue(args.GetValue())

	msg.SetTimestampClient(args.GetTimestamp())

	msg.SetClock(logicalClock)
	msg.SetNumberAck(numberAck)

	return msg
}

// ----- Metodi SET ----- //

// Assegna un id univoco al messaggio
func (msg *MessageS) setIdMessage() {
	msg.Common.Id = common.GenerateUniqueID()
}

// SetIdSender Assegna l'id tra 0 e common.Replicas-1 del server che ha inviato il messaggio
func (msg *MessageS) SetIdSender(idSender int) {
	msg.Common.IdSender = idSender
}

func (msg *MessageS) SetTypeOfMessage(typeOfMessage string) {
	msg.Common.TypeOfMessage = typeOfMessage
}

func (msg *MessageS) SetNumberAck(numberAck int) {
	msg.NumberAck = numberAck
}

// SetClock Associa il clock logico al messaggio
func (msg *MessageS) SetClock(clock int) {
	msg.LogicalClock = clock
}

func (msg *MessageS) SetKey(key string) {
	msg.Common.Args.SetKey(key)
}

func (msg *MessageS) SetValue(value string) {
	msg.Common.Args.SetValue(value)
}

// SetTimestampClient Associa il timestamp logico del client al messaggio
// - timestamp logico del client che lui associa alla richiesta per poi inviarlo al server
func (msg *MessageS) SetTimestampClient(timestamp int) {
	msg.Common.Args.SetTimestamp(timestamp)
}

func (msg *MessageS) SetClientID(client string) {
	msg.Common.Args.SetIDClient(client)
}

/* METODI GET */

func (msg *MessageS) GetArgs() common.Args {
	return msg.Common.Args
}

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
	args := msg.GetArgs()
	return args.GetTimestamp()
}

func (msg *MessageS) GetIdClient() string {
	return msg.Common.Args.GetIDClient()
}
