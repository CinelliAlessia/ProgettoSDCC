package main

import (
	"fmt"
	"github.com/fatih/color"
	"main/common"
	"net/rpc"
	"sort"
)

// TotalOrderedMulticast gestione dell'evento esterno ricevuto da un server
// Implementazione del multicast totalmente ordinato -> Il server ha inviato in multicast il messaggio di update
func (kvs *KeyValueStoreSequential) TotalOrderedMulticast(message MessageS, response *common.Response) error {

	// Aggiunta della richiesta in coda
	kvs.addToSortQueue(message)

	kvs.mutexClock.Lock()
	if kvs.Id != message.IdSender {
		kvs.printDebugBlue("RICEVUTO da server", message)
	}
	kvs.mutexClock.Unlock()

	// Invio ack a tutti i server per notificare la ricezione della richiesta
	sendAck(message)

	response.Result = false // Inizializzo la risposta a false, corrisponde alla risposta che leggerà il client

	// Ciclo finché controlSendToApplication non restituisce true
	// Controllo quando la richiesta può essere eseguita a livello applicativo
	for {
		kvs.executeFunctionMutex.Lock()

		canSend := kvs.controlSendToApplication(message)
		if canSend {
			// Invio a livello applicativo
			err := kvs.realFunction(message, response)

			kvs.executeFunctionMutex.Unlock()

			if err != nil {
				return err
			}
			break

		}
		kvs.executeFunctionMutex.Unlock()
	}
	if response.Result {
		return nil
	}
	return fmt.Errorf("TotalOrderedMulticast non andato a buon fine")
}

// ReceiveAck gestisce gli ack dei messaggi ricevuti.
// Se il messaggio è presente nella coda incrementa il numero di ack e restituisce true, altrimenti lo inserisce in coda e incrementa il numero di ack ricevuti.
func (kvs *KeyValueStoreSequential) ReceiveAck(message MessageS, reply *bool) error {
	*reply = kvs.updateMessage(message)
	if !(*reply) {
		//fmt.Println("Ricevuto ack di un messaggio non presente", message.TypeOfMessage, message.Args.Key, ":", message.Args.Value)
		kvs.addToSortQueue(message)
		*reply = kvs.updateMessage(message)
	}
	fmt.Println("Ricevuto ack di:", message.TypeOfMessage, message.Args.Key+":"+message.Args.Value)
	return nil
}

// addToSortQueue aggiunge un messaggio alla coda locale, ordinandola in base al timeStamp, a parità di timestamp
// l'ordinamento è deterministico: per garantire l'ordinamento totale verranno usati gli ID associati al messaggio.
// La funzione è threadSafe per l'utilizzo della coda kvs.queue tramite kvs.mutexQueue
// Prima di aggiungere il messaggio alla coda, verifica che non sia già presente.
func (kvs *KeyValueStoreSequential) addToSortQueue(message MessageS) {
	kvs.mutexQueue.Lock()
	defer kvs.mutexQueue.Unlock()

	// Verifica se il messaggio è già presente nella coda
	isPresent := false
	for _, msg := range kvs.Queue {
		if msg.Id == message.Id {
			isPresent = true
			break
		}
	}

	// Se il messaggio non è presente, aggiungilo alla coda
	if !isPresent {
		kvs.Queue = append(kvs.Queue, message)

		// Ordina la coda in base al logicalClock, a parità di timestamp l'ordinamento è deterministico in base all'ID
		sort.Slice(kvs.Queue, func(i, j int) bool {
			if kvs.Queue[i].LogicalClock == kvs.Queue[j].LogicalClock {
				return kvs.Queue[i].Id < kvs.Queue[j].Id
			}
			return kvs.Queue[i].LogicalClock < kvs.Queue[j].LogicalClock
		})

		//printQueue(kvs)
	}
}

// removeMessageToQueue Rimuove il messaggio passato come argomento dalla coda solamente se è l'elemento in testa, l'eliminazione si basa sull'ID
func (kvs *KeyValueStoreSequential) removeMessageToQueue(message MessageS) {
	kvs.mutexQueue.Lock()
	defer kvs.mutexQueue.Unlock()
	if kvs.Queue[0].LogicalClock == message.LogicalClock && kvs.Queue[0].Id == message.Id {
		// Rimuovi l'elemento dalla slice
		kvs.Queue = kvs.Queue[1:]
		return
	}
	fmt.Println("removeMessageToQueue: Messaggio con ID", message.Id, "non trovato nella coda")
}

// updateMessage aggiorna, incrementando il numero di ack ricevuti, il messaggio in coda corrispondente all'id del messaggio passato come argomento
func (kvs *KeyValueStoreSequential) updateMessage(message MessageS) bool {
	kvs.mutexQueue.Lock()
	defer kvs.mutexQueue.Unlock()
	for i := range kvs.Queue {
		if kvs.Queue[i].LogicalClock == message.LogicalClock && kvs.Queue[i].Id == message.Id {
			// Aggiorna il messaggio nella coda incrementando il numero di ack ricevuti
			//fmt.Println("Ricevuto ack di:", message.TypeOfMessage, message.Args.Key+":"+message.Args.Value)
			kvs.Queue[i].NumberAck++
			return true
		}
	}
	return false
}

// controlSendToApplication verifica se è possibile inviare la richiesta a livello applicativo
// 4. pj consegna msg_i all’applicazione se msg_i è in testa a queue_j, tutti gli ack relativi a msg_i sono stati ricevuti
// da pj e, per ogni processo pk, c’è un messaggio msg_k in queue_j con timestamp maggiore di quello di msg_i
// (quest’ultima condizione sta a indicare che nessun altro processo può inviare in multicast un messaggio con
// timestamp potenzialmente minore o uguale a quello di msg_i)
func (kvs *KeyValueStoreSequential) controlSendToApplication(message MessageS) bool {
	if kvs.Queue[0].Id == message.Id && kvs.Queue[0].NumberAck == common.Replicas {

		// Aggiornamento del clock -> Prendo il max timestamp tra il mio e quello allegato al messaggio ricevuto
		kvs.mutexClock.Lock()
		kvs.LogicalClock = common.Max(message.LogicalClock, kvs.LogicalClock)
		if kvs.Id != message.IdSender {
			kvs.LogicalClock++ // Devo incrementare il clock per gestire l'evento di receive
		}
		kvs.mutexClock.Unlock()

		// Ho ricevuto tutti gli ack, posso eliminare il messaggio dalla coda
		go kvs.removeMessageToQueue(message)
		return true

	} else if kvs.Queue[0].NumberAck > common.Replicas {
		fmt.Println(color.RedString("controlSendToApplication: Il messaggio in testa alla coda ha ricevuto più ack del previsto"))
	}

	return false
}

// sendAck invia con una goroutine a ciascun server un ack del messaggio ricevuto
func sendAck(message MessageS) {
	canSend := 0

	//Invio in una goroutine controllando se il server a cui ho inviato l'ACK fosse a conoscenza del messaggio a cui mi stavo riferendo.
	for i := 0; i < common.Replicas; i++ {
		go func(replicaPort string, index int) {

			reply := false
			serverName := common.GetServerName(replicaPort, index)
			conn, err := rpc.Dial("tcp", serverName)
			if err != nil {
				fmt.Printf("sendAck: Errore durante la connessione al server %s: %v\n", replicaPort, err)
				return
			}
			fmt.Println("Inviato ack di:", message.TypeOfMessage, message.Args.Key+":"+message.Args.Value)
			err = sendAckRPC(conn, message, &reply)
			if err != nil {
				fmt.Printf("sendAck: Errore durante la chiamata RPC receiveAck %v\n", err)
				return
			}

			canSend++

			// Chiudo la connessione dopo essere sicuro che l'ack è stato inviato
			err = conn.Close()
			if err != nil {
				fmt.Println("Errore durante la chiusura della connessione in TotalOrderedMulticast.sendAckRPC:", err)
				return
			}

		}(common.ReplicaPorts[i], i)
	}

	for {
		// In questo modo controllo che tutti i miei ack siano arrivati a destinazione
		// sendAck non va chiamata in una goroutine quindi aspetteremo fin quando tutti avranno letto l'ack
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
		return fmt.Errorf("sendAckRPC: Errore ReceiveAck ha risposto false")
	}
	// Problema: Se un messaggio ha un ritardo eccessivo, ...
	// La soluzione al ritardo eccessivo è non inviare a livello applicativo il messaggio se il server che contatto non mi risponde true
	// mi aspetto una risposta ad un ack
	// Posso creare un lucchetto per messaggio o un canale, se è lockato non ho ricevuto almeno una risposta all'ack che ho inviato.
	// Questo dovrebbe rientrare nelle assunzioni di comunicazione affidabile
	return nil
}
