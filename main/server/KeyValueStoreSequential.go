// KeyValueStoreSequential.go
package main

import (
	"fmt"
	"main/common"
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
	Id            string
	TypeOfMessage string
	Args          common.Args
	LogicalClock  int
	NumberAck     int
}

// Get restituisce il valore associato alla chiave specificata -> Lettura -> Evento interno
func (kvs *KeyValueStoreSequential) Get(args common.Args, response *common.Response) error {

	kvs.mutexClock.Lock()

	kvs.logicalClock++
	val, ok := kvs.dataStore[args.Key]

	kvs.mutexClock.Unlock()

	if !ok {
		return fmt.Errorf("KeyValueStoreSequential: key '%s' not found", args.Key)
	}

	response.Reply = val
	return nil
}

// Put inserisce una nuova coppia chiave-valore, se la chiave è già presente, sovrascrive il valore associato
func (kvs *KeyValueStoreSequential) Put(args common.Args, response *common.Response) error {
	fmt.Println("KeyValueStoreSequential: Comando PUT eseguito")

	kvs.mutexClock.Lock()
	kvs.logicalClock++
	kvs.mutexClock.Unlock()

	message := Message{common.GenerateUniqueID(), "Put", args, kvs.logicalClock, 0}
	//fmt.Println("SERVER: Key " + args.Key)
	//fmt.Println("SERVER: Value " + args.Value)
	fmt.Println("KeyValueStoreSequential: ID associato al messaggio " + message.Id)

	err := kvs.handleTotalOrderedMulticast(message)
	if err != nil {
		return err
	}

	// TODO !!! Se lo faccio qui lo fa solo il server contattato !!!
	kvs.mutexClock.Lock()
	kvs.dataStore[args.Key] = args.Value
	kvs.mutexClock.Unlock()

	response.Reply = "true"
	return nil
}

// Delete elimina una coppia chiave-valore, è un operazione di scrittura
func (kvs *KeyValueStoreSequential) Delete(args common.Args, response *common.Response) error {

	kvs.mutexClock.Lock()
	kvs.logicalClock++
	kvs.mutexClock.Unlock()

	message := Message{common.GenerateUniqueID(), "Delete", args, kvs.logicalClock, 0}
	err := kvs.handleTotalOrderedMulticast(message)
	if err != nil {
		return err
	}

	kvs.mutexClock.Lock()
	delete(kvs.dataStore, args.Key)
	kvs.mutexClock.Unlock()

	response.Reply = "true"
	return nil
}

// handleTotalOrderedMulticast invia la richiesta a tutte le repliche del sistema
func (kvs *KeyValueStoreSequential) handleTotalOrderedMulticast(args Message) error {
	fmt.Println("KeyValueStoreSequential: Inoltro a tutti i server la richiesta ricevuta dal client")

	// Creare un canale per ricevere le risposte dai server RPC
	responseChannel := make(chan bool, common.Replicas)

	for i := 0; i < common.Replicas; i++ {
		go func(replicaPort string) {
			// Connessione al server RPC
			//fmt.Println("KeyValueStoreSequential: Inoltro della richiesta ricevuta al server " + replicaPort)
			server, err := rpc.Dial("tcp", ":"+replicaPort)
			if err != nil {
				//fmt.Println("KeyValueStoreSequential: Errore durante la connessione al server:", err)
				responseChannel <- false // Invia falso al canale se c'è un errore
				return
			}

			reply := false
			// Chiama il metodo SentToEveryOne sul server RPC
			err = server.Call("MulticastTotalOrdered.SendToEveryone", args, &reply)
			if err != nil {
				fmt.Println("KeyValueStoreSequential: Errore durante la chiamata RPC SentToEveryOne:", err)
				responseChannel <- false // Invia falso al canale se c'è un errore
				return
			}
			responseChannel <- reply // Invia la risposta al canale
		}(common.ReplicaPorts[i])
	}

	// Attendere tutte le risposte
	for i := 0; i < common.Replicas; i++ {
		reply := <-responseChannel
		if reply {
			fmt.Println("KeyValueStoreSequential: Risposta ricevuta da un server.")
			// Puoi gestire la risposta qui
			return nil
		} else {
			return fmt.Errorf("KeyValueStoreSequential: No replica ack")
			// Gestisci l'errore qui
		}
	}
	return nil

}
