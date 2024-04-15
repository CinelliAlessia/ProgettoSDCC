package main

import (
	"fmt"
	"github.com/fatih/color"
	"main/common"
	"net/rpc"
	"sync"
)

// KeyValueStoreCausale rappresenta la struttura di memorizzazione chiave-valore per garantire consistenza causale
type KeyValueStoreCausale struct {
	datastore map[string]string // Mappa -> struttura dati che associa chiavi a valori
	id        int               // Id che identifica il server stesso

	vectorClock [common.Replicas]int // Orologio vettoriale
	mutexClock  sync.Mutex

	queue      []MessageC
	mutexQueue sync.Mutex // Mutex per proteggere l'accesso concorrente alla coda

}

type MessageC struct {
	Id            string // Id del messaggio stesso
	IdSender      int    // IdSender rappresenta l'indice del server che invia il messaggio
	TypeOfMessage string
	Args          common.Args
	VectorClock   [common.Replicas]int // Orologio vettoriale
}

// Get restituisce il valore associato alla chiave specificata -> Lettura -> Evento interno
func (kvc *KeyValueStoreCausale) Get(args common.Args, response *common.Response) error {

	kvc.mutexClock.Lock()

	kvc.vectorClock[kvc.id]++
	message := MessageC{common.GenerateUniqueID(), kvc.id, "Get", args, kvc.vectorClock}
	//fmt.Println(color.BlueString("RICEVUTO da client"), message.TypeOfMessage, message.Args.Key, "msg clock:", message.VectorClock, "my clock:", kvc.vectorClock)
	kvc.printDebugBlue("RICEVUTO da client", message)

	kvc.mutexClock.Unlock()

	err := kvc.sendToAllServer("KeyValueStoreCausale.CausallyOrderedMulticast", message, response)
	if err != nil {
		response.Result = false
		return err
	}
	return nil
}

// Put inserisce una nuova coppia chiave-valore, se la chiave è già presente, sovrascrive il valore associato
func (kvc *KeyValueStoreCausale) Put(args common.Args, response *common.Response) error {

	kvc.mutexClock.Lock()
	kvc.vectorClock[kvc.id]++

	// CREO IL MESSAGGIO E DEVO FAR SI CHE TUTTI LO SCRIVONO NEL DATASTORE
	message := MessageC{common.GenerateUniqueID(), kvc.id, "Put", args, kvc.vectorClock}
	//fmt.Println(color.BlueString("RICEVUTO da client"), message.TypeOfMessage, message.Args.Key+":"+message.Args.Key, "msg clock:", message.VectorClock, "my clock:", kvc.vectorClock)
	kvc.printDebugBlue("RICEVUTO da client", message)

	kvc.mutexClock.Unlock()

	err := kvc.sendToAllServer("KeyValueStoreCausale.CausallyOrderedMulticast", message, response)
	if err != nil {
		response.Result = false
		return err
	}
	return nil
}

// Delete elimina una coppia chiave-valore, è un operazione di scrittura
func (kvc *KeyValueStoreCausale) Delete(args common.Args, response *common.Response) error {

	kvc.mutexClock.Lock()
	kvc.vectorClock[kvc.id]++

	// CREO IL MESSAGGIO E DEVO FAR SI CHE TUTTI LO SCRIVONO NEL DATASTORE
	message := MessageC{common.GenerateUniqueID(), kvc.id, "Delete", args, kvc.vectorClock}
	//fmt.Println(color.BlueString("RICEVUTO da client"), message.TypeOfMessage, message.Args.Key, "msg clock:", message.VectorClock, "my clock:", kvc.vectorClock)
	kvc.printDebugBlue("RICEVUTO da client", message)
	kvc.mutexClock.Unlock()

	err := kvc.sendToAllServer("KeyValueStoreCausale.CausallyOrderedMulticast", message, response)
	if err != nil {
		response.Result = false
		return err
	}
	return nil
}

// RealFunction esegue l'operazione di put e di delete realmente
func (kvc *KeyValueStoreCausale) RealFunction(message MessageC, response *common.Response) error {

	if message.TypeOfMessage == "Put" { // Scrittura
		kvc.datastore[message.Args.Key] = message.Args.Value
		kvc.printGreen("ESEGUITO", message)

	} else if message.TypeOfMessage == "Delete" { // Scrittura
		delete(kvc.datastore, message.Args.Key)
		kvc.printGreen("ESEGUITO", message)

	} else if message.TypeOfMessage == "Get" { // Lettura

		val, ok := kvc.datastore[message.Args.Key]
		if !ok {
			fmt.Println(color.RedString("NON ESEGUITO"), message.TypeOfMessage, message.Args.Key, "datastore:", kvc.datastore, "msg clock:", message.VectorClock, "my clock:", kvc.vectorClock)
			response.Result = false
			return nil
		}

		response.Value = val
		message.Args.Value = val //Fatto solo per DEBUG per la funzione sottostante
		kvc.printGreen("ESEGUITO", message)

	} else {
		return fmt.Errorf("command not found")
	}

	printDatastore(kvc)
	response.Result = true
	return nil
}

// sendToAllServer invia a tutti i server la richiesta rpcName
func (kvc *KeyValueStoreCausale) sendToAllServer(rpcName string, message MessageC, response *common.Response) error {
	// Canale per ricevere i risultati delle chiamate RPC
	resultChan := make(chan error, common.Replicas)

	// Itera su tutte le repliche e avvia le chiamate RPC
	for i := 0; i < common.Replicas; i++ {
		go kvc.callRPC(rpcName, message, response, resultChan, i)
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
// Utilizzata unicamente per la chiamata RPC CausallyOrderedMulticast
func (kvc *KeyValueStoreCausale) callRPC(rpcName string, message MessageC, response *common.Response, resultChan chan<- error, replicaIndex int) {

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
		// Gestione dell'errore durante la chiamata RPC
		resultChan <- fmt.Errorf("errore durante la connessione in KeyValueStoreCausale.callRPC: %s", err)
		return
	}

	// Aggiungi il risultato al canale dei risultati
	resultChan <- nil
}
