package sequential

import (
	"main/common"
	"main/server/algorithms"
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
	message := kvs.createMessage(args, common.Get)

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

	message := kvs.createMessage(args, common.Put)

	// Invio la richiesta a tutti i server per sincronizzare i datastore
	err := algorithms.SendToAllServer(common.Sequential+".Update", *message, response)
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

	message := kvs.createMessage(args, common.Del)

	// Invio la richiesta a tutti i server per sincronizzare i datastore
	err := algorithms.SendToAllServer(common.Sequential+".Update", *message, response)
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
func (kvs *KeyValueStoreSequential) realFunction(message *commonMsg.MessageS, response *common.Response) error {

	result := true

	if message.GetTypeOfMessage() == common.Put { // Scrittura

		if kvs.isEndKeyMessage(message) { // Se è un messaggio di fine chiave
			//Ignoro il messaggio di fine chiave, non lo inserisco nel datastore
		} else {
			kvs.PutInDatastore(message.GetKey(), message.GetValue())
		}

	} else if message.GetTypeOfMessage() == common.Del { // Scrittura
		kvs.DeleteFromDatastore(message.GetKey())

	} else if message.GetTypeOfMessage() == common.Get { // Lettura

		val, ok := kvs.GetDatastore()[message.GetKey()]
		if !ok {

			printRed("NON ESEGUITO", *message, kvs)
			if message.GetIdSender() == kvs.GetIdServer() { // Controllo superfluo, non avrò mai messaggi get generati da altri server
				result = false
			}

		} else if message.GetIdSender() == kvs.GetIdServer() {
			response.SetValue(val)
			message.SetValue(val) //Fatto solo per DEBUG per il print
		}
	}

	// A prescindere da result, verrà inviata una risposta al client
	if message.GetIdSender() == kvs.GetIdServer() {
		response.SetResult(result)
		response.SetReceptionFIFO(kvs.GetResponseOrderingFIFO(message.GetClientID())) // Setto il numero di risposte inviate al determinato client

		kvs.SetResponseOrderingFIFO(message.GetClientID(), kvs.GetResponseOrderingFIFO(message.GetClientID())+1) // Incremento il numero di risposte inviate al determinato client
	}

	// Stampa di debug
	/* if result && message.GetIdSender() == kvs.GetIdServer() {
		printGreen("ESEGUITO MIO", *message, nil, kvs)
	} else  */
	if result {
		printGreen("ESEGUITO", *message, kvs)
	}

	return nil
}

// createMessage preso in input gli argomenti della chiamata RPC, crea un messaggio da inviare:
//  1. il messaggio è creato con il clock scalare incrementato di 1
//  2. se il messaggio è di tipo get, il numero di ack è impostato a common.Replicas
//  3. thread-safe con mutexClock
func (kvs *KeyValueStoreSequential) createMessage(args common.Args, typeFunc string) *commonMsg.MessageS {
	// Blocco il mutex qui cosi mi assicuro che il clock associato al messaggio sarà corretto
	// e non modificato da nessun altro
	kvs.mutexClock.Lock()
	defer kvs.mutexClock.Unlock()

	kvs.SetClock(kvs.GetClock() + 1)

	numberAck := 0
	if typeFunc == common.Get { // se è una get non serve aspettare ack dato che è un evento interno
		numberAck = common.Replicas
	}

	message := commonMsg.NewMessageSeq(kvs.GetIdServer(), typeFunc, args, kvs.GetClock(), numberAck)
	printDebugBlue("RICEVUTO da client", *message, kvs)

	// Questo mutex mi permette di evitare scheduling tra il lascia passare di canReceive e la creazione del messaggio
	kvs.UnlockMutexMessage(message.GetClientID()) // Mutex chiuso in canReceive

	return message
}

/* In canReceive, si vuole realizzare una mappa che aiuti nell'assunzione di una rete FIFO Ordered */
func (kvs *KeyValueStoreSequential) canReceive(args common.Args) bool {

	// Se il client non è nella mappa, lo aggiungo e imposto il timestamp di ricezione a zero
	if !kvs.ExistClient(args.GetClientID()) {
		// Non ho mai ricevuto un messaggio da questo client
		kvs.NewClientMap(args.GetClientID()) // Inserisco il client nella mappa
	}

	// Il client è sicuramente nella mappa
	requestTs, err := kvs.GetReceiveTsFromClient(args.GetClientID()) // Ottengo il timestamp di ricezione del client

	if args.GetSendingFIFO() == requestTs && // Se il messaggio che ricevo dal client è quello che mi aspetto
		err == nil { // Se non ci sono stati errori

		// Blocco il mutex per evitare che il client possa inviare un nuovo messaggio prima che io abbia finito di creare il precedente
		kvs.LockMutexMessage(args.GetClientID())
		kvs.SetRequestClient(args.GetClientID(), args.GetSendingFIFO()+1) // Incremento il timestamp di ricezione del client
		return true                                                       // è possibile accettare il messaggio
	}
	return false
}
