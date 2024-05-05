package keyvaluestore

import (
	"fmt"
	"main/common"
	"main/server/message"
	"net/rpc"
	"sort"
)

// TotalOrderedMulticast gestione dell'evento esterno ricevuto da un server
// Implementazione del multicast totalmente ordinato
// Il server ha inviato in multicast il messaggio di update per msg
func (kvs *KeyValueStoreSequential) TotalOrderedMulticast(message commonMsg.MessageS, response *common.Response) error {
	// Solo per debug
	kvs.mutexClock.Lock()
	if kvs.GetIdServer() != message.GetIdSender() {
		printDebugBlue("RICEVUTO da server", message, nil, kvs)
	}
	kvs.mutexClock.Unlock()

	// Aggiunta della richiesta in coda
	//kvs.addToSortQueue(&message), non serve, lo faccio già in put e delete

	// Invio ack a tutti i server per notificare la ricezione della richiesta
	sendAck(&message)

	// Inizializzo la risposta a false, corrisponde alla risposta che leggerà il client
	response.SetResult(false)

	// Ciclo finché canExecute non restituisce true, in quel caso
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
// Se il messaggio è presente nella coda incrementa il numero di ack e restituisce true, altrimenti lo inserisce in coda e incrementa il numero di ack ricevuti.
func (kvs *KeyValueStoreSequential) ReceiveAck(message commonMsg.MessageS, reply *bool) error {
	*reply = kvs.updateAckMessage(&message)
	if !(*reply) {
		kvs.addToSortQueue(&message)
		*reply = kvs.updateAckMessage(&message)
	}
	return nil
}

// sendAck invia con una goroutine a ciascun server un ack del messaggio ricevuto
func sendAck(message *commonMsg.MessageS) {
	canSend := 0 // Contatore del numero di ack inviati che sono stati ricevuti (ho avuto una risposta alla chiamata RPC)
	//fmt.Println("sendAck", message.TypeOfMessage, message.Args.Key+":"+message.Args.Value)
	//Invio in una goroutine controllando se il server a cui ho inviato l'ACK fosse a conoscenza del messaggio a cui mi stavo riferendo.
	for i := 0; i < common.Replicas; i++ {
		// Invio ack in una goroutine
		go func(replicaPort string, index int) {

			reply := false

			serverName := common.GetServerName(replicaPort, index)
			conn, err := rpc.Dial("tcp", serverName)
			if err != nil {
				fmt.Printf("sendAck: Errore durante la connessione al server %s: %v\n", replicaPort, err)
				return
			}

			//fmt.Println("Inviato ack di:", message.TypeOfMessage, message.Args.Key+":"+message.Args.Value)
			err = sendAckRPC(conn, message, &reply)
			if err != nil {
				fmt.Printf("sendAck: Errore durante la chiamata RPC receiveAck %v\n", err)
				return
			}

			canSend++ // Incremento il numero di ack inviati che sono stati ricevuti

			// Chiudo la connessione dopo essere sicuro che l'ack è stato inviato
			err = conn.Close()
			if err != nil {
				fmt.Println("sendAck: Errore durante la chiusura della connessione in TotalOrderedMulticast.sendAckRPC: ", err)
				return
			}

		}(common.ReplicaPorts[i], i)
	}

	// Controllo e attendo che tutti i miei ack siano arrivati a destinazione
	for {
		if canSend == common.Replicas {
			break
		}
	}
}

// sendAckRPC invia l'ack tramite RPC, applicando un ritardo random
func sendAckRPC(conn *rpc.Client, message *commonMsg.MessageS, reply *bool) error {
	common.RandomDelay()

	err := conn.Call("KeyValueStoreSequential.ReceiveAck", &message, reply)
	if err != nil {
		return err
	}

	if !*reply {
		return fmt.Errorf("sendAckRPC: Errore ReceiveAck ha risposto false, non dovrebbe accadere %v\n", &message)
	}
	// Problema: Se un messaggio ha un ritardo eccessivo, ...
	// La soluzione al ritardo eccessivo è non inviare a livello applicativo il messaggio se il server che contatto non mi risponde true
	// mi aspetto una risposta ad un ack
	// Posso creare un lucchetto per messaggio o un canale, se è lockato non ho ricevuto almeno una risposta all'ack che ho inviato.
	// Questo dovrebbe rientrare nelle assunzioni di comunicazione affidabile
	return nil
}

// ----- FUNZIONI PER GESTIRE LA CODA ----- //

// addToSortQueue aggiunge un messaggio alla coda locale, ordinandola in base al timeStamp, a parità di timestamp
// l'ordinamento è deterministico: per garantire l'ordinamento totale verranno usati gli ID associati al messaggio.
// La funzione è threadSafe per l'utilizzo della coda kvs.queue tramite kvs.mutexQueue
// Prima di aggiungere il messaggio alla coda, verifica che non sia già presente.
func (kvs *KeyValueStoreSequential) addToSortQueue(message *commonMsg.MessageS) {
	kvs.mutexQueue.Lock()
	defer kvs.mutexQueue.Unlock()

	// Verifica se il messaggio è già presente nella coda se è già presente non fare nulla
	// Controllo effettuato perché è possibile che una richiesta venga aggiunta in coda sia alla ricezione della richiesta stessa
	// sia di un ack che ne faccia riferimento, e quella richiesta non era ancora in coda
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
		//fmt.Println(message.TypeOfMessage, message.Args.Key+":"+message.Args.Value, "aggiunto alla coda")

		// Ordina la coda in base al logicalClock, a parità di timestamp l'ordinamento è deterministico in base all'ID
		sort.Slice(kvs.GetQueue(), func(i, j int) bool {
			// Forzo l'aggiunta di endKey in coda
			if kvs.GetMsgToQueue(i).GetKey() == common.EndKey {
				return false // i messaggi con key: endKey vanno sempre alla fine
			}
			if kvs.GetMsgToQueue(j).GetKey() == common.EndKey {
				return true // i messaggi con key: endKey vanno sempre alla fine
			}
			if kvs.GetMsgToQueue(i).GetClock() == kvs.GetMsgToQueue(j).GetClock() {
				return kvs.GetMsgToQueue(i).GetIdMessage() < kvs.GetMsgToQueue(j).GetIdMessage()
			}
			return kvs.GetMsgToQueue(i).GetClock() < kvs.GetMsgToQueue(j).GetClock()
		})
	}
}

// removeMessageToQueue Rimuove il messaggio passato come argomento dalla coda solamente se è l'elemento in testa, l'eliminazione si basa sull'ID
func (kvs *KeyValueStoreSequential) removeMessageToQueue(message *commonMsg.MessageS) {
	kvs.mutexQueue.Lock()
	defer kvs.mutexQueue.Unlock()

	if len(kvs.GetQueue()) > 0 && kvs.GetMsgToQueue(0).GetClock() == message.GetClock() && kvs.GetMsgToQueue(0).GetIdMessage() == message.GetIdMessage() {
		// Rimuovi il primo elemento dalla slice
		kvs.SetQueue(kvs.GetQueue()[1:])
		//fmt.Println("Rimosso messaggio dalla coda:", message.GetTypeOfMessage(), message.GetKey()+":"+message.GetValue(), "tsClient", message.GetSendingFIFO(), "lenQ", len(kvs.GetQueue()))
		return
	}
}

// updateAckMessage aggiorna, incrementando il numero di ack ricevuti, il messaggio in coda corrispondente all'id del messaggio passato come argomento
func (kvs *KeyValueStoreSequential) updateAckMessage(message *commonMsg.MessageS) bool {
	kvs.mutexQueue.Lock()
	defer kvs.mutexQueue.Unlock()

	for i := range kvs.GetQueue() {
		if kvs.GetMsgToQueue(i).GetClock() == message.GetClock() && kvs.GetMsgToQueue(i).GetIdMessage() == message.GetIdMessage() {
			//fmt.Println("Ricevuto ack di:", message.TypeOfMessage, message.Args.Key+":"+message.Args.Value)
			// Aggiorna il messaggio nella coda incrementando il numero di ack ricevuti
			kvs.GetMsgToQueue(i).SetNumberAck(kvs.GetMsgToQueue(i).GetNumberAck() + 1)
			//kvs.Queue[i].NumberAck++
			return true
		}
	}
	return false
}

// ----- FUNZIONI PER GESTIRE L'ESECUZIONE DEL MESSAGGIO A LIVELLO APPLICATIVO ----- //

func (kvs *KeyValueStoreSequential) canExecute(message *commonMsg.MessageS, response *common.Response) (bool, error) {
	kvs.executeFunctionMutex.Lock()
	defer kvs.executeFunctionMutex.Unlock()

	canSend := kvs.controlSendToApplication(message)
	if canSend {
		// Invio a livello applicativo
		err := kvs.realFunction(message, response)
		if err != nil {
			return false, err
		}
		return true, nil
	}

	//time.Sleep(1 * time.Second)
	return false, nil
}

// controlSendToApplication verifica se è possibile inviare la richiesta a livello applicativo, Pj consegna msg_i all’applicazione se:
// 1. msg_i è in testa a queue_j
// 2. tutti gli ack relativi a msg_i sono stati ricevuti da Pj e,
// per ogni processo pk, c’è un messaggio msg_k in queue_j con timestamp maggiore di quello di msg_i (quest’ultima condizione sta a indicare
// che nessun altro processo può inviare in multicast un messaggio con timestamp potenzialmente minore o uguale a quello di msg_i)
func (kvs *KeyValueStoreSequential) controlSendToApplication(message *commonMsg.MessageS) bool {
	// Per la corretta gestione di Common.EndKey
	if len(kvs.GetQueue()) == 0 {
		return true
	}

	if kvs.conditionMessage(message) { // Entrambe le condizioni per il Multicast Totalmente Ordinato sono soddisfatte

		// Ho ricevuto tutti gli ack, posso eliminare il messaggio dalla coda
		go kvs.removeMessageToQueue(message)

		// Aggiornamento del clock -> Prendo il max timestamp tra il mio e quello allegato al messaggio ricevuto
		kvs.updateLogicalClock(message)
		return true
	}

	// Non è stato eseguito il messaggio a livello applicativo
	kvs.isAllEndKey()

	//fmt.Println(kvs.GetQueue())
	//fmt.Println(message.GetTypeOfMessage(), message.GetKey()+":"+message.GetValue(), "controlSendToApplication non rispettata", "tsClient", message.GetSendingFIFO())
	return false
}

// conditionMessage ritorna true se:
// 1. msg_i è in testa a queue_j
// 2. tutti gli ack relativi a msg_i sono stati ricevuti da Pj
// 3. per ogni processo pk, c’è un messaggio msg_k in queue_j con timestamp maggiore di quello di msg_i
func (kvs *KeyValueStoreSequential) conditionMessage(message *commonMsg.MessageS) bool {
	return len(kvs.GetQueue()) > 0 && // Se ci sono elementi in coda
		kvs.GetMsgToQueue(0).GetIdMessage() == message.GetIdMessage() && // Se il messaggio in testa è il messaggio in argomento
		kvs.GetMsgToQueue(0).GetNumberAck() == common.Replicas && // Se ha ricevuto tutti gli ack
		kvs.secondCondition(message) // Se la seconda condizione è rispettata
}

// secondCondition ritorna true se:
//   - Per ogni processo pk, c’è un messaggio msg_k in queue_j con timestamp maggiore di message passato come argomento
func (kvs *KeyValueStoreSequential) secondCondition(message *commonMsg.MessageS) bool {
	// Controllo che ci sia almeno un messaggio per ciascun server con timestamp maggiore
	// rispetto a quello del messaggio in argomento
	for i := 0; i < common.Replicas; i++ {
		found := false
		for _, msg := range kvs.GetQueue() {
			if msg.GetClock() > message.GetClock() && // il messaggio in coda ha un timestamp maggiore
				msg.GetIdSender() == i { // Se il messaggio è stato inviato dal server i
				found = true
				break
			}
		}
		if !found {
			kvs.isAllEndKey()
			//fmt.Println(message.GetTypeOfMessage(), message.GetKey()+":"+message.GetValue(),
			//	"secondCondition non rispettata", "tsClient", message.GetSendingFIFO())
			return false
		}
	}
	return true
}

// updateLogicalClock aggiorna il LogicalClock del server, calcolando il max tra il LogicalClock del server e quello del messaggio ricevuto
// e incrementando il LogicalClock di uno se il messaggio ricevuto non è stato inviato in multicast dal server stesso
func (kvs *KeyValueStoreSequential) updateLogicalClock(message *commonMsg.MessageS) {
	kvs.mutexClock.Lock()
	defer kvs.mutexClock.Unlock()

	kvs.SetClock(common.Max(message.GetClock(), kvs.GetClock()))
	if kvs.GetIdServer() != message.GetIdSender() {
		kvs.SetClock(kvs.GetClock() + 1) // Devo incrementare il clock per gestire l'evento di receive
	}
}

// ----- FUNZIONI AUSILIARIE UTILIZZATE PER GESTIRE ENDKEY NELLA CODA ----- //

// isEndKeyMessage ritorna true se:
// 1. il messaggio in testa alla coda è uguale a message
// 2. il messaggio in testa alla coda ha ricevuto tutti gli ack
// 3. il messaggio in testa alla coda ha come chiave common.EndKey
func (kvs *KeyValueStoreSequential) isEndKeyMessage(message *commonMsg.MessageS) bool {
	return len(kvs.GetQueue()) > 0 &&
		message.GetKey() == common.EndKey &&
		kvs.GetMsgToQueue(0).GetIdMessage() == message.GetIdMessage() &&
		kvs.GetMsgToQueue(0).GetNumberAck() == common.Replicas
}

// isAllEndKey controlla se tutti i messaggi rimanenti in coda sono endKey
// se lo sono, svuota la coda
func (kvs *KeyValueStoreSequential) isAllEndKey() {
	//kvs.updateEndKeyTimestamp()

	// Se in coda sono rimasti solamente i Common.Replicas
	// messaggi di endKey, e il receiveAssumeFIFO è pari a message.args.timestamp -> svuota la coda
	if len(kvs.GetQueue()) == common.Replicas {
		allEndKey := true
		for _, messageS := range kvs.GetQueue() {
			if messageS.GetKey() != common.EndKey {
				allEndKey = false
				break
			}
		}

		// Ho eseguito tutte le operazioni del client e in coda ho solamente endKey
		// posso svuotare la coda
		if allEndKey {
			for range kvs.GetQueue() {
				// Rimuovi il primo elemento dalla coda
				kvs.SetQueue(kvs.GetQueue()[1:])
			}
			return
		}
	}
}
