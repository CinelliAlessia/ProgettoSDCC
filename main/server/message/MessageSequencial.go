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

// GetIdMessage restituisce l'id (uuid) del messaggio
func (msg *MessageS) GetIdMessage() string {
	return msg.Common.Id
}

// SetIdSender Assegna l'id tra 0 e common.Replicas-1 del server che ha inviato il messaggio
func (msg *MessageS) SetIdSender(idSender int) {
	msg.Common.IdSender = idSender
}

// GetIdSender Restituisce l'id tra 0 e common.Replicas-1 del server che ha inviato il messaggio
func (msg *MessageS) GetIdSender() int {
	return msg.Common.IdSender
}

// SetTypeOfMessage Associa il tipo di messaggio al messaggio (put, get, delete)
func (msg *MessageS) SetTypeOfMessage(typeOfMessage string) {
	msg.Common.TypeOfMessage = typeOfMessage
}

// GetTypeOfMessage Restituisce il tipo di messaggio associato al messaggio (put, get, delete)
func (msg *MessageS) GetTypeOfMessage() string {
	return msg.Common.TypeOfMessage
}

// SetNumberAck Associa il numero di ack ricevuti dal server relativi al messaggio
func (msg *MessageS) SetNumberAck(numberAck int) {
	msg.NumberAck = numberAck
}

// GetNumberAck Restituisce il numero di ack ricevuti dal server associati al messaggio
func (msg *MessageS) GetNumberAck() int {
	return msg.NumberAck
}

// SetClock Associa il clock logico scalare del server al messaggio
func (msg *MessageS) SetClock(clock int) {
	msg.LogicalClock = clock
}

// GetClock Restituisce il clock logico scalare del server associato al messaggio
func (msg *MessageS) GetClock() int {
	return msg.LogicalClock
}

// ----- ARGS ----- //

// GetArgs Restituisce gli argomenti del messaggio al messaggio
func (msg *MessageS) GetArgs() common.Args {
	return msg.Common.Args
}

// SetKey Associa al messaggio la chiave appartenente alla richiesta
func (msg *MessageS) SetKey(key string) {
	msg.Common.Args.SetKey(key)
}

// GetKey Restituisce la chiave associata al messaggio
func (msg *MessageS) GetKey() string {
	return msg.Common.Args.GetKey()
}

// SetValue Associa al messaggio il valore appartenente alla richiesta
func (msg *MessageS) SetValue(value string) {
	msg.Common.Args.SetValue(value)
}

// GetValue Restituisce il valore associato al messaggio
func (msg *MessageS) GetValue() string {
	return msg.Common.Args.GetValue()
}

// SetClientID Associa al messaggio l'id del client che ha effettuato la richiesta
func (msg *MessageS) SetClientID(client string) {
	msg.Common.Args.SetClientID(client)
}

// GetClientID Restituisce l'id del client associato al messaggio
func (msg *MessageS) GetClientID() string {
	return msg.Common.Args.GetClientID()
}

// ----- Assunzione FIFO Ordering -----

// SetSendingFIFO Associa il timestamp logico del client al messaggio
// - timestamp logico del client che lui associa alla richiesta per poi inviarlo al server
func (msg *MessageS) SetSendingFIFO(timestamp int) {
	msg.Common.Args.SetSendingFIFO(timestamp)
}

// GetSendingFIFO Restituisce il timestamp logico del client associato al messaggio
func (msg *MessageS) GetSendingFIFO() int {
	args := msg.GetArgs()
	return args.GetSendingFIFO()
}
