package keyvaluestore

import (
	"fmt"
	"main/common"

	"main/server/message"
)

// Get gestisce una chiamata RPC di un evento interno, genera un messaggio e allega il suo clock scalare.
// Restituisce il valore associato alla chiave specificata, non notifica ad altri server replica l'evento,
// ma l'esecuzione avviene rispettando l'ordine di programma.
func (kvs *KeyValueStoreSequential) Get(args common.Args, response *common.Response) error {
	// Incrementa il clock logico e genera il messaggio da inviare a livello applicativo
	// Si crea un messaggio con 3 ack "ricevuti" così che per inviarlo a livello applicativo si controllerà
	// solamente l'ordinamento del messaggio nella coda.

	/*for !kvs.workWhitClient(args) {
		// Aspetta di ricevere il messaggio precedente
	}*/

	message := kvs.createMessage(args, get)
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
	/*for !kvs.workWhitClient(args) {
		// Aspetta di ricevere il messaggio precedente
	}*/

	message := kvs.createMessage(args, put)
	kvs.addToSortQueue(message) //Aggiunge alla coda ordinandolo per timestamp, cosi verrà letto esclusivamente se

	// Invio la richiesta a tutti i server per sincronizzare i datastore
	err := sendToAllServer("KeyValueStoreSequential.TotalOrderedMulticast", *message, response)
	if err != nil {
		response.SetResult(false)
		return err
	}
	return nil
}

// Delete elimina una coppia chiave-valore, è un operazione di scrittura
func (kvs *KeyValueStoreSequential) Delete(args common.Args, response *common.Response) error {
	/*for !kvs.workWhitClient(args) {
		// Aspetta di ricevere il messaggio precedente
	}*/

	message := kvs.createMessage(args, del)
	kvs.addToSortQueue(message)

	// Invio la richiesta a tutti i server per sincronizzare i datastore
	err := sendToAllServer("KeyValueStoreSequential.TotalOrderedMulticast", *message, response)
	if err != nil {
		response.SetResult(false)
		return err
	}
	return nil
}

// realFunction esegue l'operazione di get, put e di delete realmente,
// inserendo la risposta adeguata nella struttura common.Response
// Se l'operazione è andata a buon fine, restituisce true, altrimenti restituisce false,
// sarà la risposta che leggerà il client
func (kvs *KeyValueStoreSequential) realFunction(message *msg.MessageS, response *common.Response) error {
	if message.GetTypeOfMessage() == put { // Scrittura
		if kvs.isEndKeyMessage(message) {
			kvs.isAllEndKey()
			return nil
		}

		kvs.SetDatastore(message.GetKey(), message.GetValue())
		//kvs.Datastore[message.GetKey()] = message.GetValue()

	} else if message.GetTypeOfMessage() == del { // Scrittura
		delete(kvs.GetDatastore(), message.GetValue())

	} else if message.GetTypeOfMessage() == get { // Lettura
		val, ok := kvs.GetDatastore()[message.GetValue()]
		if !ok {
			printRed("NON ESEGUITO", *message, nil, kvs)
			if message.GetIdSender() == kvs.GetIdServer() {
				response.SetResult(false)
			}
			return nil
		}
		if message.GetIdSender() == kvs.GetIdServer() { // Solo se io sono il sender imposto la risposta per il client
			response.SetValue(val)
			message.SetValue(val) //Fatto solo per DEBUG per il print
		}
	}

	printGreen("ESEGUITO", *message, nil, kvs)

	if message.GetIdSender() == kvs.GetIdServer() {
		response.SetResult(true)
	}

	return nil
}

func (kvs *KeyValueStoreSequential) createMessage(args common.Args, typeFunc string) *msg.MessageS {
	kvs.mutexClock.Lock()
	defer kvs.mutexClock.Unlock()

	kvs.SetLogicalClock(kvs.GetClock() + 1)
	//kvs.LogicalClock++

	numberAck := 0
	if typeFunc == get { // se è una get non serve aspettare ack
		numberAck = common.Replicas
	}

	message := msg.NewMessageSeq(kvs.GetIdServer(), typeFunc, args, kvs.GetClock(), numberAck)

	printDebugBlue("RICEVUTO da client", *message, nil, kvs)
	return message
}

/* In workWhitClient, si vuole realizzare una mappa che aiuti nell'assunzione di una rete FIFO Ordered */
func (kvs *KeyValueStoreSequential) workWhitClient(args common.Args) bool {
	fmt.Println("workWhitClient", args, "ts suo", args.GetTimestamp())
	// Se il client non è nella mappa, lo aggiungo e imposto il timestamp di ricezione a zero
	if _, ok := kvs.ClientMaps[args.GetIdClient()]; !ok {
		// Non ho ma ricevuto un messaggio da questo client
		if args.TimestampClient == 0 {
			kvs.ClientMaps[args.GetIdClient()] = &ClientMap{RequestTs: 0, ExecuteTs: 0}
			return true
		}
	} else if args.TimestampClient == kvs.GetRequestTsClient(args.GetIdClient())+1 {
		// Se il messaggio che ricevo dal client è il successivo rispetto all'ultimo ricevuto,
		// lo posso aggiungere alla coda
		kvs.IncreaseRequestTsClient(args)
		return true
	}
	return false
}