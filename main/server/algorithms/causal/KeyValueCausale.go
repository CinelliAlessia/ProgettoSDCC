package causal

import (
	"fmt"
	"main/common"
	"main/server/algorithms"
	"main/server/message"
	"sync"
)

// KeyValueStoreCausale rappresenta la struttura di memorizzazione chiave-valore per garantire consistenza causale
type KeyValueStoreCausale struct {
	Common algorithms.KeyValueStore

	VectorClock [common.Replicas]int // Orologio vettoriale
	mutexClock  sync.Mutex

	Queue      []commonMsg.MessageC
	mutexQueue sync.Mutex // Mutex per proteggere l'accesso concorrente alla coda

	executeFunctionMutex sync.Mutex // Mutex aggiunto per evitare scheduling che interrompano l'invio a livello applicativo del messaggio

	BufferedMessage []commonMsg.MessageC // Buffer per memorizzare i messaggi che non possono essere eseguiti a livello applicativo
	mutexBuffered   sync.Mutex

	BufferedArgsReceive  []common.Args // Buffer per memorizzare gli argomenti che non possono essere ricevuti
	mutexBufferedReceive sync.Mutex
}

// ----- Consistenza causale ----- //

// NewKeyValueStoreCausal crea un nuovo KeyValueStoreCausale, inizializzando l'orologio vettoriale e la coda
// prende come argomento l'id del server replica, da 0 a common.Replicas-1
func NewKeyValueStoreCausal(idServer int) *KeyValueStoreCausale {
	kvc := &KeyValueStoreCausale{
		Common: algorithms.KeyValueStore{
			Datastore:  make(map[string]string),
			ClientMaps: make(map[string]*algorithms.ClientMap),
		},
	}

	for i := 0; i < common.Replicas; i++ {
		kvc.SetVectorClock(0, 0) // Inizializzazione dell'orologio logico vettoriale
	}

	kvc.SetQueue(make([]commonMsg.MessageC, 0)) // Inizializzazione della coda
	kvc.SetIdServer(idServer)

	return kvc
}

// ----- Operazioni Datastore ----- //

// PutInDatastore inserisce una coppia chiave-valore nel datastore del server
// prende come argomenti la chiave e il valore da inserire
func (kvc *KeyValueStoreCausale) PutInDatastore(key string, value string) {
	kvc.Common.Datastore[key] = value
}

// DeleteFromDatastore elimina una coppia chiave-valore dal datastore del server
func (kvc *KeyValueStoreCausale) DeleteFromDatastore(key string) {
	delete(kvc.Common.Datastore, key)
}

// GetDatastore restituisce il datastore del server
func (kvc *KeyValueStoreCausale) GetDatastore() map[string]string {
	return kvc.Common.Datastore
}

// ----- Orologio Logico Vettoriale ----- //

func (kvc *KeyValueStoreCausale) SetVectorClock(index int, value int) {
	kvc.VectorClock[index] = value
}

// GetClock restituisce l'orologio logico scalare del server
func (kvc *KeyValueStoreCausale) GetClock() [common.Replicas]int {
	return kvc.VectorClock
}

func (kvc *KeyValueStoreCausale) GetClockIDServer(id int) int {
	return kvc.VectorClock[id]
}

// ----- Coda ----- //

// SetQueue imposta la coda del server
func (kvc *KeyValueStoreCausale) SetQueue(queue []commonMsg.MessageC) {
	kvc.Queue = queue
}

// GetQueue restituisce la coda del server
func (kvc *KeyValueStoreCausale) GetQueue() []commonMsg.MessageC {
	return kvc.Queue
}

// ----- ID Server ----- //

func (kvc *KeyValueStoreCausale) SetIdServer(id int) {
	kvc.Common.Id = id
}

func (kvc *KeyValueStoreCausale) GetServerID() int {
	return kvc.Common.Id
}

// ----- Print ----- //
func (kvc *KeyValueStoreCausale) PrintDatastore() {
	fmt.Println(kvc.Common.Datastore)
}

func (kvc *KeyValueStoreCausale) PrintQueue() {
	for _, message := range kvc.GetQueue() {
		fmt.Println(message)
	}
}

// ----- Client Map ----- //

// NewClientMap crea una nuova mappa client per tenere conto dell'assunzione FIFO Ordering dei messaggi
func (kvc *KeyValueStoreCausale) NewClientMap(idClient string) {
	kvc.Common.MutexMaps.Lock()
	defer kvc.Common.MutexMaps.Unlock()

	kvc.Common.ClientMaps[idClient] = &algorithms.ClientMap{
		ReceiveOrderingFIFO:  0,
		ResponseOrderingFIFO: 0,
	}
}

// ExistClient verifica se un client è presente nella mappa client
func (kvc *KeyValueStoreCausale) ExistClient(idClient string) bool {
	_, ok := kvc.GetClientMap(idClient)
	return ok
}

// GetClientMap restituisce la mappa client associata a un client, identificato da un id univoco preso come argomento
func (kvc *KeyValueStoreCausale) GetClientMap(id string) (*algorithms.ClientMap, bool) {
	kvc.Common.MutexMaps.Lock()
	defer kvc.Common.MutexMaps.Unlock()

	val, ok := kvc.Common.ClientMaps[id]
	return val, ok
}

// SetReceiveTsFromClient imposta il numero di messaggi ricevuti da un singolo client
//   - id: id del client
//   - ts: timestamp del messaggio ricevuto
func (kvc *KeyValueStoreCausale) SetReceiveTsFromClient(id string, ts int) {
	val, _ := kvc.GetClientMap(id)
	val.SetReceiveOrderingFIFO(ts)
}

// GetReceiveTsFromClient restituisce il contatore di messaggi ricevuti allo specifico client
//   - id: id del client
func (kvc *KeyValueStoreCausale) GetReceiveTsFromClient(id string) int {
	if clientMap, ok := kvc.GetClientMap(id); ok {
		return clientMap.GetReceiveOrderingFIFO()
	}
	return -1
}

// SetResponseOrderingFIFO Incrementa il numero di risposte inviate al client
//   - id: id del client
//   - ts: timestamp del messaggio inviato
func (kvc *KeyValueStoreCausale) SetResponseOrderingFIFO(id string, ts int) {
	val, _ := kvc.GetClientMap(id)
	i := val.GetResponseOrderingFIFO()
	val.SetResponseOrderingFIFO(i + ts)
}

// GetResponseOrderingFIFO Restituisce il contatore di messaggi inviati da me al singolo client
//   - id: id del client
func (kvc *KeyValueStoreCausale) GetResponseOrderingFIFO(ClientID string) int {
	val, _ := kvc.GetClientMap(ClientID)
	return val.GetResponseOrderingFIFO()
}

// ----- Mutex ----- //

func (kvc *KeyValueStoreCausale) LockMutexMessage(ClientID string) {
	val, _ := kvc.GetClientMap(ClientID)
	val.LockMutexMessage()
}

func (kvc *KeyValueStoreCausale) UnlockMutexMessage(ClientID string) {
	val, _ := kvc.GetClientMap(ClientID)
	if val == nil {
		fmt.Println("Errore: clientMap nil")
	}
	val.UnlockMutexMessage()
}

func (kvc *KeyValueStoreCausale) LockMutexClock() {
	kvc.mutexClock.Lock()
}

func (kvc *KeyValueStoreCausale) UnlockMutexClock() {
	kvc.mutexClock.Unlock()
}

func (kvc *KeyValueStoreCausale) LockMutexQueue() {
	kvc.mutexQueue.Lock()
}

func (kvc *KeyValueStoreCausale) UnlockMutexQueue() {
	kvc.mutexQueue.Unlock()
}

// Buffer Args per FIFO Ordered

// AddBufferedArgs Aggiunge un messaggio al buffer
func (kvc *KeyValueStoreCausale) AddBufferedArgs(args common.Args) {
	kvc.BufferedArgsReceive = append(kvc.BufferedArgsReceive, args)
}

func (kvc *KeyValueStoreCausale) GetBufferedArgs() []common.Args {
	return kvc.BufferedArgsReceive
}

// RemoveBufferedArgs Rimuove un messaggio dal buffer
func (kvc *KeyValueStoreCausale) RemoveBufferedArgs(args common.Args) {
	for i, a := range kvc.BufferedArgsReceive {
		if a.GetClientID() == args.GetClientID() &&
			a.GetSendingFIFO() == args.GetSendingFIFO() {
			kvc.BufferedArgsReceive = append(kvc.BufferedArgsReceive[:i], kvc.BufferedArgsReceive[i+1:]...)
			return
		}
	}
}

func (kvc *KeyValueStoreCausale) canHandleOtherRequest() {
	kvc.mutexBufferedReceive.Lock()
	defer kvc.mutexBufferedReceive.Unlock()
	// Controlla tutta la coda dei messaggi bufferizzati, per i messaggi che rispettano le
	// condizioni di invio, allora imposti il canale a true
	for _, args := range kvc.GetBufferedArgs() {

		idClient := args.GetClientID()

		// Il client è sicuramente nella mappa
		requestTs := kvc.GetReceiveTsFromClient(idClient) // Ottengo il numero di messaggi ricevuti da questo client
		if args.GetSendingFIFO() == requestTs {           // Se il messaggio che ricevo dal client è quello che mi aspetto
			// Blocco il mutex per evitare che il client possa inviare un nuovo messaggio prima che io abbia
			// finito di creare il precedente
			kvc.LockMutexMessage(idClient)
			kvc.SetReceiveTsFromClient(idClient, args.GetSendingFIFO()+1) // Incremento il timestamp di ricezione del client
			args.SetCondition(true)                                       // Imposto la condizione a true, il messaggio può essere eseguito
			kvc.RemoveBufferedArgs(args)
		}
	}
}

// Messaggi bufferizzati

// AddBufferedMessage Aggiunge un messaggio al buffer
func (kvc *KeyValueStoreCausale) AddBufferedMessage(message commonMsg.MessageC) {
	kvc.BufferedMessage = append(kvc.BufferedMessage, message)
}

// GetBufferedMessage Restituisce il buffer
func (kvc *KeyValueStoreCausale) GetBufferedMessage() []commonMsg.MessageC {
	return kvc.BufferedMessage
}

// RemoveBufferedMessage Rimuove un messaggio dal buffer
func (kvc *KeyValueStoreCausale) RemoveBufferedMessage(message commonMsg.MessageC) {
	for i, m := range kvc.BufferedMessage {
		if m.GetIdMessage() == message.GetIdMessage() {
			kvc.BufferedMessage = append(kvc.BufferedMessage[:i], kvc.BufferedMessage[i+1:]...)
			return
		}
	}
}

func (kvc *KeyValueStoreCausale) canHandleOtherResponse() {
	// Controlla tutta la coda dei messaggi bufferizzati, per i messaggi che rispettano le
	// condizioni di invio, allora imposti il canale a true
	kvc.mutexBuffered.Lock()
	defer kvc.mutexBuffered.Unlock()

	for i, message := range kvc.GetBufferedMessage() {
		msg := &kvc.GetBufferedMessage()[i]

		canSend := kvc.controlSendToApplication(msg) // Controllo se le due condizioni del M.C.O sono soddisfatte
		if canSend {
			msg.SetCondition(true) // Imposto la condizione a true
			kvc.RemoveBufferedMessage(message)
			break
		}
	}
}
