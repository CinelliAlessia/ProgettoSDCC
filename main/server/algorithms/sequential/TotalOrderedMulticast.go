package sequential

import (
	"fmt"
	"main/common"
	"main/server/message"
	"net/rpc"
	"sort"
)

// Update gestisce l'evento esterno ricevuto da un server
// Implementazione del multicast totalmente ordinato
// Il server ha inviato in multicast il messaggio di update per msg
func (kvs *KeyValueStoreSequential) Update(message commonMsg.MessageS, response *common.Response) error {

	// Solo per debug
	kvs.LockMutexClock()
	if kvs.GetServerID() != message.GetSenderID() {
		printDebugBlue("RICEVUTO da server", message, kvs)
	}
	kvs.UnlockMutexClock()

	//Aggiunta del messaggio alla coda ordinata per timestamp
	kvs.addToSortQueue(&message)

	// Invio ack a tutti i server per notificare la ricezione della richiesta
	err := sendAck(&message)
	if err != nil {
		return err
	}

	// Ciclo fin quando canExecute restituisce true, in quel caso
	// la richiesta può essere eseguita a livello applicativo
	for {
		stop, err := kvs.canExecute(&message, response)
		if stop {
			break
		}
		if err != nil {
			return err
		}
	}
	return nil
}

// ----- FUNZIONI PER GESTIRE GLI ACK ----- //

// ReceiveAck gestisce gli ack dei messaggi ricevuti.
//
//	Se il messaggio è presente nella coda incrementa il numero di ack e restituisce true, altrimenti lo inserisce in
//	coda e incrementa il numero di ack ricevuti.
func (kvs *KeyValueStoreSequential) ReceiveAck(message commonMsg.MessageS, reply *bool) error {
	*reply = kvs.updateAckMessage(&message)
	if !(*reply) {
		kvs.addToSortQueue(&message)
		*reply = kvs.updateAckMessage(&message)
	}
	return nil
}

// sendAck invia con una goroutine a ciascun server un ack del messaggio ricevuto e attende la risposta
func sendAck(message *commonMsg.MessageS) error {

	ackChannel := make(chan error, common.Replicas)

	for i := 0; i < common.Replicas; i++ {
		// Invio ack in una goroutine
		go func(replicaPort string, index int) {

			serverName := common.GetServerName(replicaPort, index)
			conn, err := rpc.Dial("tcp", serverName)
			if err != nil {
				fmt.Printf("sendAck: Errore durante la connessione al server %s: %v\n", replicaPort, err)
				ackChannel <- err
				return
			}

			err = sendAckRPC(conn, message)
			if err != nil {
				fmt.Printf("sendAck: Errore durante la chiamata RPC receiveAck %v\n", err)
				ackChannel <- err
				return
			}

			ackChannel <- err

			// Chiudo la connessione dopo essere sicuro che l'ack è stato inviato
			err = conn.Close()
			if err != nil {
				fmt.Println("sendAck: Errore durante la chiusura della connessione in sendAckRPC: ", err)
				return
			}

		}(common.ReplicaPorts[i], i)
	}

	// Controllo e attendo che tutti i miei ack siano arrivati a destinazione
	for i := 0; i < common.Replicas; i++ {
		err := <-ackChannel
		if err != nil {
			return fmt.Errorf("sendAck: Errore durante l'invio dell'ack al server %d: %v", i, err)
		}
	}

	return nil
}

// sendAckRPC invia l'ack tramite RPC, applicando un ritardo random
func sendAckRPC(conn *rpc.Client, message *commonMsg.MessageS) error {
	reply := false

	common.RandomDelay()
	err := conn.Call(common.Sequential+".ReceiveAck", &message, &reply)
	if err != nil {
		return err
	}

	if !reply {
		return fmt.Errorf("sendAckRPC: Errore ReceiveAck ha risposto false, non dovrebbe accadere %v\n", &message)
	}
	return nil
}

// ----- FUNZIONI PER GESTIRE LA CODA ----- //

// addToSortQueue aggiunge un messaggio alla coda locale, ordinandola in base al timeStamp.
// A parità di timestamp, l'ordinamento è deterministico:
// per garantire l'ordinamento totale verranno usati gli ID associati al messaggio.
// La funzione è threadSafe per l'utilizzo della coda kvs.queue tramite kvs.mutexQueue
// Prima di aggiungere il messaggio alla coda, verifica che non sia già presente.
func (kvs *KeyValueStoreSequential) addToSortQueue(message *commonMsg.MessageS) {
	kvs.LockMutexQueue()
	defer kvs.UnlockMutexQueue()

	// Verifica se il messaggio è già presente nella coda se è già presente non fare nulla
	// Controllo effettuato perché è possibile che una richiesta venga aggiunta in coda sia alla ricezione della
	// richiesta stessa sia di un ack che ne faccia riferimento, e quella richiesta non era ancora in coda
	isPresent := false
	for _, messageS := range kvs.GetQueue() {
		if messageS.GetIdMessage() == message.GetIdMessage() {
			isPresent = true
			return
		}
	}

	// Se il messaggio non è presente, aggiungilo alla coda
	if !isPresent {
		kvs.SetQueue(append(kvs.GetQueue(), *message))

		// Ordina la coda in base al logicalClock, a parità di timestamp l'ordinamento è deterministico in base all'ID
		sort.Slice(kvs.GetQueue(), func(i, j int) bool {
			if kvs.GetMsgFromQueue(i).GetClock() == kvs.GetMsgFromQueue(j).GetClock() {
				return kvs.GetMsgFromQueue(i).GetIdMessage() < kvs.GetMsgFromQueue(j).GetIdMessage()
			}
			return kvs.GetMsgFromQueue(i).GetClock() < kvs.GetMsgFromQueue(j).GetClock()
		})
	}
}

// removeMessageToQueue Rimuove il messaggio passato come argomento dalla coda solamente se è l'elemento in testa,
// l'eliminazione si basa sull'ID
func (kvs *KeyValueStoreSequential) removeMessageToQueue() {

	if len(kvs.GetQueue()) > 0 {
		// Rimuovi il primo elemento dalla slice
		kvs.SetQueue(kvs.GetQueue()[1:])
		return
	}
}

// updateAckMessage aggiorna, incrementando il numero di ack ricevuti, il messaggio in coda corrispondente all'id del messaggio passato come argomento
func (kvs *KeyValueStoreSequential) updateAckMessage(message *commonMsg.MessageS) bool {
	kvs.LockMutexQueue()
	defer kvs.UnlockMutexQueue()

	for i := range kvs.GetQueue() {
		if kvs.GetMsgFromQueue(i).GetIdMessage() == message.GetIdMessage() && // Se il messaggio in coda ha lo stesso id
			kvs.GetMsgFromQueue(i).GetClock() == message.GetClock() { // Se il messaggio in coda ha lo stesso timestamp

			// Aggiorna il messaggio nella coda incrementando il numero di ack ricevuti
			kvs.GetMsgFromQueue(i).SetNumberAck(kvs.GetMsgFromQueue(i).GetNumberAck() + 1)
			return true
		}
	}
	return false
}

// ----- FUNZIONI PER GESTIRE L'ESECUZIONE DEL MESSAGGIO A LIVELLO APPLICATIVO ----- //

// canExecute controlla se è possibile eseguire il messaggio a livello applicativo, se è possibile lo esegue
// e restituisce true, altrimenti restituisce false
func (kvs *KeyValueStoreSequential) canExecute(message *commonMsg.MessageS, response *common.Response) (bool, error) {
	kvs.executeFunctionMutex.Lock()
	defer kvs.executeFunctionMutex.Unlock()

	canSend := kvs.controlSendToApplication(message) // Controllo se le due condizioni del M.T.O sono soddisfatte

	if canSend {

		err := kvs.realFunction(message, response) // Invio a livello applicativo
		if err != nil {
			response.SetResult(false)
			return false, err
		}
		return true, nil
	}

	return false, nil
}

// controlSendToApplication verifica se è possibile inviare la richiesta a livello applicativo,
// Pj consegna msg_i all’applicazione se:
//   - msg_i è in testa a queue_j
//   - tutti gli ack relativi a msg_i sono stati ricevuti da Pj
//   - per ogni processo Pk, c’è un messaggio msg_k in queue_j con timestamp maggiore di quello di msg_i
func (kvs *KeyValueStoreSequential) controlSendToApplication(message *commonMsg.MessageS) bool {
	kvs.LockMutexQueue()
	defer kvs.UnlockMutexQueue()

	if kvs.GetServerID() == message.GetSenderID() { // Se il messaggio è stato inviato dal server stesso
		if kvs.GetResponseOrderingFIFO(message.GetClientID()) != message.GetSendingFIFO() {
			// Non ho ancora inviato tutti i messaggi precedenti chiesti dallo stesso client
			return false
		}
	}

	if (len(kvs.GetQueue()) > 0 && // Se ci sono elementi in coda
		kvs.GetMsgFromQueue(0).GetIdMessage() == message.GetIdMessage() && // Se il messaggio è in testa alla coda
		kvs.GetMsgFromQueue(0).GetNumberAck() == common.Replicas && // Se ha ricevuto tutti gli ack
		kvs.secondCondition(message)) || // Se per ogni processo pk, c’è un messaggio msg_k in queue con timestamp maggiore del messaggio passato come argomento

		kvs.isAllEndKey() { // Oppure, Se tutti i messaggi rimanenti in coda sono endKey, li elimino

		// Incremento il numero di risposte inviate al determinato client
		if message.GetSenderID() == kvs.GetServerID() {
			kvs.SetResponseOrderingFIFO(message.GetClientID(), 1)
		}

		// Tutte le condizioni sono soddisfatte
		kvs.removeMessageToQueue()

		// Aggiornamento del clock del server:
		// - Prendo il max timestamp tra il mio e quello allegato al messaggio ricevuto e lo si incrementa di uno
		kvs.updateLogicalClock(message)
		return true

	}
	return false
}

// secondCondition ritorna true se:
//   - Per ogni processo pk, c’è un messaggio msg_k in queue con timestamp maggiore del messaggio passato come argomento
func (kvs *KeyValueStoreSequential) secondCondition(message *commonMsg.MessageS) bool {

	// Controllo che ci sia almeno un messaggio per ciascun server con timestamp maggiore
	// rispetto a quello del messaggio in argomento

	for i := 0; i < common.Replicas; i++ {
		found := false
		for _, msg := range kvs.GetQueue() {
			if msg.GetClock() > message.GetClock() && // il messaggio in coda ha un timestamp maggiore
				msg.GetSenderID() == i { // Se il messaggio è stato inviato dal server i

				found = true
				break
			}
		}
		if !found {

			return false
		}
	}
	return true
}

// updateLogicalClock aggiorna il LogicalClock del server, calcolando il max tra il LogicalClock del server e quello del
// messaggio eseguibile a livello applicativo e incrementando il LogicalClock di uno
// se il messaggio ricevuto non è stato inviato in multicast dal server stesso
func (kvs *KeyValueStoreSequential) updateLogicalClock(message *commonMsg.MessageS) {
	kvs.LockMutexClock()
	defer kvs.UnlockMutexClock()

	kvs.SetClock(common.Max(message.GetClock(), kvs.GetClock()))

	if kvs.GetServerID() != message.GetSenderID() { // Se non ho già conteggiato il messaggio
		kvs.SetClock(kvs.GetClock() + 1) // Devo incrementare il clock per gestire l'evento di receive
	}
}

// ----- FUNZIONI AUSILIARIE UTILIZZATE PER GESTIRE ENDKEY NELLA CODA ----- //

// isAllEndKey controlla se tutti i messaggi rimanenti in coda sono endKey con numero di ack pari a common.Replicas
func (kvs *KeyValueStoreSequential) isAllEndKey() bool {
	allEndKey := true

	// Se in coda sono rimasti solamente messaggi di endKey con ack ricevuti pari a common.Replicas
	if len(kvs.GetQueue()) > 0 {
		for _, msg := range kvs.GetQueue() {
			if msg.GetKey() != common.EndKey ||
				msg.GetNumberAck() != common.Replicas {
				allEndKey = false
				break
			}
		}
	}
	return allEndKey
}
