package sequential

import (
	"errors"
	"fmt"
	"main/common"
	"main/server/algorithms"
	"main/server/message"
	"sort"
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

	msgSent      int // Contatore per i messaggi inviati con msgSenderID == kvs.senderID
	mutexMsgSent sync.Mutex

	BufferedMessage []commonMsg.MessageS // Buffer per memorizzare i messaggi che non possono essere eseguiti a livello applicativo
	mutexBuffered   sync.Mutex

	BufferedArgsReceive  []common.Args // Buffer per memorizzare gli argomenti che non possono essere ricevuti
	mutexBufferedReceive sync.Mutex
}

// NewKeyValueStoreSequential crea un nuovo KeyValueStoreSequential, inizializzando l'orologio logico scalare e la coda
// prende come argomento l'id del server replica, da 0 a common.Replicas-1
func NewKeyValueStoreSequential(idServer int) *KeyValueStoreSequential {
	kvs := &KeyValueStoreSequential{
		Common: algorithms.KeyValueStore{
			Datastore:  make(map[string]string),
			ClientMaps: make(map[string]*algorithms.ClientMap),
		},
	}

	kvs.SetClock(0)                             // Inizializzazione dell'orologio logico scalare
	kvs.SetQueue(make([]commonMsg.MessageS, 0)) // Inizializzazione della coda
	kvs.SetIdServer(idServer)                   // Inizializzazione dell'id del server
	kvs.SetMsgSent(0)                           // Inizializzazione del contatore dei messaggi inviati

	return kvs
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

func (kvs *KeyValueStoreSequential) PrintQueue() {
	for _, msg := range kvs.GetQueue() {
		fmt.Println(msg.GetTypeOfMessage(), msg.GetKey()+":"+msg.GetValue(), "ack", msg.GetNumberAck(), "clock", msg.GetClock(), "orderClient", msg.GetSendingFIFO())
	}
	if len(kvs.GetQueue()) == 0 {
		fmt.Println("Coda vuota")
	}
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

// GetServerID restituisce l'id del server
func (kvs *KeyValueStoreSequential) GetServerID() int {
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

// ExistClient verifica se un client è presente nella mappa client
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

// SetReceiveTsFromClient imposta il numero di richieste ricevute da un client
func (kvs *KeyValueStoreSequential) SetReceiveTsFromClient(id string, ts int) {
	val, _ := kvs.GetClientMap(id)
	val.SetReceiveOrderingFIFO(ts)
}

// GetReceiveTsFromClient restituisce il contatore di messaggi ricevuti allo specifico client
func (kvs *KeyValueStoreSequential) GetReceiveTsFromClient(id string) int {
	if clientMap, ok := kvs.GetClientMap(id); ok {
		return clientMap.GetReceiveOrderingFIFO()
	}
	return -1
}

// SetResponseOrderingFIFO Incrementa il numero di risposte inviate al client
func (kvs *KeyValueStoreSequential) SetResponseOrderingFIFO(id string, ts int) {
	val, _ := kvs.GetClientMap(id)
	i := val.GetResponseOrderingFIFO()
	val.SetResponseOrderingFIFO(i + ts)
}

// GetResponseOrderingFIFO Restituisce il contatore di messaggi inviati da me al singolo client
// protetto da mutexMaps (all)
func (kvs *KeyValueStoreSequential) GetResponseOrderingFIFO(ClientID string) int {
	val, ok := kvs.GetClientMap(ClientID)
	if !ok {
		fmt.Println("Client non esistente", val, kvs.Common.ClientMaps)
		return -1
	}
	return val.GetResponseOrderingFIFO()
}

// ----- Messaggi Inviati ----- //

func (kvs *KeyValueStoreSequential) SetMsgSent(value int) {
	kvs.mutexMsgSent.Lock()
	defer kvs.mutexMsgSent.Unlock()

	kvs.msgSent = value
}

func (kvs *KeyValueStoreSequential) GetMsgSent() int {
	kvs.mutexMsgSent.Lock()
	defer kvs.mutexMsgSent.Unlock()

	return kvs.msgSent
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

func (kvs *KeyValueStoreSequential) LockMutexClock() {
	kvs.mutexClock.Lock()
}

func (kvs *KeyValueStoreSequential) UnlockMutexClock() {
	kvs.mutexClock.Unlock()
}

func (kvs *KeyValueStoreSequential) LockMutexQueue() {
	kvs.mutexQueue.Lock()
}

func (kvs *KeyValueStoreSequential) UnlockMutexQueue() {
	kvs.mutexQueue.Unlock()
}

// Request bufferizzate per fifo ordering

// AddBufferedArgs Aggiunge un messaggio al buffer
func (kvs *KeyValueStoreSequential) AddBufferedArgs(args common.Args) {
	// Ordinata in base a Sending fifo
	kvs.BufferedArgsReceive = append(kvs.BufferedArgsReceive, args)
	sort.Slice(kvs.BufferedArgsReceive, func(i, j int) bool {
		return kvs.BufferedArgsReceive[i].GetSendingFIFO() < kvs.BufferedArgsReceive[j].GetSendingFIFO()
	})
}

// GetBufferedArgs Restituisce il buffer
func (kvs *KeyValueStoreSequential) GetBufferedArgs() []common.Args {
	return kvs.BufferedArgsReceive
}

// RemoveBufferedArgs Rimuove un messaggio dal buffer
func (kvs *KeyValueStoreSequential) RemoveBufferedArgs(args common.Args) {
	for i, a := range kvs.BufferedArgsReceive {
		if a.GetClientID() == args.GetClientID() &&
			a.GetSendingFIFO() == args.GetSendingFIFO() {
			kvs.BufferedArgsReceive = append(kvs.BufferedArgsReceive[:i], kvs.BufferedArgsReceive[i+1:]...)
			return
		}
	}
}

// canHandleOtherRequest controlla se ci sono messaggi bufferizzati che possono essere eseguiti
func (kvs *KeyValueStoreSequential) canHandleOtherRequest() {
	kvs.mutexBufferedReceive.Lock()
	defer kvs.mutexBufferedReceive.Unlock()

	// Controlla tutta la coda dei messaggi bufferizzati, per i messaggi che rispettano le
	// condizioni di invio, allora imposti il canale a true
	for _, args := range kvs.GetBufferedArgs() {
		idClient := args.GetClientID()

		// Il client è sicuramente nella mappa
		requestTs := kvs.GetReceiveTsFromClient(idClient) // Ottengo il numero di messaggi ricevuti da questo client
		if args.GetSendingFIFO() == requestTs {           // Se il messaggio che ricevo dal client è quello che mi aspetto
			// Blocco il mutex per evitare che il client possa inviare un nuovo messaggio prima che io abbia
			// finito di creare il precedente
			kvs.LockMutexMessage(idClient)
			kvs.SetReceiveTsFromClient(idClient, args.GetSendingFIFO()+1) // Incremento il timestamp di ricezione del client
			args.SetCondition(true)                                       // Imposto la condizione a true, il messaggio può essere eseguito
			kvs.RemoveBufferedArgs(args)
		}
	}
}

// Messaggi bufferizzati

// AddBufferedMessage Aggiunge un messaggio al buffer
func (kvs *KeyValueStoreSequential) AddBufferedMessage(message commonMsg.MessageS) {
	kvs.BufferedMessage = append(kvs.BufferedMessage, message)
	sort.Slice(kvs.BufferedMessage, func(i, j int) bool {
		return kvs.BufferedMessage[i].GetClock() < kvs.BufferedMessage[j].GetClock()
	})
}

// GetBufferedMessage Restituisce il buffer
func (kvs *KeyValueStoreSequential) GetBufferedMessage() []commonMsg.MessageS {
	return kvs.BufferedMessage
}

// RemoveBufferedMessage Rimuove un messaggio dal buffer
func (kvs *KeyValueStoreSequential) RemoveBufferedMessage(message commonMsg.MessageS) {
	for i, m := range kvs.BufferedMessage {
		if m.GetIdMessage() == message.GetIdMessage() {
			kvs.BufferedMessage = append(kvs.BufferedMessage[:i], kvs.BufferedMessage[i+1:]...)
			return
		}
	}
}

// canHandleOtherResponse controlla se ci sono messaggi bufferizzati che possono essere eseguiti
func (kvs *KeyValueStoreSequential) canHandleOtherResponse() {
	kvs.mutexBuffered.Lock()
	defer kvs.mutexBuffered.Unlock()
	// Controlla tutta la coda dei messaggi bufferizzati, per i messaggi che rispettano le
	// condizioni di invio, allora imposti il canale a true
	for i, message := range kvs.GetBufferedMessage() {
		msg := &kvs.GetBufferedMessage()[i]

		canSend := kvs.controlSendToApplication(msg) // Controllo se le due condizioni del M.C.O sono soddisfatte
		if canSend {
			msg.SetCondition(true) // Imposto la condizione a true
			kvs.RemoveBufferedMessage(message)
			break
		}
	}
}
