package main

import (
	"fmt"
	"main/common"
	"net/rpc"
	"os"
	"strconv"
	"sync"
)

// KeyValueStoreCausale rappresenta il servizio di memorizzazione chiave-valore
type KeyValueStoreCausale struct {
	datastore map[string]string // Mappa -> struttura dati che associa chiavi a valori
	id        int

	vectorClock [common.Replicas]int // Orologio vettoriale
	mutexClock  sync.Mutex

	queue      []MessageC
	mutexQueue sync.Mutex // Mutex per proteggere l'accesso concorrente alla coda
}

type MessageC struct {
	Id            string
	TypeOfMessage string
	Args          common.Args
	VectorClock   [common.Replicas]int // Orologio vettoriale
	NumberAck     int
}

// Get restituisce il valore associato alla chiave specificata -> Lettura -> Evento interno
func (kvc *KeyValueStoreCausale) Get(args common.Args, response *common.Response) error {
	fmt.Println("Richiesta GET")

	kvc.mutexClock.Lock()
	kvc.vectorClock[kvc.id]++
	kvc.mutexClock.Unlock()

	id := common.GenerateUniqueID()
	id = "b"

	message := MessageC{id, "Get", args, kvc.vectorClock, 0}
	kvc.addToSortQueue(message)

	for {
		if kvc.queue[0].Id == message.Id /*&& sort come si deve*/ {
			kvc.mutexClock.Lock()
			val, ok := kvc.datastore[message.Args.Key]
			if !ok {
				fmt.Println("Key non trovata nel datastore", message.Args.Key)
				fmt.Println(kvc.datastore)
				return fmt.Errorf("KeyValueStoreCausale: key '%s' not found", id)
			}
			kvc.removeMessage(message)
			kvc.mutexClock.Unlock()
			response.Reply = val
			break
		}
	}
	fmt.Println("DATASTORE:")
	fmt.Println(kvc.datastore)
	return nil
}

// Put inserisce una nuova coppia chiave-valore, se la chiave è già presente, sovrascrive il valore associato
func (kvc *KeyValueStoreCausale) Put(args common.Args, response *common.Response) error {
	fmt.Println("KeyValueStoreCausale: Comando PUT eseguito")

	kvc.mutexClock.Lock()
	kvc.vectorClock[kvc.id]++
	kvc.mutexClock.Unlock()

	id := common.GenerateUniqueID()
	id = "a"

	// CREO IL MESSAGGIO E DEVO FAR SI CHE TUTTI LO SCRIVONO NEL DATASTORE
	message := MessageC{id, "Put", args, kvc.vectorClock, 0}
	var reply *bool
	err := kvc.sendToOtherServer("KeyValueStoreCausale.CausallyOrderedMulticast", message, reply)
	if err != nil {
		return err
	}

	response.Reply = "true"
	return nil
}

// Delete elimina una coppia chiave-valore, è un operazione di scrittura
func (kvc *KeyValueStoreCausale) Delete(args common.Args, response *common.Response) error {
	fmt.Println("KeyValueStoreCausale: Comando PUT eseguito")

	kvc.mutexClock.Lock()
	kvc.vectorClock[kvc.id]++
	kvc.mutexClock.Unlock()

	id := common.GenerateUniqueID()
	id = "a"

	// CREO IL MESSAGGIO E DEVO FAR SI CHE TUTTI LO SCRIVONO NEL DATASTORE
	message := MessageC{id, "Delete", args, kvc.vectorClock, 0}
	var reply *bool
	err := kvc.sendToOtherServer("KeyValueStoreCausale.CausallyOrderedMulticast", message, reply)
	if err != nil {
		return err
	}

	response.Reply = "true"
	return nil
}

// sendToOtherServer invia a tutti i server la richiesta rpcName
func (kvc *KeyValueStoreCausale) sendToOtherServer(rpcName string, message MessageC, response *bool) error {

	//var responseValues [common.Replicas]common.Response
	for i := 0; i < common.Replicas; i++ {
		i := i
		go func(replicaPort string) {
			// Connessione al server RPC casuale
			var conn *rpc.Client
			var err error

			if os.Getenv("CONFIG") == "1" {
				/*---LOCALE---*/
				// Connessione al server RPC casuale
				fmt.Println("sendToOtherServer: Contatto il server:", ":"+replicaPort)
				conn, err = rpc.Dial("tcp", ":"+replicaPort)

			} else {
				/*---DOCKER---*/
				// Connessione al server RPC casuale
				fmt.Println("sendToOtherServer: Contatto il server:", "server"+strconv.Itoa(i+1)+":"+replicaPort)
				conn, err = rpc.Dial("tcp", "server"+strconv.Itoa(i+1)+":"+replicaPort)
			}
			if err != nil {
				fmt.Printf("KeyValueStoreCausale: Errore durante la connessione al server:"+replicaPort, err)
				return
			}

			// Chiama il metodo "rpcName" sul server
			err = conn.Call(rpcName, message, &response)
			if err != nil {
				fmt.Println("KeyValueStoreCausale: Errore durante la chiamata RPC "+rpcName+": ", err)
				return
			}

		}(common.ReplicaPorts[i])
	}
	return nil
}

// RealFunction esegue l'operazione di put e di delete realmente
func (kvc *KeyValueStoreCausale) RealFunction(message MessageC, _ *common.Response) error {

	if message.TypeOfMessage == "Put" { // Scrittura
		kvc.mutexClock.Lock()
		fmt.Println("Key inserita ", message.Args.Key)
		kvc.datastore[message.Args.Key] = message.Args.Value
		fmt.Println("DATASTORE:")
		fmt.Println(kvc.datastore)
		kvc.mutexClock.Unlock()

	} else if message.TypeOfMessage == "Delete" { // Scrittura
		kvc.mutexClock.Lock()
		fmt.Println("Key eliminata ", message.Args.Key)
		delete(kvc.datastore, message.Args.Key)
		fmt.Println("DATASTORE:")
		fmt.Println(kvc.datastore)
		kvc.mutexClock.Unlock()

	} else {
		return fmt.Errorf("command not found")
	}

	return nil
}
