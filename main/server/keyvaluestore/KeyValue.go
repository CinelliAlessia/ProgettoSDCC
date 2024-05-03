package keyvaluestore

import (
	"sync"
)

type KeyValueStore struct {
	Datastore map[string]string // Mappa -> struttura dati che associa chiavi a valori
	Id        int               // Id che identifica il server stesso
}

// ClientMap rappresenta la struttura dati per memorizzare i timestamp delle richieste dei client, cosi da realizzare
// le assunzioni di comunicazione FIFO
type ClientMap struct {
	MutexRequest sync.Mutex
	RequestTs    int // TimestampClient della richiesta ricevuta dal client

	MutexExecute sync.Mutex
	ExecuteTs    int // TimestampClient di esecuzione della richiesta
}

func (m *ClientMap) SetRequestTs(ts int) {
	m.MutexRequest.Lock()
	defer m.MutexRequest.Unlock()
	m.RequestTs = ts
}

func (m *ClientMap) GetRequestTs() int {
	return m.RequestTs
}

func (m *ClientMap) SetExecuteTs(ts int) {
	m.MutexExecute.Lock()
	defer m.MutexExecute.Unlock()
	m.ExecuteTs = ts
}

func (m *ClientMap) GetExecuteTs() int {
	return m.ExecuteTs
}
