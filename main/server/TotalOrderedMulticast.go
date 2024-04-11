package main

import (
	"fmt"
	"github.com/fatih/color"
	"main/common"
	"net/rpc"
	"sort"
	"time"
)

// TotalOrderedMulticast gestione dell'evento esterno ricevuto da un server
func (kvs *KeyValueStoreSequential) TotalOrderedMulticast(message Message, reply *bool) error {
	// Implementazione del multicast totalmente ordinato -> Il server ha inviato in multicast il messaggio di update
	//fmt.Println("TotalOrderedMulticast: Ho ricevuto la richiesta che mi è stata inoltrata da un server", message.TypeOfMessage, message.Args.Key, ":", message.Args.Value)

	// Aggiunta della richiesta in coda
	kvs.addToSortQueue(message)

	// Aggiornamento del clock -> Prendo il max timestamp tra il mio e quello allegato al messaggio ricevuto
	kvs.mutexClock.Lock()
	kvs.logicalClock = common.Max(message.LogicalClock, kvs.logicalClock)
	fmt.Println(color.BlueString("RICEVUTO"), message.TypeOfMessage, message.Args.Key+":"+message.Args.Value, "with clock", message.LogicalClock, "my clock", kvs.logicalClock)
	kvs.mutexClock.Unlock()

	// Invio ack a tutti i server per notificare la ricezione della richiesta
	kvs.sendAck(message)

	// Ciclo finché controlSendToApplication non restituisce true
	// Controllo quando la richiesta può essere eseguita a livello applicativo
	*reply = false
	for {
		canSend := kvs.controlSendToApplication(message)
		if canSend {

			// Invio a livello applicativo
			var replySaveInDatastore common.Response
			err := kvs.RealFunction(message, &replySaveInDatastore)
			if err != nil {
				return err
			}

			*reply = true
			break
		}

		// Altrimenti, attendi un breve periodo prima di riprovare
		time.Sleep(time.Millisecond * 100)
		//printQueue(kvs)
		//printDatastore(kvs)
	}

	if *reply {
		return nil
	}
	return fmt.Errorf("TotalOrderedMulticast non andato a buon fine")
}

// ReceiveAck gestisce gli ack dei messaggi ricevuti.
// Se il messaggio è presente nella coda incrementa il numero di ack e restituisce true, altrimenti restituisce false
func (kvs *KeyValueStoreSequential) ReceiveAck(message Message, reply *bool) error {
	//fmt.Println("ReceiveAck: Ho ricevuto un ack", message.TypeOfMessage, message.Args.Key, ":", message.Args.Value)
	*reply = kvs.updateMessage(message)
	return nil
}

// addToSortQueue aggiunge un messaggio alla coda locale, ordinandola in base al timeStamp, a parità di timestamp
// l'ordinamento è deterministico, per garantire ordinamento totale, ordinandolo in base all'id.
func (kvs *KeyValueStoreSequential) addToSortQueue(message Message) {
	kvs.mutexQueue.Lock()
	defer kvs.mutexQueue.Unlock()
	kvs.queue = append(kvs.queue, message)

	// Ordina la coda in base al logicalClock
	sort.Slice(kvs.queue, func(i, j int) bool {
		if kvs.queue[i].LogicalClock == kvs.queue[j].LogicalClock {
			return kvs.queue[i].Id < kvs.queue[j].Id
		}
		return kvs.queue[i].LogicalClock < kvs.queue[j].LogicalClock
	})
}

// removeMessageToQueue Rimuove un messaggio dalla coda basato sull'ID
func (kvs *KeyValueStoreSequential) removeMessageToQueue(message Message) {
	kvs.mutexQueue.Lock()
	defer kvs.mutexQueue.Unlock()
	if kvs.queue[0].Id == message.Id {
		// Rimuovi l'elemento dalla slice
		kvs.queue = kvs.queue[1:]
		return
	}
	//fmt.Println("removeMessageToQueue: Messaggio con ID", message.Id, "non trovato nella coda")
}

// updateMessage aggiorna, incrementando il numero di ack ricevuti, il messaggio in coda corrispondente all'id del messaggio passato come argomento
func (kvs *KeyValueStoreSequential) updateMessage(message Message) bool {
	// Aggiorna il messaggio nella coda
	kvs.mutexQueue.Lock()
	defer kvs.mutexQueue.Unlock()
	for i := range kvs.queue {
		if kvs.queue[i].Id == message.Id {
			kvs.queue[i].NumberAck++
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
func (kvs *KeyValueStoreSequential) controlSendToApplication(message Message) bool {
	if kvs.queue[0].Id == message.Id && kvs.queue[0].NumberAck >= common.Replicas {
		// Ho ricevuto tutti gli ack, posso eliminare il messaggio dalla coda
		kvs.removeMessageToQueue(message)
		return true
	}
	return false
}

// sendAck invia a tutti i server un Ack
func (kvs *KeyValueStoreSequential) sendAck(message Message) {
	for i := 0; i < common.Replicas; i++ {
		go func(replicaPort string, index int) {

			reply := false
			serverName := common.GetServerName(replicaPort, index)
			conn, err := rpc.Dial("tcp", serverName)
			if err != nil {
				fmt.Printf("sendAck: Errore durante la connessione al server %s: %v\n", replicaPort, err)
				return
			}

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

// sendAckRPC invia l'ack tramite RPC e gestisce eventuali errori
func sendAckRPC(conn *rpc.Client, message Message, reply *bool) error {
	common.RandomDelay()
	err := conn.Call("KeyValueStoreSequential.ReceiveAck", message, reply)
	if err != nil {
		return err
	}
	return nil
}

func (kvs *KeyValueStoreSequential) findMessage(message Message) *Message {
	for i := range kvs.queue {
		if kvs.queue[i].Id == message.Id && kvs.queue[i].LogicalClock == message.LogicalClock {
			return &kvs.queue[i]
		}
	}
	return &Message{}
}
