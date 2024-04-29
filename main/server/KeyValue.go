package main

import (
	"main/common"
	"main/server/message"
	"sync"
)

type ClockServer interface {
	GetDatastore() map[string]string
	GetIdServer() int
}

// ClientMap rappresenta la struttura dati per memorizzare i timestamp delle richieste dei client, cosi da realizzare
// le assunzioni di comunicazione FIFO
// TODO: Domanda, un client contatta sempre la stessa replica? altrimenti cosi non mi funziona
type ClientMap struct {
	requestTs int // Timestamp della richiesta ricevuta dal client
	executeTs int // Timestamp di esecuzione della richiesta
}

// KeyValueStoreCausale rappresenta la struttura di memorizzazione chiave-valore per garantire consistenza causale
type KeyValueStoreCausale struct {
	Common KeyValueStore

	VectorClock [common.Replicas]int // Orologio vettoriale
	mutexClock  sync.Mutex

	Queue      []msg.MessageC
	mutexQueue sync.Mutex // Mutex per proteggere l'accesso concorrente alla coda

	executeFunctionMutex sync.Mutex // Mutex aggiunto per evitare scheduling che interrompano l'invio a livello applicativo del messaggio
}

// KeyValueStoreSequential rappresenta il servizio di memorizzazione chiave-valore specializzato nel sequenziale
type KeyValueStoreSequential struct {
	Common KeyValueStore

	LogicalClock int // Orologio logico scalare
	mutexClock   sync.Mutex

	Queue      []msg.MessageS
	mutexQueue sync.Mutex // Mutex per proteggere l'accesso concorrente alla coda

	executeFunctionMutex sync.Mutex // Mutex aggiunto per evitare scheduling che interrompano l'invio a livello applicativo del messaggio
}

type KeyValueStore struct {
	Datastore map[string]string // Mappa -> struttura dati che associa chiavi a valori
	Id        int               // Id che identifica il server stesso

	//clientMap map[string]ClientMap // Mappa -> struttura dati che associa chiavi a valori
}

// ----- Consistenza Causale ----- //

func NewKeyValueStoreCausal(idServer int) *KeyValueStoreCausale {
	kvc := &KeyValueStoreCausale{
		Common: KeyValueStore{
			Datastore: make(map[string]string),
		},
	}

	for i := 0; i < common.Replicas; i++ {
		kvc.SetVectorClock(0, 0) // Inizializzazione dell'orologio logico vettoriale
	}

	kvc.SetQueue(make([]msg.MessageC, 0)) // Inizializzazione della coda
	kvc.SetIdServer(idServer)

	return kvc
}

func (kvc *KeyValueStoreCausale) SetDatastore(key string, value string) {
	kvc.Common.Datastore[key] = value
}

func (kvc *KeyValueStoreCausale) SetVectorClock(index int, value int) {
	kvc.VectorClock[index] = value
}

func (kvc *KeyValueStoreCausale) SetQueue(queue []msg.MessageC) {
	kvc.Queue = queue
}

func (kvc *KeyValueStoreCausale) SetIdServer(id int) {
	kvc.Common.Id = id
}

func (kvc *KeyValueStoreCausale) GetClock() [common.Replicas]int {
	return kvc.VectorClock
}

func (kvc *KeyValueStoreCausale) GetDatastore() map[string]string {
	return kvc.Common.Datastore
}

func (kvc *KeyValueStoreCausale) GetQueue() []msg.MessageC {
	return kvc.Queue
}

func (kvc *KeyValueStoreCausale) GetIdServer() int {
	return kvc.Common.Id
}

// ----- Consistenza Sequenziale ----- //

func NewKeyValueStoreSequential(idServer int) *KeyValueStoreSequential {
	kvSequential := &KeyValueStoreSequential{
		Common: KeyValueStore{
			Datastore: make(map[string]string),
		},
	}
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
		if kvs.receiveAssumeFIFO == args.Timestamp && args.Key != Common.EndKey {
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
		if kvs.sendAssumeFIFO == msg.Args.Timestamp && msg.Args.Key != Common.EndKey {
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
