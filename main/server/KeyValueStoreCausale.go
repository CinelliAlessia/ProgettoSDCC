package main

import (
	"fmt"
	"main/common"
)

// Get restituisce il valore associato alla chiave specificata -> Lettura -> Evento interno
func (kvc *KeyValueStoreCausale) Get(args common.Args, response *common.Response) error {

	message := kvc.createMessage(args, "Get")

	err := sendToAllServer("KeyValueStoreCausale.CausallyOrderedMulticast", message, response)
	if err != nil {
		response.Result = false
		return err
	}
	return nil
}

// Put inserisce una nuova coppia chiave-valore, se la chiave è già presente, sovrascrive il valore associato
func (kvc *KeyValueStoreCausale) Put(args common.Args, response *common.Response) error {

	message := kvc.createMessage(args, "Put")

	err := sendToAllServer("KeyValueStoreCausale.CausallyOrderedMulticast", message, response)
	if err != nil {
		response.Result = false
		return err
	}
	return nil
}

// Delete elimina una coppia chiave-valore, è un operazione di scrittura
func (kvc *KeyValueStoreCausale) Delete(args common.Args, response *common.Response) error {

	message := kvc.createMessage(args, "Delete")

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
			printRed("NON ESEGUITO", message, kvc)
			response.Result = false
			return nil
		}
		response.Value = val
		message.Args.Value = val //Fatto solo per DEBUG per il print
	} else {
		return fmt.Errorf("command not found")
	}

	printGreen("ESEGUITO", message, kvc)
	response.Result = true
	return nil
}

func (kvc *KeyValueStoreCausale) createMessage(args common.Args, typeFunc string) MessageC {
	kvc.mutexClock.Lock()
	defer kvc.mutexClock.Unlock()

	kvc.VectorClock[kvc.Id]++
	message := MessageC{common.GenerateUniqueID(), kvc.Id, typeFunc, args, kvc.VectorClock}

	printDebugBlue("RICEVUTO da client", message, kvc)
	return message
}
