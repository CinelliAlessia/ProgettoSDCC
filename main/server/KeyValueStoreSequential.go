package main

import (
	"main/common"
)

// Get gestisce una chiamata RPC di un evento interno, genera un messaggio e allega il suo clock scalare.
// Restituisce il valore associato alla chiave specificata, non notifica ad altri server replica l'evento,
// ma l'esecuzione avviene rispettando l'ordine di programma.
func (kvs *KeyValueStoreSequential) Get(args common.Args, response *common.Response) error {
	// Incrementa il clock logico e genera il messaggio da inviare a livello applicativo
	// Si crea un messaggio con 3 ack "ricevuti" così che per inviarlo a livello applicativo si controllerà
	// solamente l'ordinamento del messaggio nella coda.

	message := kvs.createMessage(args, "Get")
	kvs.addToSortQueue(message) //Aggiunge alla coda ordinandolo per timestamp, cosi verrà letto esclusivamente se

	// Controllo in while se il messaggio può essere inviato a livello applicativo
	for {
		stop, err := kvs.canExecute(message, response)
		if stop {
			break
		}
		if err != nil {
			return err
		}
	}
	return nil
}

// Put inserisce una nuova coppia chiave-valore, se la chiave è già presente, sovrascrive il valore associato
func (kvs *KeyValueStoreSequential) Put(args common.Args, response *common.Response) error {

	message := kvs.createMessage(args, "Put")
	kvs.addToSortQueue(message) //Aggiunge alla coda ordinandolo per timestamp, cosi verrà letto esclusivamente se

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

	message := kvs.createMessage(args, "Delete")
	kvs.addToSortQueue(message)

	// Invio la richiesta a tutti i server per sincronizzare i datastore
	err := sendToAllServer("KeyValueStoreSequential.TotalOrderedMulticast", message, response)
	if err != nil {
		response.Result = false
		return err
	}
	return nil
}

// realFunction esegue l'operazione di get, put e di delete realmente,
// inserendo la risposta adeguata nella struttura common.Response
// Se l'operazione è andata a buon fine, restituisce true, altrimenti restituisce false,
// sarà la risposta che leggerà il client
func (kvs *KeyValueStoreSequential) realFunction(message MessageS, response *common.Response) error {
	if message.TypeOfMessage == "Put" { // Scrittura
		if kvs.isEndKeyMessage(message) {
			kvs.isAllEndKey()
			return nil
		}

		kvs.Datastore[message.Args.Key] = message.GetValue()

	} else if message.TypeOfMessage == "Delete" { // Scrittura
		delete(kvs.Datastore, message.GetValue())

	} else if message.TypeOfMessage == "Get" { // Lettura
		val, ok := kvs.Datastore[message.GetValue()]
		if !ok {
			printRed("NON ESEGUITO", message, kvs)
			if message.GetIdSender() == kvs.Id {
				response.Result = false
			}
			return nil
		}
		if message.IdSender == kvs.Id { // Solo se io sono il sender imposto la risposta per il client
			response.Value = val
			message.Args.Value = val //Fatto solo per DEBUG per il print
		}
	}

	printGreen("ESEGUITO", message, kvs)

	if message.IdSender == kvs.Id {
		response.Result = true
	}

	return nil
}

func (kvs *KeyValueStoreSequential) createMessage(args common.Args, typeFunc string) MessageS {
	kvs.mutexClock.Lock()
	defer kvs.mutexClock.Unlock()

	kvs.LogicalClock++
	numberAck := 0
	if typeFunc == "Get" { // se è una get non serve aspettare ack
		numberAck = common.Replicas
	}

	message := MessageS{common.GenerateUniqueID(), kvs.Id, typeFunc, args, kvs.LogicalClock, numberAck}

	printDebugBlue("RICEVUTO da client", message, kvs)
	return message
}
