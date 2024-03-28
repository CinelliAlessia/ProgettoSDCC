// KeyValueStoreSequential.go

package main

import (
	"fmt"
	"main/common"
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

// MulticastTotalOrdered non fa come dice, ma riceve già un messaggio.
func (kvs *KeyValueStoreSequential) MulticastTotalOrdered(message Message, reply *bool) error {
	// Implementazione del multicast totalmente ordinato -> Il server ha inviato in multicast il messaggio di update
	fmt.Println("MulticastTotalOrdered: Ho ricevuto la richiesta che mi è stata inoltrata da un server")

	// Gestisce i messaggi in ingresso aggiungendo i messaggi in coda e inviando gli ack
	kvs.mu.Lock()
	kvs.queue = append(kvs.queue, message)

	// Ordina la coda in base al logicalClock
	sort.Slice(kvs.queue, func(i, j int) bool {
		return kvs.queue[i].LogicalClock < kvs.queue[j].LogicalClock
	})
	kvs.mu.Unlock()

	kvs.mutexClock.Lock()
	kvs.logicalClock = max(message.LogicalClock, kvs.logicalClock)
	kvs.mutexClock.Unlock()

	kvs.printMessageQueue() // DEBUG
	kvs.sendAck(message)

	//fmt.Println("KeyValueStoreSequential: Messaggio in coda " + message.Id)
	//fmt.Println("KeyValueStoreSequential: Messaggio in coda " + kvs.queue[0].Id)

	// Ciclo finché controlSendToApplication non restituisce true
	for {
		canSend := kvs.controlSendToApplication(&message)
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

// ReceiveAck gestisce gli ack dei messaggi ricevuti.
func (kvs *KeyValueStoreSequential) ReceiveAck(message Message, reply *bool) error {

	// Trova il messaggio nella coda
	newMessage := kvs.findByID(message.Id)
	if newMessage.Id == "" {
		return fmt.Errorf("ReceiveAck: AAA messaggio %s non trovato nella coda", message.Id)
	}

	// Incrementa il conteggio degli ack
	newMessage.NumberAck++
	fmt.Println("ReceiveAck: Ho ricevuto un ack")

	// Aggiorna il messaggio nella coda
	kvs.updateMessageByID(newMessage)
	return nil
}

// controlSendToApplication
func (kvs *KeyValueStoreSequential) controlSendToApplication(message *Message) bool {
	//4. pj consegna msg_i all’applicazione se msg_i è in testa a queue_j, tutti gli ack relativi a msg_i sono stati ricevuti
	//	da pj e, per ogni processo pk, c’è un messaggio msg_k in queue_j con timestamp maggiore di quello di msg_i
	//(quest’ultima condizione sta a indicare che nessun altro processo può inviare in multicast un messaggio con
	//timestamp potenzialmente minore o uguale a quello di msg_i).
	if kvs.queue[0].Id == message.Id && kvs.queue[0].NumberAck == common.Replicas {
		// Invia il messaggio all'applicazione
		fmt.Println("controlSendToApplication: Ho ricevuto tutti gli ack, posso eliminare il messaggio dalla mia coda")
		kvs.removeByID(message.Id)
		return true
	}
	return false
}

// sendAck invia a tutti i server un Ack
func (kvs *KeyValueStoreSequential) sendAck(message Message) {
	fmt.Println("sendAck: Invio un ack a tutti specificando il messaggio ricevuto")
	reply := false
	err := sendToOtherServer("KeyValueStoreSequential.ReceiveAck", message, &reply)
	if err != nil {
		return
	}
}

// findByID Ritorna un messaggio cercandolo by id
func (kvs *KeyValueStoreSequential) findByID(id string) Message {
	//fmt.Println("findByID: ID associato al messaggio " + id)
	for i := range kvs.queue {
		//fmt.Println("findByID: paragone con " + kvs.queue[i].Id)

		if kvs.queue[i].Id == id {
			fmt.Println("findByID: TROVATO")
			return kvs.queue[i]
		}
	}
	return Message{}
}

// removeByID Rimuove un messaggio dalla coda basato sull'ID
func (kvs *KeyValueStoreSequential) removeByID(id string) {
	for i := range kvs.queue {
		if kvs.queue[i].Id == id {
			// Rimuovi l'elemento dalla slice
			kvs.queue = append(kvs.queue[:i], kvs.queue[i+1:]...)
			fmt.Println("removeByID: Messaggio con ID", id, "rimosso dalla coda")
			return
		}
	}
	fmt.Println("removeByID: Messaggio con ID", id, "non trovato nella coda")
}

// updateMessageByID aggiorna il messaggio in coda corrispondente all'id del messaggio passato in argomento
func (kvs *KeyValueStoreSequential) updateMessageByID(newMessage Message) {
	for i := range kvs.queue {
		if kvs.queue[i].Id == newMessage.Id {
			kvs.mu.Lock()
			kvs.queue[i] = newMessage
			kvs.mu.Unlock()
		}
	}
}

// printMessageQueue è una funzione di debug che stampa la coda
func (kvs *KeyValueStoreSequential) printMessageQueue() {
	for i := range kvs.queue {
		fmt.Println("Id " + kvs.queue[i].Id + "Value " + kvs.queue[i].Args.Value)
	}
}
