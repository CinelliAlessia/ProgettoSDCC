// MulticastTotalOrdered.go

package main

import (
	"fmt"
	"main/common"
	"net/rpc"
	"sort"
	"sync"
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

type MulticastTotalOrdered struct {
	queue []Message
	mu    sync.Mutex // Mutex per proteggere l'accesso concorrente alla coda
}

// SendToEveryone non fa come dice, ma riceve già un messaggio.
func (mto *MulticastTotalOrdered) SendToEveryone(message Message, reply *bool) error {
	// Implementazione del multicast totalmente ordinato -> Il server ha inviato in multicast il messaggio di update
	fmt.Println("MulticastTotalOrdered: Ho ricevuto la richiesta che mi è stata inoltrata da un server")

	// Gestisce i messaggi in ingresso aggiungendo i messaggi in coda e inviando gli ack
	mto.mu.Lock()
	mto.queue = append(mto.queue, message)

	// Ordina la coda in base al logicalClock
	sort.Slice(mto.queue, func(i, j int) bool {
		return mto.queue[i].LogicalClock < mto.queue[j].LogicalClock
	})
	mto.mu.Unlock()

	mto.printMessageQueue() // DEBUG
	mto.sendAck(message)

	//fmt.Println("MulticastTotalOrdered: Messaggio in coda " + message.Id)
	//fmt.Println("MulticastTotalOrdered: Messaggio in coda " + mto.queue[0].Id)

	// Ciclo finché controlSendToApplication non restituisce true
	for {
		canSend := mto.controlSendToApplication(&message)
		if canSend {
			// TODO: Invio a livello applicativo
			/*var replySaveInDatastore common.Response
			err := sendToOtherServer("KeyValueStoreSequential.RealFunction", message, &replySaveInDatastore)
			if err != nil {
				// VEDI UN PO'
			}*/
			*reply = true
			break // Esci dal ciclo se controlSendToApplication restituisce true
		}
		// Altrimenti, attendi un breve periodo prima di riprovare
		time.Sleep(time.Millisecond * 100)
	}

	return nil
}

func (mto *MulticastTotalOrdered) controlSendToApplication(message *Message) bool {
	// Controlla se il messaggio è in testa alla coda locale
	//fmt.Printf("AAA %d\n", message.NumberAck)

	//4. pj consegna msg_i all’applicazione se msg_i è in testa a queue_j, tutti gli ack relativi a msg_i sono stati ricevuti
	//	da pj e, per ogni processo pk, c’è un messaggio msg_k in queue_j con timestamp maggiore di quello di msg_i
	//(quest’ultima condizione sta a indicare che nessun altro processo può inviare in multicast un messaggio con
	//timestamp potenzialmente minore o uguale a quello di msg_i).
	if mto.queue[0].Id == message.Id && mto.queue[0].NumberAck == common.Replicas {
		// Invia il messaggio all'applicazione
		fmt.Println("MulticastTotalOrdered-controlSendToApplication: Ho ricevuto tutti gli ack, posso eliminare il messaggio dalla mia coda")
		//mto.queue.remo
		return true
	}
	return false
}

// sendAck invia a tutti i server un Ack
func (mto *MulticastTotalOrdered) sendAck(message Message) {
	fmt.Println("MulticastTotalOrdered-sendAck: Invio un ack a tutti specificando il messaggio ricevuto")
	for i := 0; i < common.Replicas; i++ {
		go func(replicaPort string) {
			// Connessione al server RPC
			server, err := rpc.Dial("tcp", ":"+replicaPort)
			if err != nil {
				fmt.Println("MulticastTotalOrdered-sendAck: Errore durante la connessione al server:", err)
			}

			reply := false
			// Chiama il metodo Multiply sul server RPC
			err = server.Call("MulticastTotalOrdered.ReceiveAck", message, &reply)
			if err != nil {
				fmt.Println("MulticastTotalOrdered-sendAck: Errore durante la chiamata RPC ReceiveAck:", err)
			}
		}(common.ReplicaPorts[i])
	}
}

// ReceiveAck gestisce gli ack dei messaggi ricevuti.
func (mto *MulticastTotalOrdered) ReceiveAck(message Message, _ *bool) error {

	// Trova il messaggio nella coda
	newMessage := mto.findByID(message.Id)
	if newMessage.Id == "" {
		return fmt.Errorf("MulticastTotalOrdered-ReceiveAck: AAA messaggio %s non trovato nella coda", message.Id)
	}

	// Incrementa il conteggio degli ack
	newMessage.NumberAck++
	fmt.Println("MulticastTotalOrdered-ReceiveAck: Ho ricevuto un ack")

	// Aggiorna il messaggio nella coda
	mto.updateMessageByID(newMessage)
	return nil
}

// findByID Ritorna un messaggio cercandolo by id
func (mto *MulticastTotalOrdered) findByID(id string) Message {
	//fmt.Println("MulticastTotalOrdered-findByID: ID associato al messaggio " + id)
	for i := range mto.queue {
		//fmt.Println("MulticastTotalOrdered-findByID: paragone con " + mto.queue[i].Id)

		if mto.queue[i].Id == id {
			fmt.Println("MulticastTotalOrdered-findByID: TROVATO")
			return mto.queue[i]
		}
	}
	return Message{}
}

// updateMessageByID aggiorna il messaggio in coda corrispondente all'id del messaggio passato in argomento
func (mto *MulticastTotalOrdered) updateMessageByID(newMessage Message) {
	for i := range mto.queue {
		if mto.queue[i].Id == newMessage.Id {
			mto.mu.Lock()
			mto.queue[i] = newMessage
			mto.mu.Unlock()
		}
	}
}

// printMessageQueue è una funzione di debug che stampa la coda
func (mto *MulticastTotalOrdered) printMessageQueue() {
	for i := range mto.queue {
		fmt.Println("Id " + mto.queue[i].Id + "Value " + mto.queue[i].Args.Value)
	}
}
