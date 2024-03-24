package main

import (
	"fmt"
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

type TotalOrderedMulticast struct {
	queue []Message
}

func (mto *TotalOrderedMulticast) MulticastTotalOrdered(message Message, reply *bool) error {
	// Implementazione del multicast totalmente ordinato -> Il server ha inviato in multicast il messaggio di update

	// Gestisce i messaggi in ingresso aggiungendo i messaggi in coda e inviando gli ack
	mto.queue = append(mto.queue, message)
	// Ordina la coda in base al logicalClock
	sort.Slice(mto.queue, func(i, j int) bool {
		return mto.queue[i].logicalClock < mto.queue[j].logicalClock
	})

	// Ciclo finché controlSendToApplication non restituisce true
	for {
		canSend := mto.controlSendToApplication(&message)
		if canSend {
			*reply = true
			break // Esci dal ciclo se controlSendToApplication restituisce true
		}
		// Altrimenti, attendi un breve periodo prima di riprovare
		time.Sleep(time.Millisecond * 100)
	}

	return nil
}

func (mto *TotalOrderedMulticast) controlSendToApplication(message *Message) bool {
	// Controlla se il messaggio è in testa alla coda locale

	//4. pj consegna msg_i all’applicazione se msg_i è in testa a queue_j, tutti gli ack relativi a msg_i sono stati ricevuti
	//	da pj e, per ogni processo pk, c’è un messaggio msg_k in queue_j con timestamp maggiore di quello di msg_i
	//(quest’ultima condizione sta a indicare che nessun altro processo può inviare in multicast un messaggio con
	//timestamp potenzialmente minore o uguale a quello di msg_i).

	if mto.queue[0].id == message.id && message.numberAck == Replicas {
		// Invia il messaggio all'applicazione
		return true
	}
	return false
}

func (mto *TotalOrderedMulticast) sendAck(message Message) {
	for i := 0; i < Replicas; i++ {
		// Connessione al server RPC
		client, err := rpc.Dial("tcp", ":"+ReplicaPorts[i])
		if err != nil {
			fmt.Println("Errore durante la connessione al server:", err)
		}

		// Chiama il metodo Multiply sul server RPC
		err = client.Call("TotalOrderedMulticast.ReceiveAck", message, false)
		if err != nil {
			fmt.Println("Errore durante la chiamata RPC:", err)
		}

	}
}

// ReceiveAck gestisce gli ack dei messaggi ricevuti.
func (mto *TotalOrderedMulticast) ReceiveAck(message Message, _ bool) error {
	// Trova il messaggio nella coda
	newMessage := mto.findByID(message.id)
	if newMessage == nil {
		return fmt.Errorf("messaggio %s non trovato nella coda", message.id)
	}

	// Incrementa il conteggio degli ack
	newMessage.numberAck++

	// Aggiorna il messaggio nella coda
	mto.updateMessageByID(newMessage)
	return nil
}

func (mto *TotalOrderedMulticast) findByID(id string) *Message {
	for i := range mto.queue {
		if mto.queue[i].id == id {
			return &mto.queue[i]
		}
	}
	return nil
}

func (mto *TotalOrderedMulticast) updateMessageByID(newMessage *Message) {
	for i := range mto.queue {
		if mto.queue[i].id == newMessage.id {
			mto.queue[i] = *newMessage
		}
	}
}
