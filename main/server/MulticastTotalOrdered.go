// MulticastTotalOrdered.go Totally Ordered Multicast

package main

import (
	"fmt"
	"main/common"
	"net/rpc"
	"os"
	"sort"
	"strconv"
	"time"
)

// Algoritmo distribuito MULTICAST TOT. ORDINATO:
// Ogni messaggio abbia come timestamp il clock logico scalare del processo che lo invia.
//1. pi invia in multicast (incluso se stesso) il messaggio di update msg_i.
//2. msg_i viene posto da ogni processo destinatario pj in una coda locale queue_j, ordinata in base al valore del timestamp.
//3. pj invia in multicast un messaggio di ack della ricezione di msg_i.
//4. pj consegna msg_i all’applicazione se msg_i è in testa a queue_j, tutti gli ack relativi a msg_i sono stati ricevuti
//	da pj e, per ogni processo pk, c’è un messaggio msg_k in queue_j con timestamp maggiore di quello di msg_i
//(quest’ultima condizione sta a indicare che nessun altro processo può inviare in multicast un messaggio con
//timestamp potenzialmente minore o uguale a quello di msg_i).

// MulticastTotalOrdered gestione dell'evento esterno ricevuto da un server
func (kvs *KeyValueStoreSequential) MulticastTotalOrdered(message Message, reply *bool) error {
	// Implementazione del multicast totalmente ordinato -> Il server ha inviato in multicast il messaggio di update
	fmt.Println("MulticastTotalOrdered: Ho ricevuto la richiesta che mi è stata inoltrata da un server", message.TypeOfMessage, message.Args.Key, ":", message.Args.Value)

	// Aggiunta della richiesta in coda
	kvs.addToSortQueue(message)

	// Aggiornamento del clock -> Prendo il max timestamp tra il mio e quello allegato al messaggio ricevuto
	kvs.mutexClock.Lock()
	kvs.logicalClock = common.Max(message.LogicalClock, kvs.logicalClock)
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
	}

	if *reply {
		return nil
	}
	return fmt.Errorf("MulticastTotalOrdered non andato a buon fine")
}

// ReceiveAck gestisce gli ack dei messaggi ricevuti.
// Se il messaggio è presente nella coda incrementa il numero di ack e restituisce true, altrimenti restituisce false
func (kvs *KeyValueStoreSequential) ReceiveAck(message Message, reply *bool) error {
	fmt.Println("ReceiveAck: Ho ricevuto un ack", message.TypeOfMessage, message.Args.Key, ":", message.Args.Value)

	// Aggiorna il messaggio nella coda
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

// controlSendToApplication verifica se è possibile inviare la richiesta a livello applicativo
// 4. pj consegna msg_i all’applicazione se msg_i è in testa a queue_j, tutti gli ack relativi a msg_i sono stati ricevuti
// da pj e, per ogni processo pk, c’è un messaggio msg_k in queue_j con timestamp maggiore di quello di msg_i
// (quest’ultima condizione sta a indicare che nessun altro processo può inviare in multicast un messaggio con
// timestamp potenzialmente minore o uguale a quello di msg_i)
func (kvs *KeyValueStoreSequential) controlSendToApplication(message Message) bool {
	if kvs.queue[0].Id == message.Id && kvs.queue[0].NumberAck >= common.Replicas {
		//fmt.Println("controlSendToApplication: Ho ricevuto tutti gli ack, posso eliminare il messaggio dalla mia coda")
		kvs.removeMessageToQueue(message)
		return true
	}
	return false
}

// sendAck invia a tutti i server un Ack
func (kvs *KeyValueStoreSequential) sendAck(message Message) {
	fmt.Println("sendAck: Invio un ack a tutti specificando il messaggio ricevuto", message.TypeOfMessage, message.Args.Key, ":", message.Args.Value)

	reply := false

	for i := 0; i < common.Replicas; i++ {
		i := i
		go func(replicaPort string) { // Invio tutti gli ack in maniera asincrona

			var serverName string

			if os.Getenv("CONFIG") == "1" {
				/*---LOCALE---*/
				serverName = ":" + replicaPort
			} else if os.Getenv("CONFIG") == "2" {
				/*---DOCKER---*/
				serverName = "server" + strconv.Itoa(i+1) + ":" + replicaPort
			} else {
				fmt.Println("VARIABILE DI AMBIENTE ERRATA")
				return
			}

			//fmt.Println("sendAck: Contatto il server:", serverName)
			conn, err := rpc.Dial("tcp", serverName)
			if err != nil {
				fmt.Printf("sendAck: Errore durante la connessione al server:"+replicaPort, err)
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
					// reply false vuol dire che il server contattato ha ricevuto un ack di una richiesta a lui non nota
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
func (kvs *KeyValueStoreSequential) removeMessageToQueue(message Message) {
	kvs.mutexQueue.Lock()
	defer kvs.mutexQueue.Unlock()
	if kvs.queue[0].Id == message.Id {
		// Rimuovi l'elemento dalla slice
		kvs.queue = kvs.queue[1:]
		return
	}
	//fmt.Println("removeByID: Messaggio con ID", message.Id, "non trovato nella coda")
}

// updateMessage aggiorna il messaggio in coda corrispondente all'id del messaggio passato in argomento
func (kvs *KeyValueStoreSequential) updateMessage(message Message) bool {
	for i := range kvs.queue {
		if kvs.queue[i].Id == message.Id && kvs.queue[i].LogicalClock == message.LogicalClock {
			kvs.mutexQueue.Lock()
			kvs.queue[i].NumberAck++
			kvs.mutexQueue.Unlock()

			fmt.Println("NumeroAck", kvs.queue[i].NumberAck, "di", message.TypeOfMessage, message.Args.Key, ":", message.Args.Value)
			return true
		}
	}
	return false
}
