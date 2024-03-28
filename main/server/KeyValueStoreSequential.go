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

// Get restituisce il valore associato alla chiave specificata -> è un evento interno, di lettura
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

	var reply *bool
	// CREO IL MESSAGGIO E DEVO FAR SI CHE TUTTI LO SCRIVONO NEL DATASTORE
	err := sendToOtherServerB("MulticastTotalOrdered.SendToEveryone", message, reply)
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
	err := sendToOtherServerB("MulticastTotalOrdered.SendToEveryone", message, reply)
	if err != nil {
		return err
	}

	response.Reply = "true"
	return nil
}

func (kvs *KeyValueStoreSequential) RealFunction(args Message, response *common.Response) error {
	// Stampa la mappa
	fmt.Println("MAPPA:")
	fmt.Println(kvs.dataStore)

	if args.TypeOfMessage == "Put" { // Scrittura
		kvs.mutexClock.Lock()
		kvs.dataStore[args.Args.Key] = args.Args.Value
		kvs.mutexClock.Unlock()
	} else if args.TypeOfMessage == "Delete" { // Scrittura
		kvs.mutexClock.Lock()
		delete(kvs.dataStore, args.Args.Key)
		kvs.mutexClock.Unlock()
	} else {
		response.Reply = "false"
		return fmt.Errorf("not found")
	}
	response.Reply = "true"
	return nil

	/*else if args.TypeOfMessage == "Get" { // Lettura non dovrebbe fare nulla !!!
		kvs.mutexClock.Lock()
		response.Reply = kvs.dataStore[args.Args.Key]
		kvs.mutexClock.Unlock()
	} */
}

// sendToOtherServer invia a tutti i server la richiesta rpcName
func sendToOtherServer(rpcName string, message Message, response *common.Response) error {

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
				fmt.Println("KeyValueStoreSequential-sendToOtherServer: Errore durante la chiamata RPC "+rpcName+": ", err)
				return
			}

		}(common.ReplicaPorts[i])
	}

	return nil
}

// sendToOtherServer invia a tutti i server la richiesta rpcName
func sendToOtherServerB(rpcName string, message Message, response *bool) error {

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
				fmt.Println("KeyValueStoreSequential-sendToOtherServer: Errore durante la chiamata RPC "+rpcName+": ", err)
				return
			}

		}(common.ReplicaPorts[i])
	}

	return nil
}
