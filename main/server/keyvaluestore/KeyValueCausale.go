package keyvaluestore

import (
	"main/common"
	"main/server/message"
	"sync"
)

// KeyValueStoreCausale rappresenta la struttura di memorizzazione chiave-valore per garantire consistenza causale
type KeyValueStoreCausale struct {
	Common KeyValueStore

	VectorClock [common.Replicas]int // Orologio vettoriale
	mutexClock  sync.Mutex

	Queue      []commonMsg.MessageC
	mutexQueue sync.Mutex // Mutex per proteggere l'accesso concorrente alla coda

	executeFunctionMutex sync.Mutex // Mutex aggiunto per evitare scheduling che interrompano l'invio a livello applicativo del messaggio
}

// ----- Consistenza causale ----- //

// NewKeyValueStoreCausal crea un nuovo KeyValueStoreCausale, inizializzando l'orologio vettoriale e la coda
// prende come argomento l'id del server replica, da 0 a common.Replicas-1
func NewKeyValueStoreCausal(idServer int) *KeyValueStoreCausale {
	kvc := &KeyValueStoreCausale{
		Common: KeyValueStore{
			Datastore: make(map[string]string),
		},
	}

	for i := 0; i < common.Replicas; i++ {
		kvc.SetVectorClock(0, 0) // Inizializzazione dell'orologio logico vettoriale
	}

	kvc.SetQueue(make([]commonMsg.MessageC, 0)) // Inizializzazione della coda
	kvc.SetIdServer(idServer)

	return kvc
}

func (kvc *KeyValueStoreCausale) SetDatastore(key string, value string) {
	kvc.Common.Datastore[key] = value
}

func (kvc *KeyValueStoreCausale) GetDatastore() map[string]string {
	return kvc.Common.Datastore
}

func (kvc *KeyValueStoreCausale) DeleteFromDatastore(key string) {
	delete(kvc.Common.Datastore, key)
}

func (kvc *KeyValueStoreCausale) SetVectorClock(index int, value int) {
	kvc.VectorClock[index] = value
}

func (kvc *KeyValueStoreCausale) GetClock() [common.Replicas]int {
	return kvc.VectorClock
}

func (kvc *KeyValueStoreCausale) GetClockIDServer(id int) int {
	return kvc.VectorClock[id]
}

func (kvc *KeyValueStoreCausale) SetQueue(queue []commonMsg.MessageC) {
	kvc.Queue = queue
}

func (kvc *KeyValueStoreCausale) GetQueue() []commonMsg.MessageC {
	return kvc.Queue
}

func (kvc *KeyValueStoreCausale) SetIdServer(id int) {
	kvc.Common.Id = id
}

func (kvc *KeyValueStoreCausale) GetIdServer() int {
	return kvc.Common.Id
}
