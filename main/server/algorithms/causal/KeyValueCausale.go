package causal

import (
	"fmt"
	"main/common"
	"main/server/algorithms"
	"main/server/message"
	"sync"
)

// KeyValueStoreCausale rappresenta la struttura di memorizzazione chiave-valore per garantire consistenza causale
type KeyValueStoreCausale struct {
	Common algorithms.KeyValueStore

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
		Common: algorithms.KeyValueStore{
			Datastore:  make(map[string]string),
			ClientMaps: make(map[string]*algorithms.ClientMap),
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

// ----- Operazioni Mappa Client ----- //
// ----- Client Map ----- //

// NewClientMap crea una nuova mappa client per tenere conto dell'assunzione FIFO Ordering dei messaggi
func (kvc *KeyValueStoreCausale) NewClientMap(idClient string) {
	kvc.Common.MutexMaps.Lock()
	defer kvc.Common.MutexMaps.Unlock()

	kvc.Common.ClientMaps[idClient] = &algorithms.ClientMap{
		ReceiveOrderingFIFO:  0,
		ResponseOrderingFIFO: 0,
	}
}

func (kvc *KeyValueStoreCausale) ExistClient(idClient string) bool {
	_, ok := kvc.Common.ClientMaps[idClient]
	return ok
}

// GetClientMap restituisce la mappa client associata a un client, identificato da un id univoco preso come argomento
func (kvc *KeyValueStoreCausale) GetClientMap(id string) (*algorithms.ClientMap, bool) {
	val, ok := kvc.Common.ClientMaps[id]
	return val, ok
}

// SetRequestClient imposta il timestamp di richiesta di un client
func (kvc *KeyValueStoreCausale) SetRequestClient(id string, ts int) {
	val, _ := kvc.GetClientMap(id)
	val.SetReceiveOrderingFIFO(ts)
}

func (kvc *KeyValueStoreCausale) IncreaseReceiveTsClient(args common.Args) {
	kvc.SetRequestClient(args.GetClientID(), args.GetSendingFIFO()+1)
}

func (kvc *KeyValueStoreCausale) GetReceiveTsFromClient(id string) (int, error) {
	if clientMap, ok := kvc.GetClientMap(id); ok {
		return clientMap.GetReceiveOrderingFIFO(), nil
	}
	// Gestisci l'errore qui. Potresti restituire un valore predefinito o generare un errore.
	return -1, fmt.Errorf("key non presente")
}

func (kvc *KeyValueStoreCausale) SetResponseOrderingFIFO(id string, ts int) {
	val, _ := kvc.GetClientMap(id)
	val.SetResponseOrderingFIFO(ts)
}

func (kvc *KeyValueStoreCausale) IncreaseResponseOrderingFIFO(ClientID string) {
	kvc.SetResponseOrderingFIFO(ClientID, kvc.GetResponseOrderingFIFO(ClientID)+1)
}

func (kvc *KeyValueStoreCausale) GetResponseOrderingFIFO(ClientID string) int {
	val, _ := kvc.GetClientMap(ClientID)
	return val.GetResponseOrderingFIFO()
}

// ----- Mutex ----- //

func (kvc *KeyValueStoreCausale) LockMutexMessage(ClientID string) {
	val, _ := kvc.GetClientMap(ClientID)
	val.LockMutexMessage()
}

func (kvc *KeyValueStoreCausale) UnlockMutexMessage(ClientID string) {
	val, _ := kvc.GetClientMap(ClientID)
	val.UnlockMutexMessage()
}
