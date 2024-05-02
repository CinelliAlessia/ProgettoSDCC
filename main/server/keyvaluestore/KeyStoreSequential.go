package keyvaluestore

import (
	"main/common"
	"main/server/message"
	"sync"
)

// KeyValueStoreSequential rappresenta il servizio di memorizzazione chiave-valore specializzato nel sequenziale
type KeyValueStoreSequential struct {
	Common KeyValueStore

	LogicalClock int // Orologio logico scalare
	mutexClock   sync.Mutex

	Queue      []msg.MessageS
	mutexQueue sync.Mutex // Mutex per proteggere l'accesso concorrente alla coda

	executeFunctionMutex sync.Mutex // Mutex aggiunto per evitare scheduling che interrompano l'invio a livello applicativo del messaggio

	ClientMaps map[string]*ClientMap // Mappa -> struttura dati che associa chiavi a valori
}

// ----- Consistenza Sequenziale ----- //

func NewKeyValueStoreSequential(idServer int) *KeyValueStoreSequential {
	kvSequential := &KeyValueStoreSequential{
		Common: KeyValueStore{
			Datastore: make(map[string]string),
		},
	}

	kvSequential.ClientMaps = make(map[string]*ClientMap)
	kvSequential.SetLogicalClock(0)                // Inizializzazione dell'orologio logico scalare
	kvSequential.SetQueue(make([]msg.MessageS, 0)) // Inizializzazione della coda
	kvSequential.SetIdServer(idServer)

	return kvSequential
}

func (kvs *KeyValueStoreSequential) SetDatastore(key string, value string) {
	kvs.Common.Datastore[key] = value
}

func (kvs *KeyValueStoreSequential) SetLogicalClock(logicalClock int) {
	kvs.LogicalClock = logicalClock
}

func (kvs *KeyValueStoreSequential) SetQueue(queue []msg.MessageS) {
	kvs.Queue = queue
}

func (kvs *KeyValueStoreSequential) SetIdServer(id int) {
	kvs.Common.Id = id
}

func (kvs *KeyValueStoreSequential) IncreaseRequestTsClient(args common.Args) {
	kvs.ClientMaps[args.GetIdClient()].MutexRequest.Lock()
	defer kvs.ClientMaps[args.GetIdClient()].MutexRequest.Unlock()
	kvs.ClientMaps[args.GetIdClient()].RequestTs = args.TimestampClient
}

func (kvs *KeyValueStoreSequential) GetRequestTsClient(id string) int {
	return kvs.ClientMaps[id].RequestTs
}

func (kvs *KeyValueStoreSequential) GetClock() int {
	return kvs.LogicalClock
}

func (kvs *KeyValueStoreSequential) GetDatastore() map[string]string {
	return kvs.Common.Datastore
}

func (kvs *KeyValueStoreSequential) GetQueue() []msg.MessageS {
	return kvs.Queue
}

func (kvs *KeyValueStoreSequential) GetIdServer() int {
	return kvs.Common.Id
}

/*
type orderingFIFO struct {
	receiveAssumeFIFO      int // Variabile per mantenere le richieste dal client in ordine FIFO
	receiveAssumeFIFOMutex sync.Mutex

	sendAssumeFIFO      [Common.Replicas]int // Variabile per mantenere le richieste dal client in ordine FIFO
	sendAssumeFIFOMutex sync.Mutex
}
{ // Questo in kvs
	receiveAssumeFIFO      int // Variabile per mantenere le richieste dal client in ordine FIFO
	receiveAssumeFIFOMutex sync.Mutex

	sendAssumeFIFO      int // Variabile per mantenere le richieste dal client in ordine FIFO //[Common.Replicas]
	sendAssumeFIFOMutex sync.Mutex
}


// receiveFIFOOrdered è una funzione di supporto utilizzata per mantenere l'illusione di una comunicazione FIFO
// Le richieste inviate dal client hanno un timestamp scalare crescente, ma non è garantito che le richieste
// arriveranno in ordine crescente, receiveFIFOOrdered si occupa di mantenere l'ordine delle richieste.
func (kvs *KeyValueStoreSequential) receiveFIFOOrdered(args Common.Args) bool {
	for {
		kvs.receiveAssumeFIFOMutex.Lock()
		if kvs.receiveAssumeFIFO == args.TimestampClient && args.Key != Common.EndKey {
			kvs.receiveAssumeFIFO++
			kvs.printDebugBlueArgs("RICEVUTO da client", args)
			kvs.receiveAssumeFIFOMutex.Unlock()
			return true
		} else if args.Key == Common.EndKey {
			kvs.printDebugBlueArgs("RICEVUTO da client", args)
			kvs.receiveAssumeFIFOMutex.Unlock()
			return true
		}
		kvs.receiveAssumeFIFOMutex.Unlock()
	}

}

func (kvs *KeyValueStoreSequential) sendFIFOOrderedToClient(msg MessageS) bool {
	for {
		kvs.sendAssumeFIFOMutex.Lock()
		if kvs.sendAssumeFIFO == msg.Args.TimestampClient && msg.Args.Key != Common.EndKey {
			kvs.sendAssumeFIFO++
			kvs.printGreen("ESEGUITO", msg)
			kvs.sendAssumeFIFOMutex.Unlock()
			return true
		} else if msg.Args.Key == Common.EndKey {
			kvs.sendAssumeFIFOMutex.Unlock()
			return true
		}
		kvs.sendAssumeFIFOMutex.Unlock()
	}
}
*/
