package sequential

import (
	"errors"
	"fmt"
	"main/server/algorithms"
	"main/server/message"
	"sync"
)

// ----- CONSISTENZA SEQUENZIALE ----- //

// KeyValueStoreSequential rappresenta il servizio di memorizzazione chiave-valore specializzato nel sequenziale
type KeyValueStoreSequential struct {
	Common algorithms.KeyValueStore

	LogicalClock int // Orologio logico scalare
	mutexClock   sync.Mutex

	Queue      []commonMsg.MessageS
	mutexQueue sync.Mutex // Mutex per proteggere l'accesso concorrente alla coda

	executeFunctionMutex sync.Mutex // Mutex aggiunto per evitare scheduling che interrompano l'invio a livello applicativo del messaggio
}

// NewKeyValueStoreSequential crea un nuovo KeyValueStoreSequential, inizializzando l'orologio logico scalare e la coda
// prende come argomento l'id del server replica, da 0 a common.Replicas-1
func NewKeyValueStoreSequential(idServer int) *KeyValueStoreSequential {
	kvSequential := &KeyValueStoreSequential{
		Common: algorithms.KeyValueStore{
			Datastore:  make(map[string]string),
			ClientMaps: make(map[string]*algorithms.ClientMap),
		},
	}

	kvSequential.SetClock(0)                             // Inizializzazione dell'orologio logico scalare
	kvSequential.SetQueue(make([]commonMsg.MessageS, 0)) // Inizializzazione della coda
	kvSequential.SetIdServer(idServer)

	return kvSequential
}

// ----- Operazioni Datastore ----- //

// PutInDatastore inserisce una coppia chiave-valore nel datastore del server
// prende come argomenti la chiave e il valore da inserire
func (kvs *KeyValueStoreSequential) PutInDatastore(key string, value string) {
	kvs.Common.Datastore[key] = value
}

// SetDatastore imposta il datastore del server
func (kvs *KeyValueStoreSequential) SetDatastore(datastore map[string]string) {
	kvs.Common.Datastore = datastore
}

// DeleteFromDatastore elimina una coppia chiave-valore dal datastore del server
func (kvs *KeyValueStoreSequential) DeleteFromDatastore(key string) {
	delete(kvs.Common.Datastore, key)
}

// GetFromDatastore restituisce il valore associato a una chiave nel datastore del server
func (kvs *KeyValueStoreSequential) GetFromDatastore(key string) (string, error) {
	if value, ok := kvs.Common.Datastore[key]; ok {
		return value, nil
	}
	return "", errors.New("chiave non presente nel datastore")
}

// GetDatastore restituisce il datastore del server
func (kvs *KeyValueStoreSequential) GetDatastore() map[string]string {
	return kvs.Common.Datastore
}

// ----- Orologio Logico Scalare ----- //

// SetClock imposta l'orologio logico scalare del server
func (kvs *KeyValueStoreSequential) SetClock(logicalClock int) {
	kvs.LogicalClock = logicalClock
}

// GetClock restituisce l'orologio logico scalare del server
func (kvs *KeyValueStoreSequential) GetClock() int {
	return kvs.LogicalClock
}

// ----- Coda ----- //

// SetQueue imposta la coda del server
func (kvs *KeyValueStoreSequential) SetQueue(queue []commonMsg.MessageS) {
	kvs.Queue = queue
}

// GetQueue restituisce la coda del server
func (kvs *KeyValueStoreSequential) GetQueue() []commonMsg.MessageS {
	return kvs.Queue
}

// GetMsgFromQueue restituisce un messaggio dalla coda del server
// prende come argomento l'indice del messaggio da restituire
func (kvs *KeyValueStoreSequential) GetMsgFromQueue(index int) *commonMsg.MessageS {
	return &kvs.GetQueue()[index]
}

// ----- Id Server ----- //

// SetIdServer imposta l'id del server
func (kvs *KeyValueStoreSequential) SetIdServer(id int) {
	kvs.Common.Id = id
}

// GetIdServer restituisce l'id del server
func (kvs *KeyValueStoreSequential) GetIdServer() int {
	return kvs.Common.Id
}

// ----- Client Map ----- //

// NewClientMap crea una nuova entity nella mappa client per tenere conto dell'assunzione FIFO Ordering dei messaggi
func (kvs *KeyValueStoreSequential) NewClientMap(idClient string) {
	kvs.Common.MutexMaps.Lock()
	defer kvs.Common.MutexMaps.Unlock()

	kvs.Common.ClientMaps[idClient] = &algorithms.ClientMap{
		ReceiveOrderingFIFO:  0,
		ResponseOrderingFIFO: 0,
	}
}

func (kvs *KeyValueStoreSequential) ExistClient(idClient string) bool {
	_, ok := kvs.GetClientMap(idClient)
	return ok
}

// GetClientMap restituisce la mappa client associata a un client, identificato da un id univoco preso come argomento
func (kvs *KeyValueStoreSequential) GetClientMap(id string) (*algorithms.ClientMap, bool) {
	kvs.Common.MutexMaps.Lock()
	defer kvs.Common.MutexMaps.Unlock()

	val, ok := kvs.Common.ClientMaps[id]
	return val, ok
}

// SetRequestClient imposta il timestamp di richiesta di un client
func (kvs *KeyValueStoreSequential) SetRequestClient(id string, ts int) {
	val, _ := kvs.GetClientMap(id)
	val.SetReceiveOrderingFIFO(ts)
}

func (kvs *KeyValueStoreSequential) GetReceiveTsFromClient(id string) (int, error) {
	if clientMap, ok := kvs.GetClientMap(id); ok {
		return clientMap.GetReceiveOrderingFIFO(), nil
	}
	// Gestisci l'errore qui. Potresti restituire un valore predefinito o generare un errore.
	return -1, fmt.Errorf("key non presente")
}

func (kvs *KeyValueStoreSequential) SetResponseOrderingFIFO(id string, ts int) {
	val, _ := kvs.GetClientMap(id)
	val.SetResponseOrderingFIFO(ts)
}

func (kvs *KeyValueStoreSequential) GetResponseOrderingFIFO(ClientID string) int {
	val, _ := kvs.GetClientMap(ClientID)
	return val.GetResponseOrderingFIFO()
}

// ----- Mutex ----- //

func (kvs *KeyValueStoreSequential) LockMutexMessage(ClientID string) {
	val, _ := kvs.GetClientMap(ClientID)
	val.LockMutexMessage()
}

func (kvs *KeyValueStoreSequential) UnlockMutexMessage(ClientID string) {
	val, _ := kvs.GetClientMap(ClientID)
	val.UnlockMutexMessage()
}

func (kvs *KeyValueStoreSequential) LockMutexMaps() {
	kvs.Common.MutexMaps.Lock()
}

func (kvs *KeyValueStoreSequential) UnlockMutexMaps() {
	kvs.Common.MutexMaps.Unlock()
}
