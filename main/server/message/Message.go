package commonMsg

import (
	"main/common"
	"sync"
)

type Message interface {
	GetIdMessage() string
	GetSenderID() int

	GetTypeOfMessage() string

	GetKey() string
	GetValue() string

	GetSendingFIFO() int
}

// MessageCommon rappresenta la struttura dati comune a tutti i messaggi
type MessageCommon struct {
	Id       string // Id del messaggio stesso
	SenderID int    // SenderID rappresenta l'indice del server che invia il messaggio

	TypeOfMessage string
	Args          common.Args

	SafeBool *SafeBool
}

type SafeBool struct {
	mutexCondition sync.Mutex
	cond           *sync.Cond
	Value          bool
}

func NewSafeBool() *SafeBool {
	sb := &SafeBool{}
	sb.cond = sync.NewCond(&sb.mutexCondition)
	return sb
}

func (sb *SafeBool) Set(value bool) {
	sb.mutexCondition.Lock()
	defer sb.mutexCondition.Unlock()
	sb.Value = value
	sb.cond.Broadcast()
}

func (sb *SafeBool) Wait() bool {
	sb.mutexCondition.Lock()
	defer sb.mutexCondition.Unlock()
	for !sb.Value {
		sb.cond.Wait()
	}
	return sb.Value
}
