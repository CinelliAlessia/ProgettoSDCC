package main

import (
	"fmt"
	"github.com/google/uuid"
	"net/rpc"
	"sync"
)

// KeyValueStoreSequential KeyValueStoreCausale rappresenta il servizio di memorizzazione chiave-valore
type KeyValueStoreSequential struct {
	dataStore    map[string]string // Mappa -> struttura dati che associa chiavi a valori
	logicalClock int               // Orologio logico scalare
	mutexClock   sync.Mutex
}

type Message struct {
	id            string
	typeOfMessage string
	Args
	logicalClock int
	numberAck    int
}

// Get restituisce il valore associato alla chiave specificata -> Lettura -> Evento interno
func (kvs *KeyValueStoreSequential) Get(args Args, response *Response) error {

	kvs.mutexClock.Lock()

	kvs.logicalClock++
	val, ok := kvs.dataStore[args.key]

	kvs.mutexClock.Unlock()

	if !ok {
		return fmt.Errorf("key '%s' not found", args.key)
	}

	response.reply = val
	return nil
}

// Put inserisce una nuova coppia chiave-valore, se la chiave è già presente, sovrascrive il valore associato
func (kvs *KeyValueStoreSequential) Put(args Args, response *Response) error {

	kvs.mutexClock.Lock()
	kvs.logicalClock++
	kvs.mutexClock.Unlock()

	message := Message{generateUniqueID(), "Put", args, kvs.logicalClock, 0}
	err := handleTotalOrderedMulticast(message)
	if err != nil {
		return err
	}

	kvs.mutexClock.Lock()
	kvs.dataStore[args.key] = args.value
	kvs.mutexClock.Unlock()

	response.reply = "true"
	return nil
}

// Delete elimina una coppia chiave-valore, è un operazione di scrittura
func (kvs *KeyValueStoreSequential) Delete(args Args, response *Response) error {

	kvs.mutexClock.Lock()
	kvs.logicalClock++
	kvs.mutexClock.Unlock()

	message := Message{generateUniqueID(), "Delete", args, kvs.logicalClock, 0}
	err := handleTotalOrderedMulticast(message)
	if err != nil {
		return err
	}

	kvs.mutexClock.Lock()
	delete(kvs.dataStore, args.key)
	kvs.mutexClock.Unlock()

	response.reply = "true"
	return nil
}

// handleTotalOrderedMulticast invia la richiesta a tutte le repliche del sistema
func handleTotalOrderedMulticast(args Message) error {
	for i := 0; i < Replicas; i++ {
		// Connessione al server RPC
		client, err := rpc.Dial("tcp", ":"+ReplicaPorts[i])
		if err != nil {
			fmt.Println("Errore durante la connessione al server:", err)
			return nil
		}

		reply := false
		// Chiama il metodo Multiply sul server RPC
		err = client.Call("TotalOrderedMulticast.MulticastTotalOrdered", args, &reply)
		if err != nil {
			fmt.Println("Errore durante la chiamata RPC:", err)
			return nil
		}
		if reply {
			return nil
		}
	}
	return fmt.Errorf("no replica ack")
}

// Genera un ID univoco utilizzando UUID
func generateUniqueID() string {
	id := uuid.New()
	return id.String()
}
