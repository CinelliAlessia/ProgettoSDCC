// Package sequential KeyValueStoreSequential.go
package main

import (
	"fmt"
	"main/common"
	"sync"
	"time"
)

// KeyValueStoreSequential rappresenta il servizio di memorizzazione chiave-valore specializzato nel sequenziale
type KeyValueStoreSequential struct {
	Datastore map[string]string // Mappa -> struttura dati che associa chiavi a valori
	Id        int               // Id che identifica il server stesso

	LogicalClock int // Orologio logico scalare
	mutexClock   sync.Mutex

	Queue      []MessageS
	mutexQueue sync.Mutex // Mutex per proteggere l'accesso concorrente alla coda

	executeFunctionMutex sync.Mutex
}

type MessageS struct {
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
	kvs.LogicalClock++
	message := MessageS{common.GenerateUniqueID(), kvs.Id, "Get", args, kvs.LogicalClock, 3}
	kvs.printDebugBlue("RICEVUTO da client", message)
	kvs.mutexClock.Unlock()

	// TODO: problema, la get anche con timestamp maggiore prende la precedenza perché le altre richieste non si sono ancora messe in coda
	kvs.addToSortQueue(message) //Aggiunge alla coda ordinandolo per timestamp, cosi verrà letto esclusivamente se

	// Controllo in while se il messaggio può essere inviato a livello applicativo
	for {
		canSend := kvs.controlSendToApplication(message)
		if canSend {
			// Invio a livello applicativo
			err := kvs.realFunction(message, response)
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

	kvs.mutexClock.Lock()
	kvs.LogicalClock++

	// CREO IL MESSAGGIO E DEVO FAR SI CHE TUTTI LO SCRIVONO NEL DATASTORE
	message := MessageS{common.GenerateUniqueID(), kvs.Id, "Put", args, kvs.LogicalClock, 0}
	kvs.printDebugBlue("RICEVUTO da client", message)
	//fmt.Println(color.BlueString("RICEVUTO da client"), message.TypeOfMessage, message.Args.Key+":"+message.Args.Value, "msg clock:", message.LogicalClock, "my clock:", kvs.logicalClock)
	kvs.mutexClock.Unlock()

	//kvs.addToSortQueue(message) //Aggiunge alla coda ordinandolo per timestamp, cosi verrà letto esclusivamente se

	// Invio la richiesta a tutti i server per sincronizzare i datastore
	err := sendToAllServer("KeyValueStoreSequential.TotalOrderedMulticast", message, response)
	if err != nil {
		response.Result = false
		return err
	}
	return nil
}

// Delete elimina una coppia chiave-valore, è un operazione di scrittura
func (kvs *KeyValueStoreSequential) Delete(args common.Args, response *common.Response) error {

	kvs.mutexClock.Lock()
	kvs.LogicalClock++

	// CREO IL MESSAGGIO E DEVO FAR SI CHE TUTTI LO SCRIVONO NEL DATASTORE
	id := common.GenerateUniqueID()
	message := MessageS{id, kvs.Id, "Delete", args, kvs.LogicalClock, 0}
	kvs.printDebugBlue("RICEVUTO da client", message)
	//fmt.Println(color.BlueString("RICEVUTO da client"), message.TypeOfMessage, message.Args.Key+":"+message.Args.Value, "msg clock:", message.LogicalClock, "my clock:", kvs.logicalClock)
	kvs.mutexClock.Unlock()

	//kvs.addToSortQueue(message) //Aggiunge alla coda ordinandolo per timestamp, cosi verrà letto esclusivamente se

	// Invio la richiesta a tutti i server per sincronizzare i datastore
	err := sendToAllServer("KeyValueStoreSequential.TotalOrderedMulticast", message, response)
	if err != nil {
		response.Result = false
		return err
	}
	return nil
}

// RealFunction esegue l'operazione di get, put e di delete realmente, inserendo la risposta adeguata nella struttura common.Response
func (kvs *KeyValueStoreSequential) realFunction(message MessageS, response *common.Response) error {
	kvs.executeFunctionMutex.Lock()
	defer kvs.executeFunctionMutex.Unlock()

	if message.TypeOfMessage == "Put" { // Scrittura

		kvs.Datastore[message.Args.Key] = message.Args.Value
		kvs.printGreen("ESEGUITO", message)

	} else if message.TypeOfMessage == "Delete" { // Scrittura

		delete(kvs.Datastore, message.Args.Key)
		kvs.printGreen("ESEGUITO", message)

	} else if message.TypeOfMessage == "Get" { // Lettura

		val, ok := kvs.Datastore[message.Args.Key]
		if !ok {
			kvs.printRed("NON ESEGUITO", message)
			response.Result = false
			return nil
		}
		response.Value = val

		message.Args.Value = val //Fatto solo per DEBUG per la funzione sottostante
		kvs.printGreen("ESEGUITO", message)
	} else {
		return fmt.Errorf("command not found")
	}

	response.Result = true
	return nil
}
