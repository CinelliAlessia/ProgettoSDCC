package keyvaluestore

import (
	"errors"
	"fmt"
	"main/common"
	"main/server/message"
	"sync"
)

// KeyValueStoreSequential rappresenta il servizio di memorizzazione chiave-valore specializzato nel sequenziale
type KeyValueStoreSequential struct {
	Common KeyValueStore

	LogicalClock int // Orologio logico scalare
	mutexClock   sync.Mutex

	Queue      []commonMsg.MessageS
	mutexQueue sync.Mutex // Mutex per proteggere l'accesso concorrente alla coda

	executeFunctionMutex sync.Mutex // Mutex aggiunto per evitare scheduling che interrompano l'invio a livello applicativo del messaggio

	ClientMaps map[string]*ClientMap // Mappa -> struttura dati che associa chiavi a valori
	mutexMaps  sync.Mutex            // Mutex per proteggere l'accesso concorrente alla mappa

	ResponseOrderingFIFO      int
	mutexResponseOrderingFIFO sync.Mutex
}

// ----- CONSISTENZA SEQUENZIALE ----- //

// NewKeyValueStoreSequential crea un nuovo KeyValueStoreSequential, inizializzando l'orologio logico scalare e la coda
// prende come argomento l'id del server replica, da 0 a common.Replicas-1
func NewKeyValueStoreSequential(idServer int) *KeyValueStoreSequential {
	kvSequential := &KeyValueStoreSequential{
		Common: KeyValueStore{
			Datastore: make(map[string]string),
		},
	}

	kvSequential.ClientMaps = make(map[string]*ClientMap)
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

// NewClientMap crea una nuova mappa client per tenere conto dell'assunzione FIFO Ordering dei messaggi
func (kvs *KeyValueStoreSequential) NewClientMap(idClient string) {
	kvs.mutexMaps.Lock()
	defer kvs.mutexMaps.Unlock()

	kvs.ClientMaps[idClient] = &ClientMap{RequestTs: 0}
}

// GetClientMap restituisce la mappa client associata a un client, identificato da un id univoco preso come argomento
func (kvs *KeyValueStoreSequential) GetClientMap(id string) (*ClientMap, bool) {
	val, ok := kvs.ClientMaps[id]
	return val, ok
}

// SetRequestClient imposta il timestamp di richiesta di un client
func (kvs *KeyValueStoreSequential) SetRequestClient(id string, ts int) {
	val, _ := kvs.GetClientMap(id)
	val.SetRequestTs(ts)
}

func (kvs *KeyValueStoreSequential) IncreaseRequestTsClient(args common.Args) {
	kvs.SetRequestClient(args.GetIDClient(), args.GetSendingFIFO()+1)
}

func (kvs *KeyValueStoreSequential) GetRequestTsClient(id string) (int, error) {
	if clientMap, ok := kvs.GetClientMap(id); ok {
		return clientMap.GetRequestTs(), nil
	}
	// Gestisci l'errore qui. Potresti restituire un valore predefinito o generare un errore.
	return -1, fmt.Errorf("key non presente")
}

func (kvs *KeyValueStoreSequential) IncreaseResponseOrderingFIFO() {
	kvs.mutexResponseOrderingFIFO.Lock()
	defer kvs.mutexResponseOrderingFIFO.Unlock()
	kvs.ResponseOrderingFIFO += 1
}

func (kvs *KeyValueStoreSequential) GetResponseOrderingFIFO() int {
	kvs.mutexResponseOrderingFIFO.Lock()
	defer kvs.mutexResponseOrderingFIFO.Unlock()

	return kvs.ResponseOrderingFIFO
}
