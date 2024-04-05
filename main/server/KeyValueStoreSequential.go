// KeyValueStoreSequential.go
package main

import (
	"fmt"
	"main/common"
	"net/rpc"
	"os"
	"strconv"
	"sync"
	"time"
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

	id := common.GenerateUniqueID() //Id univoco da assegnare al messaggio

	// Incrementa il clock logico e genera il messaggio da inviare a livello applicativo
	kvs.mutexClock.Lock()
	kvs.logicalClock++
	message := Message{id, "Get", args, kvs.logicalClock, 3}
	kvs.mutexClock.Unlock()

	kvs.addToSortQueue(message) //Aggiunge alla coda ordinandolo per timestamp, cosi verrà letto esclusivamente se
	for {
		canSend := kvs.controlSendToApplication(message)
		if canSend {

			// Invio a livello applicativo
			err := kvs.RealFunction(message, response)
			if err != nil {
				return err
			}

			break
		}

		// Altrimenti, attendi un breve periodo prima di riprovare
		time.Sleep(time.Millisecond * 100)
	}
	/*
		for {
			if kvs.queue[0].Id == message.Id {
				// la seconda condizione dell'if credo sia inutile, se l'id è davvero univoco
				// proprio perché è un id associato al messaggio e non una key
				val, ok := kvs.datastore[message.Args.Key]
				if !ok {
					fmt.Println("Key non trovata nel datastore", message.Args.Key)
					fmt.Println(kvs.datastore)
					return fmt.Errorf("KeyValueStoreSequential: key '%s' not found", message.Args.Key)
				}
				response.Reply = val

				fmt.Println("CONTROLLO GET OK")

				kvs.removeMessageToQueue(message)
				break
			} else {
				fmt.Println("CONTROLLO GET", kvs.queue[0].Id, kvs.queue[0].Args.Key, kvs.queue[0].LogicalClock, "messaggio", message.Id, message.Args.Key, message.Args.Value, message.LogicalClock)
			}
			time.Sleep(time.Millisecond * 500)
		}*/
	return nil
}

// Put inserisce una nuova coppia chiave-valore, se la chiave è già presente, sovrascrive il valore associato
func (kvs *KeyValueStoreSequential) Put(args common.Args, response *common.Response) error {
	fmt.Println("Richiesta PUT di", args.Key, args.Value)

	// CREO IL MESSAGGIO E DEVO FAR SI CHE TUTTI LO SCRIVONO NEL DATASTORE
	kvs.mutexClock.Lock()
	kvs.logicalClock++
	id := common.GenerateUniqueID()
	message := Message{id, "Put", args, kvs.logicalClock, 0}
	kvs.mutexClock.Unlock()

	err := kvs.sendToOtherServer("KeyValueStoreSequential.MulticastTotalOrdered", message)
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

	// CREO IL MESSAGGIO E DEVO FAR SI CHE TUTTI LO SCRIVONO NEL DATASTORE
	id := common.GenerateUniqueID()
	message := Message{id, "Delete", args, kvs.logicalClock, 0}

	kvs.mutexClock.Unlock()

	err := kvs.sendToOtherServer("KeyValueStoreSequential.MulticastTotalOrdered", message)
	if err != nil {
		return err
	}

	response.Reply = "true"
	return nil
}

// RealFunction esegue l'operazione di put e di delete realmente
func (kvs *KeyValueStoreSequential) RealFunction(message Message, response *common.Response) error {

	if message.TypeOfMessage == "Put" { // Scrittura
		kvs.datastore[message.Args.Key] = message.Args.Value
	} else if message.TypeOfMessage == "Delete" { // Scrittura
		delete(kvs.datastore, message.Args.Key)
	} else if message.TypeOfMessage == "Get" {
		val, ok := kvs.datastore[message.Args.Key]
		if !ok {
			fmt.Println("Key non trovata nel datastore", message.Args.Key)
			fmt.Println(kvs.datastore)
			return fmt.Errorf("KeyValueStoreSequential: key '%s' not found", message.Args.Key)
		}
		response.Reply = val
	} else {
		return fmt.Errorf("command not found")
	}

	return nil
}

// sendToOtherServer invia a tutti i server la richiesta rpcName
func (kvs *KeyValueStoreSequential) sendToOtherServer(rpcName string, message Message) error {
	// Canale per ricevere i risultati delle chiamate RPC
	resultChan := make(chan error, common.Replicas)

	// Itera su tutte le repliche e avvia le chiamate RPC
	for i := 0; i < common.Replicas; i++ {
		go kvs.callRPC(rpcName, message, resultChan, i)
	}

	// Raccoglie i risultati dalle chiamate RPC
	for i := 0; i < common.Replicas; i++ {
		if err := <-resultChan; err != nil {
			// Gestisci l'errore in base alle tue esigenze
			return err
		}
	}
	return nil
}

// callRPC è una funzione ausiliaria per effettuare la chiamata RPC a una replica specifica
func (kvs *KeyValueStoreSequential) callRPC(rpcName string, message Message, resultChan chan<- error, replicaIndex int) {
	var serverName string

	if os.Getenv("CONFIG") == "1" {
		/*---LOCALE---*/
		serverName = ":" + common.ReplicaPorts[replicaIndex]
	} else if os.Getenv("CONFIG") == "2" {
		/*---DOCKER---*/
		serverName = "server" + strconv.Itoa(replicaIndex+1) + ":" + common.ReplicaPorts[replicaIndex]
	} else {
		// Gestione dell'errore se la variabile di ambiente non è corretta
		resultChan <- fmt.Errorf("VARIABILE DI AMBIENTE ERRATA")
		return
	}

	fmt.Println("sendToOtherServer: Contatto il server:", serverName)
	conn, err := rpc.Dial("tcp", serverName)
	if err != nil {
		// Gestione dell'errore durante la connessione al server
		resultChan <- fmt.Errorf("errore durante la connessione al server %s: %v", serverName, err)
		return
	}

	// Chiama il metodo "rpcName" sul server
	var response bool
	err = conn.Call(rpcName, message, &response)
	if err != nil {
		// Gestione dell'errore durante la chiamata RPC
		resultChan <- fmt.Errorf("errore durante la chiamata RPC %s a %s: %v", rpcName, serverName, err)
		return
	}

	// Aggiungi il risultato al canale dei risultati
	resultChan <- nil
}
