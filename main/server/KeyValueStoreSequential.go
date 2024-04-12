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
	id        int               // Id che identifica il server stesso

	logicalClock int // Orologio logico scalare
	mutexClock   sync.Mutex

	queue      []Message
	mutexQueue sync.Mutex // Mutex per proteggere l'accesso concorrente alla coda
}

type Message struct {
	Id            string
	IdSender      int // IdSender rappresenta l'indice del server che invia il messaggio
	TypeOfMessage string
	Args          common.Args
	LogicalClock  int
	NumberAck     int
}

// Get gestisce una chiamata RPC di un evento interno, genera un messaggio e allega il suo clock scalare.
// Restituisce il valore associato alla chiave specificata, non notifica ad altri server replica l'evento,
// ma l'esecuzione avviene rispettando l'ordine di programma.
func (kvs *KeyValueStoreSequential) Get(args common.Args, response *common.Response) error {
	// Incrementa il clock logico e genera il messaggio da inviare a livello applicativo
	// Si crea un messaggio con 3 ack "ricevuti" così che per inviarlo a livello applicativo si controllerà
	// solamente l'ordinamento del messaggio nella coda.
	kvs.mutexClock.Lock()
	kvs.logicalClock++
	message := Message{common.GenerateUniqueID(), kvs.id, "Get", args, kvs.logicalClock, 3}
	fmt.Println(color.BlueString("RICEVUTO da client"), message.TypeOfMessage, message.Args.Key, "msg clock:", message.LogicalClock, "my clock:", kvs.logicalClock)
	kvs.mutexClock.Unlock()

	// TODO: problema, la get anche con timestamp maggiore prende la precedenza perché le altre richieste non si sono ancora messe in coda
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
	message := Message{common.GenerateUniqueID(), kvs.id, "Put", args, kvs.logicalClock, 0}
	fmt.Println(color.BlueString("RICEVUTO da client"), message.TypeOfMessage, message.Args.Key+":"+message.Args.Value, "msg clock:", message.LogicalClock, "my clock:", kvs.logicalClock)
	kvs.mutexClock.Unlock()

	kvs.addToSortQueue(message) //Aggiunge alla coda ordinandolo per timestamp, cosi verrà letto esclusivamente se

	// Invio la richiesta a tutti i server per sincronizzare i datastore
	err := kvs.sendToAllServer("KeyValueStoreSequential.TotalOrderedMulticast", message, response)
	if err != nil {
		response.Result = false
		return err
	}
	return nil
}

// Delete elimina una coppia chiave-valore, è un operazione di scrittura
func (kvs *KeyValueStoreSequential) Delete(args common.Args, response *common.Response) error {

	kvs.mutexClock.Lock()
	kvs.logicalClock++

	// CREO IL MESSAGGIO E DEVO FAR SI CHE TUTTI LO SCRIVONO NEL DATASTORE
	id := common.GenerateUniqueID()
	message := Message{id, kvs.id, "Delete", args, kvs.logicalClock, 0}
	fmt.Println(color.BlueString("RICEVUTO da client"), message.TypeOfMessage, message.Args.Key+":"+message.Args.Value, "msg clock:", message.LogicalClock, "my clock:", kvs.logicalClock)
	kvs.mutexClock.Unlock()

	kvs.addToSortQueue(message) //Aggiunge alla coda ordinandolo per timestamp, cosi verrà letto esclusivamente se

	// Invio la richiesta a tutti i server per sincronizzare i datastore
	err := kvs.sendToAllServer("KeyValueStoreSequential.TotalOrderedMulticast", message, response)
	if err != nil {
		response.Result = false
		return err
	}
	return nil
}

// RealFunction esegue l'operazione di get, put e di delete realmente, inserendo la risposta adeguata nella struttura common.Response
func (kvs *KeyValueStoreSequential) RealFunction(message Message, response *common.Response) error {

	if message.TypeOfMessage == "Put" { // Scrittura

		kvs.datastore[message.Args.Key] = message.Args.Value
		fmt.Println(color.GreenString("ESEGUITO"), message.TypeOfMessage, message.Args.Key+":"+message.Args.Value, "msg clock:", message.LogicalClock, "my clock", kvs.logicalClock)
		printDatastore(kvs)
	} else if message.TypeOfMessage == "Delete" { // Scrittura

		delete(kvs.datastore, message.Args.Key)
		fmt.Println(color.GreenString("ESEGUITO"), message.TypeOfMessage, message.Args.Key+":"+message.Args.Value, "msg clock:", message.LogicalClock, "my clock:", kvs.logicalClock)
		printDatastore(kvs)
	} else if message.TypeOfMessage == "Get" { // Lettura

		val, ok := kvs.datastore[message.Args.Key]
		if !ok {
			fmt.Println(color.RedString("NON ESEGUITO"), message.TypeOfMessage, message.Args.Key, "datastore:", kvs.datastore, "msg clock:", message.LogicalClock, "my clock:", kvs.logicalClock)
			response.Result = false
			return nil
		}
		response.Value = val
		fmt.Println(color.GreenString("ESEGUITO"), message.TypeOfMessage, message.Args.Key+":"+val, "msg clock:", message.LogicalClock, "my clock:", kvs.logicalClock)
		printDatastore(kvs)
	} else {
		return fmt.Errorf("command not found")
	}

	response.Result = true
	return nil
}

// sendToAllServer invia a tutti i server la richiesta rpcName
func (kvs *KeyValueStoreSequential) sendToAllServer(rpcName string, message Message, response *common.Response) error {
	// Canale per ricevere i risultati delle chiamate RPC
	resultChan := make(chan error, common.Replicas)

	// Itera su tutte le repliche e avvia le chiamate RPC
	for i := 0; i < common.Replicas; i++ {
		go kvs.callRPC(rpcName, message, response, resultChan, i)
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
func (kvs *KeyValueStoreSequential) callRPC(rpcName string, message Message, response *common.Response, resultChan chan<- error, replicaIndex int) {

	serverName := common.GetServerName(common.ReplicaPorts[replicaIndex], replicaIndex)

	//fmt.Println("sendToAllServer: Contatto", serverName)
	conn, err := rpc.Dial("tcp", serverName)
	if err != nil {
		// Gestione dell'errore durante la connessione al server
		resultChan <- fmt.Errorf("errore durante la connessione con %s: %v", serverName, err)
		return
	}

	// Chiama il metodo "rpcName" sul server
	common.RandomDelay()
	err = conn.Call(rpcName, message, response)
	if err != nil {
		// Gestione dell'errore durante la chiamata RPC
		resultChan <- fmt.Errorf("errore durante la chiamata RPC %s a %s: %v", rpcName, serverName, err)
		return
	}

	err = conn.Close()
	if err != nil {
		resultChan <- fmt.Errorf("errore durante la chiusura della connessione in KeyValueStoreSequential.callRPC: %s", err)
		return
	}

	// Aggiungi il risultato al canale dei risultati
	resultChan <- nil
}
