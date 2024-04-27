package main

import (
	"main/common"
	"sync"
)

type ClockServer interface {
	GetClock() interface{}
	GetDatastore() map[string]string
	GetQueue() interface{}
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
	Datastore map[string]string // Mappa -> struttura dati che associa chiavi a valori
	Id        int               // Id che identifica il server stesso

	VectorClock [common.Replicas]int // Orologio vettoriale
	mutexClock  sync.Mutex

	Queue      []MessageC
	mutexQueue sync.Mutex // Mutex per proteggere l'accesso concorrente alla coda

	executeFunctionMutex sync.Mutex // Mutex aggiunto per evitare scheduling che interrompano l'invio a livello applicativo del messaggio

	clientMap map[string]ClientMap // Mappa -> struttura dati che associa chiavi a valori
}

// KeyValueStoreSequential rappresenta il servizio di memorizzazione chiave-valore specializzato nel sequenziale
type KeyValueStoreSequential struct {
	Datastore map[string]string // Mappa -> struttura dati che associa chiavi a valori
	Id        int               // Id che identifica il server stesso

	LogicalClock int // Orologio logico scalare
	mutexClock   sync.Mutex

	Queue      []MessageS
	mutexQueue sync.Mutex // Mutex per proteggere l'accesso concorrente alla coda

	executeFunctionMutex sync.Mutex // Mutex aggiunto per evitare scheduling che interrompano l'invio a livello applicativo del messaggio

}

// ----- Consistenza Causale ----- //

func (kvc *KeyValueStoreCausale) GetClock() interface{} {
	return kvc.VectorClock
}

func (kvc *KeyValueStoreCausale) GetDatastore() map[string]string {
	return kvc.Datastore
}

func (kvc *KeyValueStoreCausale) GetQueue() interface{} {
	return kvc.Queue
}

func (kvc *KeyValueStoreCausale) GetIdServer() int {
	return kvc.Id
}

// ----- Consistenza Sequenziale ----- //

func (kvs *KeyValueStoreSequential) GetClock() interface{} {
	return kvs.LogicalClock
}

func (kvs *KeyValueStoreSequential) GetDatastore() map[string]string {
	return kvs.Datastore
}

func (kvs *KeyValueStoreSequential) GetQueue() interface{} {
	return kvs.Queue
}

func (kvs *KeyValueStoreSequential) GetIdServer() int {
	return kvs.Id
}
