package main

import (
	"fmt"
	"main/common"
	"sync"
)

// KeyValueStoreCausale rappresenta la struttura di memorizzazione chiave-valore per garantire consistenza causale
type KeyValueStoreCausale struct {
	Datastore map[string]string // Mappa -> struttura dati che associa chiavi a valori
	Id        int               // Id che identifica il server stesso

	VectorClock [common.Replicas]int // Orologio vettoriale
	mutexClock  sync.Mutex

	Queue      []MessageC
	mutexQueue sync.Mutex // Mutex per proteggere l'accesso concorrente alla coda

	executeFunctionMutex sync.Mutex // Mutex aggiunto per evitare scheduling che interrompano l'invio a livello applicativo del messaggio
}

type MessageC struct {
	Id       string // Id del messaggio stesso
	IdSender int    // IdSender rappresenta l'indice del server che invia il messaggio

	TypeOfMessage string
	Args          common.Args

	VectorClock [common.Replicas]int // Orologio vettoriale
}

// Get restituisce il valore associato alla chiave specificata -> Lettura -> Evento interno
func (kvc *KeyValueStoreCausale) Get(args common.Args, response *common.Response) error {

	kvc.mutexClock.Lock()

	kvc.VectorClock[kvc.Id]++
	message := MessageC{common.GenerateUniqueID(), kvc.Id, "Get", args, kvc.VectorClock}
	kvc.printDebugBlue("RICEVUTO da client", message)

	kvc.mutexClock.Unlock()

	err := sendToAllServer("KeyValueStoreCausale.CausallyOrderedMulticast", message, response)
	if err != nil {
		response.Result = false
		return err
	}
	return nil
}

// Put inserisce una nuova coppia chiave-valore, se la chiave è già presente, sovrascrive il valore associato
func (kvc *KeyValueStoreCausale) Put(args common.Args, response *common.Response) error {

	kvc.mutexClock.Lock()

	kvc.VectorClock[kvc.Id]++
	message := MessageC{common.GenerateUniqueID(), kvc.Id, "Put", args, kvc.VectorClock}
	kvc.printDebugBlue("RICEVUTO da client", message)

	kvc.mutexClock.Unlock()

	err := sendToAllServer("KeyValueStoreCausale.CausallyOrderedMulticast", message, response)
	if err != nil {
		response.Result = false
		return err
	}
	return nil
}

// Delete elimina una coppia chiave-valore, è un operazione di scrittura
func (kvc *KeyValueStoreCausale) Delete(args common.Args, response *common.Response) error {

	kvc.mutexClock.Lock()

	kvc.VectorClock[kvc.Id]++
	message := MessageC{common.GenerateUniqueID(), kvc.Id, "Delete", args, kvc.VectorClock}
	kvc.printDebugBlue("RICEVUTO da client", message)

	kvc.mutexClock.Unlock()

	err := sendToAllServer("KeyValueStoreCausale.CausallyOrderedMulticast", message, response)
	if err != nil {
		response.Result = false
		return err
	}
	return nil
}

// RealFunction esegue l'operazione di put e di delete realmente
func (kvc *KeyValueStoreCausale) realFunction(message MessageC, response *common.Response) error {
	if message.TypeOfMessage == "Put" { // Scrittura
		kvc.Datastore[message.Args.Key] = message.Args.Value

	} else if message.TypeOfMessage == "Delete" { // Scrittura
		delete(kvc.Datastore, message.Args.Key)

	} else if message.TypeOfMessage == "Get" { // Lettura

		val, ok := kvc.Datastore[message.Args.Key]
		if !ok {
			kvc.printRed("NON ESEGUITO", message)
			response.Result = false
			return nil
		}
		response.Value = val
		message.Args.Value = val //Fatto solo per DEBUG per il print
	} else {
		return fmt.Errorf("command not found")
	}

	kvc.printGreen("ESEGUITO", message)
	response.Result = true
	return nil
}
