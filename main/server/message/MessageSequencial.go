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
	msg.SetClientID(args.GetClientID())

	msg.SetTypeOfMessage(typeOfMessage)
	msg.SetKey(args.GetKey())
	msg.SetValue(args.GetValue())

	msg.SetSendingFIFO(args.GetSendingFIFO())

	msg.SetClock(logicalClock)
	msg.SetNumberAck(numberAck)

	return msg
}

// Assegna un id univoco al messaggio
func (msg *MessageS) setIdMessage() {
	msg.Common.Id = common.GenerateUniqueID()
}

func (msg *MessageS) GetIdMessage() string {
	return msg.Common.Id
}

// SetIdSender Assegna l'id tra 0 e common.Replicas-1 del server che ha inviato il messaggio
func (msg *MessageS) SetIdSender(idSender int) {
	msg.Common.IdSender = idSender
}

func (msg *MessageS) GetIdSender() int {
	return msg.Common.IdSender
}

func (msg *MessageS) SetTypeOfMessage(typeOfMessage string) {
	msg.Common.TypeOfMessage = typeOfMessage
}

func (msg *MessageS) GetTypeOfMessage() string {
	return msg.Common.TypeOfMessage
}

func (msg *MessageS) SetNumberAck(numberAck int) {
	msg.NumberAck = numberAck
}

func (msg *MessageS) GetNumberAck() int {
	return msg.NumberAck
}

// SetClock Associa il clock logico al messaggio
func (msg *MessageS) SetClock(clock int) {
	msg.LogicalClock = clock
}

func (msg *MessageS) GetClock() int {
	return msg.LogicalClock
}

func (msg *MessageS) GetArgs() common.Args {
	return msg.Common.Args
}

func (msg *MessageS) SetKey(key string) {
	msg.Common.Args.SetKey(key)
}

func (msg *MessageS) GetKey() string {
	return msg.Common.Args.GetKey()
}

func (msg *MessageS) SetValue(value string) {
	msg.Common.Args.SetValue(value)
}

func (msg *MessageS) GetValue() string {
	return msg.Common.Args.GetValue()
}

func (msg *MessageS) SetClientID(client string) {
	msg.Common.Args.SetClientID(client)
}

func (msg *MessageS) GetClientID() string {
	return msg.Common.Args.GetClientID()
}

// ----- Assunzione FIFO Ordering -----

// SetSendingFIFO Associa il timestamp logico del client al messaggio
// - timestamp logico del client che lui associa alla richiesta per poi inviarlo al server
func (msg *MessageS) SetSendingFIFO(timestamp int) {
	msg.Common.Args.SetSendingFIFO(timestamp)
}

func (msg *MessageS) GetSendingFIFO() int {
	args := msg.GetArgs()
	return args.GetSendingFIFO()
}
