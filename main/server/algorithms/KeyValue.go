package algorithms

import (
	"sync"
)

type KeyValueStore struct {
	Datastore map[string]string // Mappa -> struttura dati che associa chiavi a valori
	Id        int               // Id che identifica il server stesso

	ClientMaps map[string]*ClientMap // Mappa -> struttura dati che associa chiavi a valori
	MutexMaps  sync.Mutex            // Mutex per proteggere l'accesso concorrente alla mappa
}

// ClientMap rappresenta la struttura dati per memorizzare i timestamp delle richieste dei client, cosi da realizzare
// le assunzioni di comunicazione FIFO
type ClientMap struct {
	ReceiveOrderingFIFO int // Contatore del server, di messaggi (da processare) ricevuti da un singolo client
	MutexReceive        sync.Mutex

	ResponseOrderingFIFO int // Contatore di messaggi inviati a un singolo client
	mutexResponse        sync.Mutex

	MutexClientMessage sync.Mutex // Mutex che protegge la creazione del messaggio
}

// SetReceiveOrderingFIFO imposta il timestamp della richiesta del client
func (m *ClientMap) SetReceiveOrderingFIFO(ts int) {
	m.MutexReceive.Lock()
	defer m.MutexReceive.Unlock()
	m.ReceiveOrderingFIFO = ts
}

func (m *ClientMap) GetReceiveOrderingFIFO() int {
	m.MutexReceive.Lock()
	defer m.MutexReceive.Unlock()
	return m.ReceiveOrderingFIFO
}

// RESPONSE ORDERING FIFO

// SetResponseOrderingFIFO imposta il contatore di messaggi inviati a un singolo client
func (m *ClientMap) SetResponseOrderingFIFO(ordering int) {
	m.mutexResponse.Lock()
	defer m.mutexResponse.Unlock()
	m.ResponseOrderingFIFO = ordering
}

// GetResponseOrderingFIFO restituisce il contatore di messaggi inviati a un singolo client
func (m *ClientMap) GetResponseOrderingFIFO() int {
	return m.ResponseOrderingFIFO
}

// Usate per evitare problemi di scheduling tra il true di canReceive e la creazione del messaggio

func (m *ClientMap) LockMutexMessage() {
	m.MutexClientMessage.Lock()
}

func (m *ClientMap) UnlockMutexMessage() {
	m.MutexClientMessage.Unlock()
}
