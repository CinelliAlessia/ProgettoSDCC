package keyvaluestore

import "sync"

type ClockServer interface {
	GetDatastore() map[string]string
	GetIdServer() int
}

// ClientMap rappresenta la struttura dati per memorizzare i timestamp delle richieste dei client, cosi da realizzare
// le assunzioni di comunicazione FIFO
type ClientMap struct {
	MutexRequest sync.Mutex
	RequestTs    int // TimestampClient della richiesta ricevuta dal client

	MutexExecute sync.Mutex
	ExecuteTs    int // TimestampClient di esecuzione della richiesta
}

type KeyValueStore struct {
	Datastore map[string]string // Mappa -> struttura dati che associa chiavi a valori
	Id        int               // Id che identifica il server stesso
}
