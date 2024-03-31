// KeyValueStoreSequential.go
package main

import (
	"fmt"
	"main/common"
	"net/rpc"
	"sync"
)

// KeyValueStoreSequential rappresenta il servizio di memorizzazione chiave-valore specializzato nel sequenziale
type KeyValueStoreSequential struct {
	datastore map[string]string // Mappa -> struttura dati che associa chiavi a valori

	logicalClock int // Orologio logico scalare
	mutexClock   sync.Mutex

	queue []Message
	mu    sync.Mutex // Mutex per proteggere l'accesso concorrente alla coda
}

type Message struct {
	Id            string
	TypeOfMessage string
	Args          common.Args
	LogicalClock  int
	NumberAck     int
}

// Get restituisce il valore associato alla chiave specificata -> è un evento interno, di lettura
func (kvs *KeyValueStoreSequential) Get(args common.Args, response *common.Response) error {
	fmt.Println("Richiesta GET")

	kvs.mutexClock.Lock()
	kvs.logicalClock++
	kvs.mutexClock.Unlock()

	id := common.GenerateUniqueID()

	message := Message{id, "Get", args, kvs.logicalClock, 0}
	kvs.addToSortQueue(message)

	for {
		if kvs.queue[0].Id == message.Id && kvs.queue[0].LogicalClock == message.LogicalClock {
			kvs.mutexClock.Lock()
			val, ok := kvs.datastore[message.Args.Key]
			if !ok {
				fmt.Println("Key non trovata nel datastore", message.Args.Key)
				fmt.Println(kvs.datastore)
				return fmt.Errorf("KeyValueStoreSequential: key '%s' not found", id)
			}
			kvs.mutexClock.Unlock()
			kvs.removeMessage(message)
			response.Reply = val
			break
		}
	}
	fmt.Println("DATASTORE:")
	fmt.Println(kvs.datastore)

	return nil
}

// Put inserisce una nuova coppia chiave-valore, se la chiave è già presente, sovrascrive il valore associato
func (kvs *KeyValueStoreSequential) Put(args common.Args, response *common.Response) error {
	fmt.Println("KeyValueStoreSequential: Comando PUT eseguito")

	kvs.mutexClock.Lock()
	kvs.logicalClock++
	kvs.mutexClock.Unlock()

	// CREO IL MESSAGGIO E DEVO FAR SI CHE TUTTI LO SCRIVONO NEL DATASTORE
	message := Message{common.GenerateUniqueID(), "Put", args, kvs.logicalClock, 0}
	var reply *bool
	err := sendToOtherServer("KeyValueStoreSequential.MulticastTotalOrdered", message, reply)
	if err != nil {
		return err
	}

	response.Reply = "true"
	return nil
}

// Delete elimina una coppia chiave-valore, è un operazione di scrittura
func (kvs *KeyValueStoreSequential) Delete(args common.Args, response *common.Response) error {

	kvs.mutexClock.Lock()
	kvs.logicalClock++
	kvs.mutexClock.Unlock()

	message := Message{common.GenerateUniqueID(), "Delete", args, kvs.logicalClock, 0}

	var reply *bool
	// CREO IL MESSAGGIO E DEVO FAR SI CHE TUTTI LO SCRIVONO NEL DATASTORE
	err := sendToOtherServer("KeyValueStoreSequential.MulticastTotalOrdered", message, reply)
	if err != nil {
		return err
	}

	response.Reply = "true"
	return nil
}

// RealFunction esegue l'operazione di put e di delete realmente
func (kvs *KeyValueStoreSequential) RealFunction(message Message, _ *common.Response) error {

	if message.TypeOfMessage == "Put" { // Scrittura
		kvs.mutexClock.Lock()
		fmt.Println("Key inserita ", message.Args.Key)
		kvs.datastore[message.Args.Key] = message.Args.Value
		fmt.Println("DATASTORE:")
		fmt.Println(kvs.datastore)
		kvs.mutexClock.Unlock()

	} else if message.TypeOfMessage == "Delete" { // Scrittura
		kvs.mutexClock.Lock()
		fmt.Println("Key eliminata ", message.Args.Key)
		delete(kvs.datastore, message.Args.Key)
		fmt.Println("DATASTORE:")
		fmt.Println(kvs.datastore)
		kvs.mutexClock.Unlock()

	} else {
		return fmt.Errorf("command not found")
	}

	return nil
}

// sendToOtherServer invia a tutti i server la richiesta rpcName
func sendToOtherServer(rpcName string, message Message, response *bool) error {

	//var responseValues [common.Replicas]common.Response
	for i := 0; i < common.Replicas; i++ {
		go func(replicaPort string) {

			conn, err := rpc.Dial("tcp", ":"+replicaPort)
			if err != nil {
				fmt.Printf("KeyValueStoreSequential: Errore durante la connessione al server "+replicaPort+": ", err)
				return
			}

			// Chiama il metodo "rpcName" sul server
			err = conn.Call(rpcName, message, &response)
			if err != nil {
				fmt.Println("sendToOtherServer: Errore durante la chiamata RPC "+rpcName+": ", err)
				return
			}

		}(common.ReplicaPorts[i])
	}
	return nil
}
