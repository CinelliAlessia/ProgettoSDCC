package main

import (
	"fmt"
	"main/common"
	"net/rpc"
	"sort"
	"time"
)

// TotalOrderedMulticast gestione dell'evento esterno ricevuto da un server
// Implementazione del multicast totalmente ordinato -> Il server ha inviato in multicast il messaggio di update
func (kvs *KeyValueStoreSequential) TotalOrderedMulticast(message MessageS, response *common.Response) error {

	// Aggiunta della richiesta in coda, solo se sono un ricevente contattato dal server
	/*if kvs.id != message.IdSender {
		kvs.addToSortQueue(message)
	}*/

	kvs.addToSortQueue(message)

	// Aggiornamento del clock -> Prendo il max timestamp tra il mio e quello allegato al messaggio ricevuto
	kvs.mutexClock.Lock()
	//kvs.logicalClock = common.Max(message.LogicalClock, kvs.logicalClock)
	if kvs.Id != message.IdSender {
		//kvs.logicalClock++ // Devo incrementare il clock per gestire l'evento di receive ? TODO: ???
		kvs.printDebugBlue("RICEVUTO da server", message)
	}
	kvs.mutexClock.Unlock()

	// Invio ack a tutti i server per notificare la ricezione della richiesta
	sendAck(message)

	// Ciclo finché controlSendToApplication non restituisce true
	// Controllo quando la richiesta può essere eseguita a livello applicativo
	response.Result = false
	for {
		canSend := kvs.controlSendToApplication(message)
		if canSend {
			// Invio a livello applicativo
			err := kvs.realFunction(message, response)
			if err != nil {
				return err
			}
			break
		}

		// Altrimenti, attendi un breve periodo prima di riprovare
		time.Sleep(time.Millisecond * 100)
		//printQueue(kvs)
		//printDatastore(kvs)
	}

	if response.Result {
		return nil
	}
	return fmt.Errorf("TotalOrderedMulticast non andato a buon fine")
}

// ReceiveAck gestisce gli ack dei messaggi ricevuti.
// Se il messaggio è presente nella coda incrementa il numero di ack e restituisce true, altrimenti restituisce false
func (kvs *KeyValueStoreSequential) ReceiveAck(message MessageS, reply *bool) error {
	//fmt.Println("ReceiveAck: Ho ricevuto un ack", message.TypeOfMessage, message.Args.Key, ":", message.Args.Value)
	*reply = kvs.updateMessage(message)
	return nil
}

// addToSortQueue aggiunge un messaggio alla coda locale, ordinandola in base al timeStamp, a parità di timestamp
// l'ordinamento è deterministico: per garantire l'ordinamento totale verranno usati gli ID associati al messaggio.
// La funzione è threadSafe per l'utilizzo della coda kvs.queue tramite kvs.mutexQueue
func (kvs *KeyValueStoreSequential) addToSortQueue(message MessageS) {
	kvs.mutexQueue.Lock()
	kvs.Queue = append(kvs.Queue, message)

	// Ordina la coda in base al logicalClock
	sort.Slice(kvs.Queue, func(i, j int) bool {
		if kvs.Queue[i].LogicalClock == kvs.Queue[j].LogicalClock {
			return kvs.Queue[i].Id < kvs.Queue[j].Id
		}
		return kvs.Queue[i].LogicalClock < kvs.Queue[j].LogicalClock
	})

	kvs.mutexQueue.Unlock()
}

// removeMessageToQueue Rimuove un messaggio dalla coda basato sull'ID
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
// timestamp potenzialmente minore o uguale a quello di msg_i) TODO: >=
func (kvs *KeyValueStoreSequential) controlSendToApplication(message MessageS) bool {
	if kvs.Queue[0].Id == message.Id && kvs.Queue[0].NumberAck >= common.Replicas {

		// Aggiornamento del clock -> Prendo il max timestamp tra il mio e quello allegato al messaggio ricevuto
		kvs.mutexClock.Lock()
		kvs.LogicalClock = common.Max(message.LogicalClock, kvs.LogicalClock)
		if kvs.Id != message.IdSender {
			kvs.LogicalClock++ // Devo incrementare il clock per gestire l'evento di receive ? TODO: ???
		}
		kvs.mutexClock.Unlock()

		// Ho ricevuto tutti gli ack, posso eliminare il messaggio dalla coda
		kvs.removeMessageToQueue(message)
		return true
	}
	return false
}

// sendAck invia a tutti i server un Ack
func sendAck(message MessageS) {
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

			// Invio controllando il valore, se il server mi risponde false provo a ricontattarlo. (Immediatamente)
			for {
				err = sendAckRPC(conn, message, &reply)
				if err != nil {
					fmt.Printf("sendAck: Errore durante la chiamata RPC receiveAck %v\n", err)
					return
				}

				if reply {
					break
				}
				//fmt.Printf("sendAck: il server %s non è riuscito ad incrementare il numero di ack!\n", replicaPort)
			}

			// Chiudo la connessione dopo essere sicuro che l'ack è stato inviato
			err = conn.Close()
			if err != nil {
				fmt.Println("Errore durante la chiusura della connessione in TotalOrderedMulticast.sendAckRPC:", err)
				return
			}
		}(common.ReplicaPorts[i], i)
	}
}

// sendAckRPC invia l'ack tramite RPC, applicando un ritardo random
func sendAckRPC(conn *rpc.Client, message MessageS, reply *bool) error {
	common.RandomDelay()
	err := conn.Call("KeyValueStoreSequential.ReceiveAck", message, reply)
	if err != nil {
		return err
	}
	return nil
}
