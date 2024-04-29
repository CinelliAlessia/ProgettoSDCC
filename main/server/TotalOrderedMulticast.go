package main

import (
	"fmt"
	"main/common"
	"net/rpc"
	"sort"
)

// TotalOrderedMulticast gestione dell'evento esterno ricevuto da un server
// Implementazione del multicast totalmente ordinato
// Il server ha inviato in multicast il messaggio di update per msg
func (kvs *KeyValueStoreSequential) TotalOrderedMulticast(message MessageS, response *common.Response) error {
	// Aggiunta della richiesta in coda
	kvs.addToSortQueue(message)

	// Solo per debug
	kvs.mutexClock.Lock()
	if kvs.GetIdServer() != message.GetIdSender() {
		printDebugBlue("RICEVUTO da server", message, kvs)
	}
	kvs.mutexClock.Unlock()

	// Invio ack a tutti i server per notificare la ricezione della richiesta
	sendAck(message)

	// Inizializzo la risposta a false, corrisponde alla risposta che leggerà il client
	response.Result = false

	// Ciclo finché canExecute non restituisce true, in quel caso
	// la richiesta può essere eseguita a livello applicativo
	for {
		stop, err := kvs.canExecute(message, response)
		if stop {
			break
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func (kvs *KeyValueStoreSequential) canExecute(message MessageS, response *common.Response) (bool, error) {
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
	return false, nil
}

// ----- FUNZIONI PER GESTIRE GLI ACK ----- //

// ReceiveAck gestisce gli ack dei messaggi ricevuti.
// Se il messaggio è presente nella coda incrementa il numero di ack e restituisce true, altrimenti lo inserisce in coda e incrementa il numero di ack ricevuti.
func (kvs *KeyValueStoreSequential) ReceiveAck(message MessageS, reply *bool) error {
	*reply = kvs.updateAckMessage(message)
	if !(*reply) {
		//fmt.Println("Ricevuto ack di un messaggio non presente", message.TypeOfMessage, message.Args.Key, ":", message.Args.Value, "ts", message.Args.Timestamp)
		kvs.addToSortQueue(message)
		*reply = kvs.updateAckMessage(message)
	}
	//fmt.Println("Ricevuto ack di:", message.TypeOfMessage, message.Args.Key+":"+message.Args.Value)
	return nil
}

// sendAck invia con una goroutine a ciascun server un ack del messaggio ricevuto
func sendAck(message MessageS) {
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
func sendAckRPC(conn *rpc.Client, message MessageS, reply *bool) error {
	common.RandomDelay()

	err := conn.Call("KeyValueStoreSequential.ReceiveAck", message, reply)
	if err != nil {
		return err
	}

	if !*reply {
		return fmt.Errorf("sendAckRPC: Errore ReceiveAck ha risposto false, non dovrebbe accadere %v\n", message)
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
func (kvs *KeyValueStoreSequential) addToSortQueue(message MessageS) {
	kvs.mutexQueue.Lock()
	defer kvs.mutexQueue.Unlock()

	// Verifica se il messaggio è già presente nella coda se è già presente non fare nulla
	// Controllo effettuato perché è possibile che una richiesta venga aggiunta in coda sia alla ricezione della richiesta stessa
	// sia di un ack che ne faccia riferimento, e quella richiesta non era ancora in coda
	isPresent := false
	for _, msg := range kvs.GetQueue() {
		if msg.GetIdMessage() == message.GetIdMessage() {
			isPresent = true
			return
		}
	}

	// Se il messaggio non è presente, aggiungilo alla coda
	if !isPresent {
		kvs.SetQueue(append(kvs.GetQueue(), message))
		//fmt.Println(message.TypeOfMessage, message.Args.Key+":"+message.Args.Value, "aggiunto alla coda")

		// Ordina la coda in base al logicalClock, a parità di timestamp l'ordinamento è deterministico in base all'ID
		sort.Slice(kvs.GetQueue(), func(i, j int) bool {
			if kvs.GetQueue()[i].GetKey() == common.EndKey {
				return false // i messaggi con key: endKey vanno sempre alla fine
			}
			if kvs.GetQueue()[j].GetKey() == common.EndKey {
				return true // i messaggi con key: endKey vanno sempre alla fine
			}
			if kvs.GetQueue()[i].GetClock() == kvs.GetQueue()[j].GetClock() {
				return kvs.GetQueue()[i].GetIdMessage() < kvs.GetQueue()[j].GetIdMessage()
			}
			return kvs.GetQueue()[i].GetClock() < kvs.GetQueue()[j].GetClock()
		})
	}
}

// removeMessageToQueue Rimuove il messaggio passato come argomento dalla coda solamente se è l'elemento in testa, l'eliminazione si basa sull'ID
func (kvs *KeyValueStoreSequential) removeMessageToQueue(message MessageS) {
	kvs.mutexQueue.Lock()
	defer kvs.mutexQueue.Unlock()

	if len(kvs.GetQueue()) > 0 && kvs.GetQueue()[0].GetClock() == message.GetClock() && kvs.GetQueue()[0].GetIdMessage() == message.GetIdMessage() {
		// Rimuovi il primo elemento dalla slice
		//kvs.Queue = kvs.Queue[1:]
		kvs.SetQueue(kvs.GetQueue()[1:])
		return
	}

	//fmt.Println("removeMessageToQueue: Messaggio con ID", message.Id, "non trovato nella coda")
	//fmt.Println(message.TypeOfMessage, message.Args.Key+":"+message.Args.Value)
}

// updateAckMessage aggiorna, incrementando il numero di ack ricevuti, il messaggio in coda corrispondente all'id del messaggio passato come argomento
func (kvs *KeyValueStoreSequential) updateAckMessage(message MessageS) bool {
	kvs.mutexQueue.Lock()
	defer kvs.mutexQueue.Unlock()

	for i := range kvs.GetQueue() {
		if kvs.GetQueue()[i].GetClock() == message.GetClock() && kvs.GetQueue()[i].GetIdMessage() == message.GetIdMessage() {
			//fmt.Println("Ricevuto ack di:", message.TypeOfMessage, message.Args.Key+":"+message.Args.Value)
			// Aggiorna il messaggio nella coda incrementando il numero di ack ricevuti
			SetNumberAck(&(kvs.GetQueue()[i]), kvs.GetQueue()[i].GetNumberAck()+1)
			//kvs.Queue[i].NumberAck++
			return true
		}
	}
	return false
}

// ----- FUNZIONI PER GESTIRE L'ESECUZIONE DEL MESSAGGIO A LIVELLO APPLICATIVO ----- //

// controlSendToApplication verifica se è possibile inviare la richiesta a livello applicativo, Pj consegna msg_i all’applicazione se:
// 1. msg_i è in testa a queue_j
// 2. tutti gli ack relativi a msg_i sono stati ricevuti da Pj e,
// per ogni processo pk, c’è un messaggio msg_k in queue_j con timestamp maggiore di quello di msg_i (quest’ultima condizione sta a indicare
// che nessun altro processo può inviare in multicast un messaggio con timestamp potenzialmente minore o uguale a quello di msg_i)
func (kvs *KeyValueStoreSequential) controlSendToApplication(message MessageS) bool {
	// Per la corretta gestione di Common.EndKey
	if len(kvs.GetQueue()) == 0 {
		return true
	}

	switch {
	case kvs.isExecutableMessage(message): // Entrambe le condizioni per il Multicast Totalmente Ordinato sono soddisfatte

		// Ho ricevuto tutti gli ack, posso eliminare il messaggio dalla coda
		go kvs.removeMessageToQueue(message)

		// Aggiornamento del clock -> Prendo il max timestamp tra il mio e quello allegato al messaggio ricevuto
		kvs.updateLogicalClock(message)
		return true

	case kvs.isEndKeyMessage(message): // Controllo se il messaggio è un endKey
		return true
	}

	// Non è stato eseguito il messaggio a livello applicativo
	return false
}

// isExecutableMessage ritorna true se:
// 1. msg_i è in testa a queue_j
// 2. tutti gli ack relativi a msg_i sono stati ricevuti da Pj
// 3. per ogni processo pk, c’è un messaggio msg_k in queue_j con timestamp maggiore di quello di msg_i
func (kvs *KeyValueStoreSequential) isExecutableMessage(message MessageS) bool {
	return len(kvs.GetQueue()) > 0 && kvs.GetQueue()[0].GetIdMessage() == message.GetIdMessage() && kvs.GetQueue()[0].GetNumberAck() == common.Replicas && kvs.secondCondition(message)
}

// secondCondition ritorna true se:
// Per ogni processo pk, c’è un messaggio msg_k in queue_j con timestamp maggiore di quello di msg_i
// Dato il messaggio in argomento: Ci devono essere almeno un messaggio per ciascun server con timestamp maggiore (msg.IdSender)
func (kvs *KeyValueStoreSequential) secondCondition(message MessageS) bool {
	//time.Sleep(2000 * time.Millisecond) // Aggiungo un ritardo per evitare che le stampe si sovrappongano
	// Controllo che ci sia almeno un messaggio per ciascun server con timestamp maggiore

	for i := 0; i < common.Replicas; i++ {
		found := false
		for _, msg := range kvs.GetQueue() {
			if msg.GetIdSender() == i && msg.GetClock() > message.GetClock() {
				found = true
				break
			}
		}
		if !found {
			kvs.isAllEndKey()
			//fmt.Println(message.TypeOfMessage, message.Args.Key+":"+message.Args.Value, "secondCondition non rispettata", "tsClient", message.Args.Timestamp, "assume", kvs.receiveAssumeFIFO)
			return false
		}
	}
	return true
}

// updateLogicalClock aggiorna il LogicalClock del server, calcolando il max tra il LogicalClock del server e quello del messaggio ricevuto
// e incrementando il LogicalClock di uno se il messaggio ricevuto non è stato inviato in multicast dal server stesso
func (kvs *KeyValueStoreSequential) updateLogicalClock(message MessageS) {
	kvs.mutexClock.Lock()
	defer kvs.mutexClock.Unlock()

	kvs.SetLogicalClock(common.Max(message.GetClock(), kvs.GetClock()))
	if kvs.GetIdServer() != message.GetIdSender() {
		kvs.SetLogicalClock(kvs.GetClock() + 1) // Devo incrementare il clock per gestire l'evento di receive
	}
}

// ----- FUNZIONI AUSILIARIE UTILIZZATE PER GESTIRE ENDKEY NELLA CODA ----- //

// isEndKeyMessage ritorna true se:
// 1. il messaggio in testa alla coda è uguale a message
// 2. il messaggio in testa alla coda ha ricevuto tutti gli ack
// 3. il messaggio in testa alla coda ha come chiave common.EndKey
func (kvs *KeyValueStoreSequential) isEndKeyMessage(message MessageS) bool {
	return len(kvs.GetQueue()) > 0 && kvs.GetQueue()[0].GetIdMessage() == message.GetIdMessage() && kvs.GetQueue()[0].GetNumberAck() == common.Replicas && message.GetKey() == common.EndKey
}

// isAllEndKey controlla se tutti i messaggi rimanenti in coda sono endKey
// se lo sono, svuota la coda
func (kvs *KeyValueStoreSequential) isAllEndKey() {
	kvs.updateEndKeyTimestamp()
	//printQueue(kvs)

	// Se in coda sono rimasti solamente i Common.Replicas
	// messaggi di endKey, e il receiveAssumeFIFO è pari a message.args.timestamp -> svuota la coda
	if len(kvs.GetQueue()) == common.Replicas {
		allEndKey := true
		for _, msg := range kvs.GetQueue() {
			if msg.GetKey() != common.EndKey {
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
				//kvs.Queue = kvs.Queue[1:]
			}
			return
		}
	}
}

// updateEndKeyTimestamp aggiorna il timestamp dei messaggi di endKey
// Se il primo messaggio in coda ha un timestamp uguale o maggiore a quello dell'endKey, incrementa il timestamp dell'endKey con il suo +1
// cosi da averli sicuramente alla fine della coda -> Non dovrebbe succedere secondo l'assunzione di una coda fifo
func (kvs *KeyValueStoreSequential) updateEndKeyTimestamp() {
	// Blocco la coda per garantire la sicurezza dei thread
	//kvs.mutexQueue.Lock()
	//defer kvs.mutexQueue.Unlock()

	// Verifico se c'è un messaggio "endKey" nella coda
	for i, msg := range kvs.GetQueue() {
		if msg.GetKey() == common.EndKey {
			// Se il primo messaggio nella coda ha un timestamp uguale o maggiore a quello dell'endKey
			if kvs.GetQueue()[0].GetClock() >= msg.GetClock() {
				// Incremento il timestamp di endKey con il mio +1
				SetClock(&(kvs.GetQueue()[i]), kvs.GetQueue()[0].GetClock()+1)
				//kvs.GetQueue()[i].LogicalClock = kvs.GetQueue()[0].GetClock() + 1
				break
			}
		}
	}
}
