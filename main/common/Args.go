package common

import (
	"sync"
)

// Args rappresenta gli argomenti delle chiamate RPC
type Args struct {
	Key         string
	Value       string
	SendingFIFO int    // Assunzione FIFO Ordering
	ClientID    string // Per identificare il client nella operazione RPC
	SafeBool    *SafeBool
}

/* ARGS STRUCT */

// NewArgs crea una nuova struttura Args, assegnando a ClientID un uuid
func NewArgs(timestamp int, key string, values ...string) Args {
	args := Args{}

	args.SetSendingFIFO(timestamp)
	args.SetKey(key)
	args.SetClientID(GenerateUniqueID())

	if len(values) > 0 {
		args.SetValue(values[0])
	}
	return args
}

// SetKey imposta la chiave del database come parametro della struttura Args
func (args *Args) SetKey(key string) {
	args.Key = key
}

// GetKey restituisce la chiave del database come parametro della struttura Args
func (args *Args) GetKey() string {
	return args.Key
}

// SetValue imposta il valore del database come parametro della struttura Args
func (args *Args) SetValue(value string) {
	args.Value = value
}

// GetValue restituisce il valore del database come parametro della struttura Args
func (args *Args) GetValue() string {
	return args.Value
}

// SetSendingFIFO imposta il timestamp di invio della richiesta
//
//	Rappresenta l'ordine con cui la richiesta deve essere processata dal server
func (args *Args) SetSendingFIFO(timestamp int) {
	args.SendingFIFO = timestamp
}

// GetSendingFIFO restituisce il timestamp di invio della richiesta
//
//	Rappresenta l'ordine con cui la richiesta deve essere processata dal server
func (args *Args) GetSendingFIFO() int {
	return args.SendingFIFO
}

// SetClientID imposta l'identificativo del client che ha effettuato la richiesta
func (args *Args) SetClientID(id string) {
	args.ClientID = id
}

// GetClientID restituisce l'identificativo del client che ha effettuato la richiesta
func (args *Args) GetClientID() string {
	return args.ClientID
}

func (args *Args) ConfigureSafeBool() {
	args.SafeBool = NewSafeBool()
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

func (args *Args) WaitCondition() bool {
	boolean := args.SafeBool.Wait()
	return boolean
}

func (args *Args) SetCondition(b bool) {
	args.SafeBool.Set(b)
}
