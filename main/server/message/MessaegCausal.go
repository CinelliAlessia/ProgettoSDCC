package commonMsg

import (
	"main/common"
)

// ----- Consistenza causale ----- //

type MessageC struct {
	Common      MessageCommon
	VectorClock [common.Replicas]int // Orologio vettoriale
}

// NewMessageC Crea un nuovo messaggio per la garanzia di consistenza causale
func NewMessageC(idSender int, typeOfMessage string, args common.Args, vectorClock [common.Replicas]int) *MessageC {
	msg := &MessageC{}

	msg.setIdMessage()
	msg.SetSenderID(idSender)
	msg.SetClientID(args.GetClientID())

	msg.SetTypeOfMessage(typeOfMessage)
	msg.SetKey(args.GetKey())
	msg.SetValue(args.GetValue())

	msg.SetSendingFIFO(args.GetSendingFIFO())

	msg.SetClock(vectorClock)
	return msg
}

// setIdMessage Imposta l'identificativo univoco del messaggio
func (msg *MessageC) setIdMessage() {
	msg.Common.Id = common.GenerateUniqueID()
}

// GetIdMessage Restituisce l'identificativo univoco del messaggio
func (msg *MessageC) GetIdMessage() string {
	return msg.Common.Id
}

// SetSenderID Imposta l'identificativo del mittente del messaggio
func (msg *MessageC) SetSenderID(senderID int) {
	msg.Common.SenderID = senderID
}

// GetSenderID Restituisce l'identificativo del mittente del messaggio
func (msg *MessageC) GetSenderID() int {
	return msg.Common.SenderID
}

// SetTypeOfMessage Imposta il tipo di messaggio (put, get, delete)
func (msg *MessageC) SetTypeOfMessage(typeOfMessage string) {
	msg.Common.TypeOfMessage = typeOfMessage
}

// GetTypeOfMessage Restituisce il tipo di messaggio (put, get, delete)
func (msg *MessageC) GetTypeOfMessage() string {
	return msg.Common.TypeOfMessage
}

// SetClock Imposta l'orologio vettoriale nel messaggio
func (msg *MessageC) SetClock(vectorClock [common.Replicas]int) {
	msg.VectorClock = vectorClock
}

// SetValueClock Imposta il valore dell'orologio vettoriale in posizione index
//   - index: posizione dell'orologio vettoriale da modificare
//   - value: valore da assegnare all'orologio vettoriale in posizione index
func (msg *MessageC) SetValueClock(index int, value int) {
	msg.VectorClock[index] = value
}

// GetClock Restituisce l'orologio vettoriale del messaggio
func (msg *MessageC) GetClock() [common.Replicas]int {
	return msg.VectorClock
}

// SetKey Imposta la chiave del messaggio
func (msg *MessageC) SetKey(key string) {
	msg.Common.Args.SetKey(key)
}

// GetKey Restituisce la chiave del messaggio
func (msg *MessageC) GetKey() string {
	return msg.Common.Args.GetKey()
}

// SetValue Imposta il valore del messaggio
func (msg *MessageC) SetValue(value string) {
	msg.Common.Args.SetValue(value)
}

// GetValue Restituisce il valore del messaggio
func (msg *MessageC) GetValue() string {
	return msg.Common.Args.GetValue()
}

// SetClientID Imposta l'identificativo del client che ha inviato il messaggio
func (msg *MessageC) SetClientID(client string) {
	msg.Common.Args.SetClientID(client)
}

// GetClientID Restituisce l'identificativo del client che ha inviato il messaggio
func (msg *MessageC) GetClientID() string {
	return msg.Common.Args.GetClientID()
}

// SetSendingFIFO Imposta il timestamp di invio del messaggio
func (msg *MessageC) SetSendingFIFO(timestamp int) {
	msg.Common.Args.SetSendingFIFO(timestamp)
}

// GetSendingFIFO Restituisce il timestamp di invio del messaggio
func (msg *MessageC) GetSendingFIFO() int {
	return msg.Common.Args.GetSendingFIFO()
}
