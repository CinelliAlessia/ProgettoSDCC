// KeyValueStoreSequential.go
package main

import (
	"fmt"
	"main/common"
	"net/rpc"
	"os"
	"strconv"
	"sync"
)

// KeyValueStoreSequential rappresenta il servizio di memorizzazione chiave-valore specializzato nel sequenziale
type KeyValueStoreSequential struct {
	datastore map[string]string // Mappa -> struttura dati che associa chiavi a valori

	logicalClock int // Orologio logico scalare
	mutexClock   sync.Mutex

	queue      []Message
	mutexQueue sync.Mutex // Mutex per proteggere l'accesso concorrente alla coda
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
	fmt.Println("Richiesta GET di", args.Key)

	kvs.mutexClock.Lock()
	kvs.logicalClock++
	kvs.mutexClock.Unlock()

	id := common.GenerateUniqueID()

	message := Message{id, "Get", args, kvs.logicalClock, 0}
	kvs.addToSortQueue(message)

	for {
		if kvs.queue[0].Id == message.Id && kvs.queue[0].LogicalClock == message.LogicalClock {
			// la seconda condizione dell'if credo sia inutile, se l'id è davvero univoco
			// proprio perché è un id associato al messaggio e non una key
			val, ok := kvs.datastore[message.Args.Key]
			if !ok {
				fmt.Println("Key non trovata nel datastore", message.Args.Key)
				fmt.Println(kvs.datastore)
				return fmt.Errorf("KeyValueStoreSequential: key '%s' not found", message.Args.Key)
			}
			kvs.removeMessage(message)
			response.Reply = val
			break
		}
	}

	fmt.Println("Comando GET eseguito correttamente, DATASTORE:")
	fmt.Println(kvs.datastore)

	return nil
}

// Put inserisce una nuova coppia chiave-valore, se la chiave è già presente, sovrascrive il valore associato
func (kvs *KeyValueStoreSequential) Put(args common.Args, response *common.Response) error {
	fmt.Println("Richiesta PUT di", args.Key, args.Value)

	kvs.mutexClock.Lock()
	kvs.logicalClock++
	kvs.mutexClock.Unlock()

	id := common.GenerateUniqueID()

	// CREO IL MESSAGGIO E DEVO FAR SI CHE TUTTI LO SCRIVONO NEL DATASTORE
	message := Message{id, "Put", args, kvs.logicalClock, 0}

	var reply *bool
	err := kvs.sendToOtherServer("KeyValueStoreSequential.MulticastTotalOrdered", message, reply)
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

	// CREO IL MESSAGGIO E DEVO FAR SI CHE TUTTI LO SCRIVONO NEL DATASTORE
	message := Message{common.GenerateUniqueID(), "Delete", args, kvs.logicalClock, 0}

	var reply *bool
	err := kvs.sendToOtherServer("KeyValueStoreSequential.MulticastTotalOrdered", message, reply)
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
func (kvs *KeyValueStoreSequential) sendToOtherServer(rpcName string, message Message, response *bool) error {
	// NON CONTROLLO MAI IL VALORE DI RESPONSE; ESSENDO UNIVOCO PER TUTTE E TRE POTREI ANCHE LEGGERLO IN MANIERA ERRATA

	//var responseValues [common.Replicas]common.Response
	for i := 0; i < common.Replicas; i++ {
		i := i
		go func(replicaPort string) {

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

			fmt.Println("sendToOtherServer: Contatto il server:", serverName)
			conn, err := rpc.Dial("tcp", serverName)
			if err != nil {
				fmt.Printf("sendToOtherServer: Errore durante la connessione al server:"+replicaPort, err)
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
