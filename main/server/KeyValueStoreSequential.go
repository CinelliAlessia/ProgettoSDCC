// KeyValueStoreSequential.go
package main

import (
	"fmt"
	"github.com/fatih/color"
	"main/common"
	"net/rpc"
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
	//fmt.Println("Richiesta GET di", args.Key)

	// Incrementa il clock logico e genera il messaggio da inviare a livello applicativo
	kvs.mutexClock.Lock()
	kvs.logicalClock++
	message := Message{common.GenerateUniqueID(), "Get", args, kvs.logicalClock, 3}
	kvs.mutexClock.Unlock()

	kvs.addToSortQueue(message) //Aggiunge alla coda ordinandolo per timestamp, cosi verrà letto esclusivamente se

	// Controllo in while se il messaggio può essere inviato a livello applicativo
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
	return nil
}

// Put inserisce una nuova coppia chiave-valore, se la chiave è già presente, sovrascrive il valore associato
func (kvs *KeyValueStoreSequential) Put(args common.Args, response *common.Response) error {
	//fmt.Println("Richiesta PUT di", args.Key, args.Value)

	kvs.mutexClock.Lock()
	kvs.logicalClock++

	// CREO IL MESSAGGIO E DEVO FAR SI CHE TUTTI LO SCRIVONO NEL DATASTORE
	id := common.GenerateUniqueID()
	message := Message{id, "Put", args, kvs.logicalClock, 0}

	kvs.mutexClock.Unlock()

	// Invio la richiesta a tutti i server per sincronizzare i datastore
	err := kvs.sendToAllServer("KeyValueStoreSequential.TotalOrderedMulticast", message)
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

	// Invio la richiesta a tutti i server per sincronizzare i datastore
	err := kvs.sendToAllServer("KeyValueStoreSequential.TotalOrderedMulticast", message)
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
		fmt.Println(color.GreenString("ESEGUITO"), message.TypeOfMessage, message.Args.Key+":"+message.Args.Value, "with logicalClock", message.LogicalClock)

	} else if message.TypeOfMessage == "Delete" { // Scrittura
		delete(kvs.datastore, message.Args.Key)
		fmt.Println(color.GreenString("ESEGUITO"), message.TypeOfMessage, message.Args.Key+":"+message.Args.Value)

	} else if message.TypeOfMessage == "Get" { // Lettura
		val, ok := kvs.datastore[message.Args.Key]
		if !ok {
			fmt.Println("Key non trovata nel datastore", message.Args.Key)
			fmt.Println(kvs.datastore)
			return fmt.Errorf("KeyValueStoreSequential: key '%s' not found", message.Args.Key)
		}
		response.Reply = val
		fmt.Println(color.GreenString("ESEGUITO"), message.TypeOfMessage, message.Args.Key, "with logicalClock", message.LogicalClock)
	} else {
		return fmt.Errorf("command not found")
	}
	return nil
}

// sendToAllServer invia a tutti i server la richiesta rpcName
func (kvs *KeyValueStoreSequential) sendToAllServer(rpcName string, message Message) error {
	// Canale per ricevere i risultati delle chiamate RPC
	resultChan := make(chan error, common.Replicas)

	// Itera su tutte le repliche e avvia le chiamate RPC
	for i := 0; i < common.Replicas; i++ {
		go kvs.callRPC(rpcName, message, resultChan, i)
	}

	// Raccoglie i risultati dalle chiamate RPC
	for i := 0; i < common.Replicas; i++ {
		if err := <-resultChan; err != nil {
			return err
		}
	}
	return nil
}

// callRPC è una funzione ausiliaria per effettuare la chiamata RPC a una replica specifica
func (kvs *KeyValueStoreSequential) callRPC(rpcName string, message Message, resultChan chan<- error, replicaIndex int) {

	serverName := common.GetServerName(common.ReplicaPorts[replicaIndex], replicaIndex)

	//fmt.Println("sendToAllServer: Contatto", serverName)
	conn, err := rpc.Dial("tcp", serverName)
	if err != nil {
		// Gestione dell'errore durante la connessione al server
		resultChan <- fmt.Errorf("errore durante la connessione con %s: %v", serverName, err)
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
