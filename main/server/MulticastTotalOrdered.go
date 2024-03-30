// MulticastTotalOrdered.go

package main

import (
	"fmt"
	"main/common"
	"net/rpc"
	"sort"
	"time"
)

// Algoritmo distribuito MULTICAST TOT. ORDINATO:
// Ogni messaggio abbia come timestamp il clock logico scalare del processo che lo invia. V
//1. pi invia in multicast (incluso se stesso) il messaggio di update msg_i. V
//2. msg_i viene posto da ogni processo destinatario pj in una coda locale queue_j, ordinata in base al valore del timestamp. V
//3. pj invia in multicast un messaggio di ack della ricezione di msg_i.
//4. pj consegna msg_i all’applicazione se msg_i è in testa a queue_j, tutti gli ack relativi a msg_i sono stati ricevuti
//	da pj e, per ogni processo pk, c’è un messaggio msg_k in queue_j con timestamp maggiore di quello di msg_i
//(quest’ultima condizione sta a indicare che nessun altro processo può inviare in multicast un messaggio con
//timestamp potenzialmente minore o uguale a quello di msg_i).

// MulticastTotalOrdered gestione dell'evento esterno ricevuto da un server
func (kvs *KeyValueStoreSequential) MulticastTotalOrdered(message Message, reply *bool) error {

	// Implementazione del multicast totalmente ordinato -> Il server ha inviato in multicast il messaggio di update
	fmt.Println("MulticastTotalOrdered: Ho ricevuto la richiesta che mi è stata inoltrata da un server")

	// Aggiunta della richiesta in coda
	kvs.addToSortQueue(message)

	// Aggiornamento del clock
	kvs.mutexClock.Lock()
	kvs.logicalClock = myMax(message.LogicalClock, kvs.logicalClock)
	kvs.mutexClock.Unlock()

	// Invio ack a tutti i server
	kvs.sendAck(message)

	// Ciclo finché controlSendToApplication non restituisce true
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
			break // Esci dal ciclo se controlSendToApplication restituisce true
		}

		// Altrimenti, attendi un breve periodo prima di riprovare
		time.Sleep(time.Millisecond * 100)
	}

	return nil
}

func myMax(clock int, clock2 int) int {
	if clock > clock2 {
		return clock
	} else {
		return clock2
	}
}

// ReceiveAck gestisce gli ack dei messaggi ricevuti.
/*
func (kvs *KeyValueStoreSequential) ReceiveAck(message Message, reply *bool) error {

	// Trova il messaggio nella coda
	newMessage := kvs.findByID(message.Id)
	if newMessage.Id == "" {
		// PROBLEMA RISOLTO: ricevo prima l'ack che la notifica che c'è un evento dovrei aggiungerlo ma ci sarebbero problemi
		// se me ne arrivano due di ack prima della richiesta? :( -> risolto, viene rinviato l'ack se io rispondo false
		*reply = false
		return nil
		// kvs.addToSortQueue(message)
	}

	// Incrementa il conteggio degli ack

	// TODO: partono due ReceiveAck in contemporanea, entrambi incrementano l'ack a due e entrambi lo impostano a 2. ma uno era 2 e l'altro 3
	// devo ricevere un puntatore al messaggio e incrementarlo con i lucchetti, cosi non va bene.
	kvs.updateMessageByID(message)
	fmt.Println("ReceiveAck: Ho ricevuto un ack")

	// Aggiorna il messaggio nella coda
	//kvs.updateMessageByID(newMessage)

	*reply = true
	return nil
} */

// ReceiveAck gestisce gli ack dei messaggi ricevuti.
// Se il messaggio è presente nella coda incrementa il numero di ack e restituisce true, altrimenti restituisce false
func (kvs *KeyValueStoreSequential) ReceiveAck(message Message, reply *bool) error {
	fmt.Println("ReceiveAck: Ho ricevuto un ack")

	// Aggiorna il messaggio nella coda
	*reply = kvs.updateMessage(message)

	return nil
}

func (kvs *KeyValueStoreSequential) addToSortQueue(message Message) {
	kvs.mu.Lock()
	kvs.queue = append(kvs.queue, message)

	// Ordina la coda in base al logicalClock
	sort.Slice(kvs.queue, func(i, j int) bool {
		return kvs.queue[i].LogicalClock < kvs.queue[j].LogicalClock
	})
	kvs.mu.Unlock()
}

// controlSendToApplication
func (kvs *KeyValueStoreSequential) controlSendToApplication(message Message) bool {
	//4. pj consegna msg_i all’applicazione se msg_i è in testa a queue_j, tutti gli ack relativi a msg_i sono stati ricevuti
	//da pj e, per ogni processo pk, c’è un messaggio msg_k in queue_j con timestamp maggiore di quello di msg_i
	//(quest’ultima condizione sta a indicare che nessun altro processo può inviare in multicast un messaggio con
	//timestamp potenzialmente minore o uguale a quello di msg_i).
	fmt.Println("controlSendToApplication: Controllo")
	if kvs.queue[0].Id == message.Id && kvs.queue[0].NumberAck == common.Replicas /*&& kvs.queue[0].LogicalClock == message.LogicalClock*/ {
		// Invia il messaggio all'applicazione
		fmt.Println("controlSendToApplication: Ho ricevuto tutti gli ack, posso eliminare il messaggio dalla mia coda")
		kvs.removeMessage(message)
		return true
	} else {
		fmt.Println("ID ", kvs.queue[0].Id, message.Id, "Ack ", kvs.queue[0].NumberAck, common.Replicas)
	}
	return false
}

// sendAck invia a tutti i server un Ack
func (kvs *KeyValueStoreSequential) sendAck(message Message) {
	fmt.Println("sendAck: Invio un ack a tutti specificando il messaggio ricevuto")

	reply := false

	for i := 0; i < common.Replicas; i++ {
		go func(replicaPort string) {

			conn, err := rpc.Dial("tcp", ":"+replicaPort)
			if err != nil {
				fmt.Printf("sendAck: Errore durante la connessione al server "+replicaPort+": ", err)
				return
			}

			for {
				err = conn.Call("KeyValueStoreSequential.ReceiveAck", message, &reply)
				if err != nil {
					fmt.Println("sendAck: Errore durante la chiamata RPC receiveAck ", err)
					return
				}

				if reply {
					break // Esci dal ciclo se reply è true
				}
			}
		}(common.ReplicaPorts[i])
	}
}

// printMessageQueue è una funzione di debug che stampa la coda
func (kvs *KeyValueStoreSequential) printMessageQueue() {
	for i := range kvs.queue {
		fmt.Println("Id: ", kvs.queue[i].Id, "Value: ", kvs.queue[i].Args.Value, "ACK: ", kvs.queue[i].NumberAck)
	}
}

func (kvs *KeyValueStoreSequential) findMessage(message Message) *Message {
	for i := range kvs.queue {
		if kvs.queue[i].Id == message.Id && kvs.queue[i].LogicalClock == message.LogicalClock {
			return &kvs.queue[i]
		}
	}
	return &Message{}
}

// removeByID Rimuove un messaggio dalla coda basato sull'ID
func (kvs *KeyValueStoreSequential) removeMessage(message Message) {
	for i := range kvs.queue {
		if kvs.queue[i].Id == message.Id && kvs.queue[i].LogicalClock == message.LogicalClock {
			// Rimuovi l'elemento dalla slice
			kvs.queue = append(kvs.queue[:i], kvs.queue[i+1:]...)
			fmt.Println("removeByID: Messaggio con ID", message.Id, "rimosso dalla coda")
			return
		}
	}
	fmt.Println("removeByID: Messaggio con ID", message.Id, "non trovato nella coda")
}

// updateMessageByID aggiorna il messaggio in coda corrispondente all'id del messaggio passato in argomento
func (kvs *KeyValueStoreSequential) updateMessage(message Message) bool {
	for i := range kvs.queue {
		if kvs.queue[i].Id == message.Id && kvs.queue[i].LogicalClock == message.LogicalClock {
			kvs.mu.Lock()
			kvs.queue[i].NumberAck++
			kvs.mu.Unlock()

			fmt.Println("NumeroAck ", kvs.queue[i].NumberAck)
			return true
		}
	}
	return false
}
