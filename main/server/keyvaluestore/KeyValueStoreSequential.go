package keyvaluestore

import (
	"main/common"
	"main/server/message"
)

// Get gestisce una chiamata RPC di un evento interno, genera un messaggio e gli allega il suo clock scalare.
// Restituisce il valore associato alla chiave specificata, non notifica ad altri server replica l'evento,
// ma l'esecuzione avviene rispettando l'ordine di programma.
func (kvs *KeyValueStoreSequential) Get(args common.Args, response *common.Response) error {

	for !kvs.canReceive(args) {
		// Aspetta di ricevere tutti i messaggi precedenti da parte del client
	}

	// Si crea un messaggio con 3 ack "ricevuti" così che, per inviarlo a livello applicativo si controllerà
	// solamente l'ordinamento del messaggio nella coda.
	message := kvs.createMessage(args, get)

	//Aggiunge alla coda ordinandolo per timestamp, cosi verrà eseguito esclusivamente se è in testa alla coda
	kvs.addToSortQueue(message)

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
	for !kvs.canReceive(args) {
		// Aspetta di ricevere tutti i messaggi precedenti da parte del client
	}

	message := kvs.createMessage(args, put)

	//Aggiunge alla coda ordinandolo per timestamp, cosi da rispettare l'esecuzione esclusivamente se è in testa alla coda
	kvs.addToSortQueue(message)

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
	for !kvs.canReceive(args) {
		// Aspetta di ricevere tutti i messaggi precedenti da parte del client
	}

	message := kvs.createMessage(args, del)

	//Aggiunge alla coda ordinandolo per timestamp, cosi da rispettare l'esecuzione esclusivamente se è in testa alla coda
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
//
// inserendo la risposta adeguata nella struttura common.Response
// Se l'operazione è andata a buon fine, restituisce true, altrimenti restituisce false,
// sarà la risposta che leggerà il client
func (kvs *KeyValueStoreSequential) realFunction(message *commonMsg.MessageS, response *common.Response) error {

	result := true

	if message.GetTypeOfMessage() == put { // Scrittura
		if kvs.isEndKeyMessage(message) {
			kvs.isAllEndKey()
		} else {
			kvs.PutInDatastore(message.GetKey(), message.GetValue())
		}

	} else if message.GetTypeOfMessage() == del { // Scrittura
		delete(kvs.GetDatastore(), message.GetValue()) //TODO: controllare se funziona

	} else if message.GetTypeOfMessage() == get { // Lettura
		val, ok := kvs.GetDatastore()[message.GetKey()]
		if !ok {
			printRed("NON ESEGUITO", *message, nil, kvs)
			if message.GetIdSender() == kvs.GetIdServer() {
				result = false
			}
		} else if message.GetIdSender() == kvs.GetIdServer() {
			// Solo se io sono il sender imposto la risposta per il client

			response.SetValue(val)
			message.SetValue(val) //Fatto solo per DEBUG per il print
		}
	}

	// Stampa di debug
	//kvs.GetQueue()

	if message.GetIdSender() == kvs.GetIdServer() {
		response.SetResult(result)
		response.SetReceptionFIFO(kvs.GetResponseOrderingFIFO())
		kvs.IncreaseResponseOrderingFIFO()
	}

	printGreen("ESEGUITO", *message, nil, kvs)

	return nil
}

// createMessage preso in input gli argomenti della chiamata RPC, crea un messaggio da inviare:
//  1. il messaggio è creato con il clock scalare incrementato di 1
//  2. se il messaggio è di tipo get, il numero di ack è impostato a common.Replicas
//  3. thread-safe con mutexClock
func (kvs *KeyValueStoreSequential) createMessage(args common.Args, typeFunc string) *commonMsg.MessageS {
	kvs.mutexClock.Lock()
	defer kvs.mutexClock.Unlock()

	kvs.SetClock(kvs.GetClock() + 1)

	numberAck := 0
	if typeFunc == get { // se è una get non serve aspettare ack dato che è un evento interno
		numberAck = common.Replicas
	}

	message := commonMsg.NewMessageSeq(kvs.GetIdServer(), typeFunc, args, kvs.GetClock(), numberAck)
	printDebugBlue("RICEVUTO da client", *message, nil, kvs)

	return message
}

/* In canReceive, si vuole realizzare una mappa che aiuti nell'assunzione di una rete FIFO Ordered */
func (kvs *KeyValueStoreSequential) canReceive(args common.Args) bool {
	//fmt.Println("CanReceive: ", args)

	// Se il client non è nella mappa, lo aggiungo e imposto il timestamp di ricezione a zero
	if _, ok := kvs.ClientMaps[args.GetIDClient()]; !ok {
		// Non ho mai ricevuto un messaggio da questo client

		if args.GetSendingFIFO() == 0 { //  Se è il primo messaggio che avrei dovuto ricevere lo prendo
			kvs.NewClientMap(args.GetIDClient())
			kvs.IncreaseRequestTsClient(args)
			return true
		} else {
			//fmt.Println("Ho ricevuto un messaggio da un client che non conosco ma me ne aspetto altri:", "ReceptionFIFO msg client", args.GetSendingFIFO())
		}
	} else { // Avevo già ricevuto messaggi da questo client
		requestTs, err := kvs.GetRequestTsClient(args.GetIDClient())
		if args.GetSendingFIFO() == requestTs && err == nil {
			// Se il messaggio che ricevo dal client è il successivo rispetto all'ultimo ricevuto, lo posso aggiungere alla coda
			kvs.IncreaseRequestTsClient(args)
			return true
		} else {
			//fmt.Println("Ho ricevuto un messaggio da un client ma me ne aspetto altri:", "ReceptionFIFO msg client", args.GetSendingFIFO(), "ts mio", requestTs, "err", err)
		}
	}
	return false
}
