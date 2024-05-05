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

func (kvs *KeyValueStoreSequential) PutInDatastore(key string, value string) {
	kvs.Common.Datastore[key] = value
}

// SetDatastore imposta il datastore del server
func (kvs *KeyValueStoreSequential) SetDatastore(datastore map[string]string) {
	kvs.Common.Datastore = datastore
}

func (kvs *KeyValueStoreSequential) GetFromDatastore(key string) (string, error) {
	if value, ok := kvs.Common.Datastore[key]; ok {
		return value, nil
	}
	return "", errors.New("chiave non presente nel datastore")
}

func (kvs *KeyValueStoreSequential) GetDatastore() map[string]string {
	return kvs.Common.Datastore
}

// ----- Orologio Logico Scalare ----- //

func (kvs *KeyValueStoreSequential) SetClock(logicalClock int) {
	kvs.LogicalClock = logicalClock
}

// GetClock gestisce una chiamata RPC di un evento interno, genera un messaggio e gli allega il suo clock scalare.
func (kvs *KeyValueStoreSequential) GetClock() int {
	return kvs.LogicalClock
}

// ----- Coda ----- //

func (kvs *KeyValueStoreSequential) SetQueue(queue []commonMsg.MessageS) {
	kvs.Queue = queue
}
func (kvs *KeyValueStoreSequential) GetQueue() []commonMsg.MessageS {
	return kvs.Queue
}
func (kvs *KeyValueStoreSequential) GetMsgToQueue(index int) *commonMsg.MessageS {
	return &kvs.GetQueue()[index]
}

// ----- Id Server ----- //

func (kvs *KeyValueStoreSequential) SetIdServer(id int) {
	kvs.Common.Id = id
}
func (kvs *KeyValueStoreSequential) GetIdServer() int {
	return kvs.Common.Id
}

// ----- Client Map ----- //

func (kvs *KeyValueStoreSequential) NewClientMap(idClient string) {
	kvs.mutexMaps.Lock()
	defer kvs.mutexMaps.Unlock()

	kvs.ClientMaps[idClient] = &ClientMap{RequestTs: 0, ExecuteTs: 0}
}

func (kvs *KeyValueStoreSequential) GetClientMap(id string) (*ClientMap, bool) {
	val, ok := kvs.ClientMaps[id]
	return val, ok
}

func (kvs *KeyValueStoreSequential) SetRequestClient(id string, ts int) {
	val, _ := kvs.GetClientMap(id)
	val.SetRequestTs(ts)
}

func (kvs *KeyValueStoreSequential) IncreaseRequestTsClient(args common.Args) {
	kvs.SetRequestClient(args.GetIDClient(), args.GetTimestamp()+1)
}

// TODO: errore qui, -1 gestire
func (kvs *KeyValueStoreSequential) GetRequestTsClient(id string) (int, error) {
	if clientMap, ok := kvs.GetClientMap(id); ok {
		return clientMap.GetRequestTs(), nil
	}
	// Gestisci l'errore qui. Potresti restituire un valore predefinito o generare un errore.
	return -1, fmt.Errorf("key non presente")
}

// TODO : non usata
func (kvs *KeyValueStoreSequential) IncreaseExecuteTsClient(args common.Args) {
	clientMap := kvs.ClientMaps[args.GetIDClient()]
	clientMap.MutexExecute.Lock()

	executeTs := args.GetTimestamp() + 1
	clientMap.SetExecuteTs(executeTs)

	clientMap.MutexExecute.Unlock()
	kvs.ClientMaps[args.GetIDClient()] = clientMap
}

func (kvs *KeyValueStoreSequential) GetExecuteTsClient(id string) (int, error) {
	if clientMap, ok := kvs.ClientMaps[id]; ok {
		return clientMap.GetExecuteTs(), nil
	}
	// Gestisci l'errore qui. Potresti restituire un valore predefinito o generare un errore.
	return -1, fmt.Errorf("key non presente")
}

func (kvs *KeyValueStoreSequential) SetResponseOrderingFIFO() {
	kvs.mutexResponseOrderingFIFO.Lock()
	defer kvs.mutexResponseOrderingFIFO.Unlock()
	kvs.ResponseOrderingFIFO += 1
}

func (kvs *KeyValueStoreSequential) GetResponseOrderingFIFO() int {
	kvs.mutexResponseOrderingFIFO.Lock()
	defer kvs.mutexResponseOrderingFIFO.Unlock()

	return kvs.ResponseOrderingFIFO
}
