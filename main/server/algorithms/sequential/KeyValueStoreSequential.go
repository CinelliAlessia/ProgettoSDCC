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
	// Controllo se è possibile processare la richiesta
	// La funzione è bloccante fin quando non potrò eseguirla
	args.ConfigureSafeBool() // Inizializzo il safeBool
	kvs.canReceiveOld(args)

	message := kvs.createMessage(args, common.Get) // Generazione del messaggio

	//go kvs.canHandleOtherRequest() // Controllo se posso gestire altre richieste

	//Aggiunge alla coda ordinandolo per timestamp, cosi verrà eseguito esclusivamente se è in testa alla coda
	kvs.addToSortQueue(message)
	// canExecute è bloccante fin quando è possibile inviare a livello applicativo il messaggio
	kvs.canExecute(message)
	// Posso eseguire il mio messaggio
	kvs.realFunction(message, response)

	go kvs.canHandleOtherResponse() // Controllo se posso gestire altri messaggi
	return nil
}

// Put inserisce una nuova coppia chiave-valore, se la chiave è già presente, sovrascrive il valore associato
func (kvs *KeyValueStoreSequential) Put(args common.Args, response *common.Response) error {
	return kvs.executeRequest(args, response, common.Put)
}

// Delete elimina una coppia chiave-valore, è un operazione di scrittura
func (kvs *KeyValueStoreSequential) Delete(args common.Args, response *common.Response) error {
	return kvs.executeRequest(args, response, common.Del)
}

func (kvs *KeyValueStoreSequential) executeRequest(args common.Args, response *common.Response, operation string) error {
	// Controllo se è possibile processare la richiesta
	// La funzione è bloccante fin quando non potrò eseguirla
	args.ConfigureSafeBool() // Inizializzo il safeBool
	kvs.canReceiveOld(args)

	message := kvs.createMessage(args, operation) // Generazione del messaggio

	//go kvs.canHandleOtherRequest() // Controllo se posso gestire altre richieste

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
func (kvs *KeyValueStoreSequential) realFunction(message *commonMsg.MessageS, response *common.Response) {

	result := true

	if message.GetTypeOfMessage() == common.Put { // Scrittura

		if message.GetKey() != common.EndKey { // Se è un messaggio di fine chiave
			//Ignoro il messaggio di fine chiave, non lo inserisco nel datastore
			kvs.PutInDatastore(message.GetKey(), message.GetValue())
		}

	} else if message.GetTypeOfMessage() == common.Del { // Scrittura
		kvs.DeleteFromDatastore(message.GetKey())

	} else if message.GetTypeOfMessage() == common.Get { // Lettura

		val, ok := kvs.GetDatastore()[message.GetKey()]
		if !ok { // Se la chiave non è presente nel datastore

			printRed("NON ESEGUITO", *message, kvs)
			result = false

		} else if message.GetSenderID() == kvs.GetServerID() {
			response.SetValue(val)
			message.SetValue(val) //Fatto solo per DEBUG per il print
		}
	}
	// A prescindere da result, verrà inviata una risposta al client
	if message.GetSenderID() == kvs.GetServerID() {
		response.SetResult(result)
		// Imposto il numero di risposte inviate al determinato client
		response.SetReceptionFIFO(kvs.GetResponseOrderingFIFO(message.GetClientID()))
	}
	// Stampa di debug
	if result {
		printGreen("ESEGUITO", *message, kvs)
	}
}

// createMessage preso in input gli argomenti della chiamata RPC, crea un messaggio da inviare:
//  1. il messaggio è creato con il clock scalare incrementato di 1
//  2. se il messaggio è di tipo get, il numero di ack è impostato a common.Replicas
//  3. thread-safe con mutexClock
func (kvs *KeyValueStoreSequential) createMessage(args common.Args, typeFunc string) *commonMsg.MessageS {
	// Blocco il mutex qui cosi mi assicuro che il clock associato al messaggio sarà corretto
	// e non modificato da nessun altro
	kvs.LockMutexClock()
	defer kvs.UnlockMutexClock()

	kvs.SetClock(kvs.GetClock() + 1)

	numberAck := 0
	if typeFunc == common.Get { // se è una get non serve aspettare ack dato che è un evento interno
		numberAck = common.Replicas
	}

	message := commonMsg.NewMessageSeq(kvs.GetServerID(), typeFunc, args, kvs.GetClock(), numberAck)
	printDebugBlue("RICEVUTO da client", *message, kvs)

	// Questo mutex mi permette di evitare scheduling tra il lascia passare di canReceive e la creazione del messaggio
	kvs.UnlockMutexMessage(message.GetClientID()) // Mutex chiuso in canReceive

	return message
}

// In canReceive, si vuole realizzare una mappa che aiuti nell'assunzione di una rete FIFO Ordered //
func (kvs *KeyValueStoreSequential) canReceive(args common.Args) {

	idClient := args.GetClientID()

	receiveMessage := make(chan bool, 1)
	// Eseguo la funzione reale solo se la condizione è true
	go func() {
		// Attendo che il canale sia true impostato da canExecute
		args.WaitCondition() // Aspetta che la condizione sia true
		receiveMessage <- true
	}()

	// Se il client non è nella mappa, lo aggiungo e imposto il timestamp di ricezione e di invio a zero
	if !kvs.ExistClient(idClient) {
		// Non ho mai ricevuto un messaggio da questo client
		kvs.NewClientMap(idClient) // Inserisco il client nella mappa
	}

	// Il client è sicuramente nella mappa
	requestTs := kvs.GetReceiveTsFromClient(idClient) // Ottengo il numero di messaggi ricevuti da questo client

	if args.GetSendingFIFO() == requestTs { // Se il messaggio che ricevo dal client è quello che mi aspetto
		// Blocco il mutex per evitare che il client possa inviare un nuovo messaggio prima che io abbia
		// finito di creare il precedente
		kvs.LockMutexMessage(idClient)
		kvs.SetReceiveTsFromClient(idClient, args.GetSendingFIFO()+1) // Incremento il timestamp di ricezione del client
		args.SetCondition(true)                                       // Imposto la condizione a true, il messaggio può essere eseguito
	} else {
		kvs.AddBufferedArgs(args) // Aggiungo il messaggio in attesa di essere eseguito
	}
	<-receiveMessage // Attendo che la condizione sia true prima di terminare la funzione
}

// In canReceive, si vuole realizzare una mappa che aiuti nell'assunzione di una rete FIFO Ordered //
func (kvs *KeyValueStoreSequential) canReceiveOld(args common.Args) {

	idClient := args.GetClientID()

	// Se il client non è nella mappa, lo aggiungo e imposto il timestamp di ricezione e di invio a zero
	if !kvs.ExistClient(idClient) {
		// Non ho mai ricevuto un messaggio da questo client
		kvs.NewClientMap(idClient) // Inserisco il client nella mappa
	}

	for {
		// Il client è sicuramente nella mappa
		requestTs := kvs.GetReceiveTsFromClient(idClient) // Ottengo il numero di messaggi ricevuti da questo client
		if args.GetSendingFIFO() == requestTs {           // Se il messaggio che ricevo dal client è quello che mi aspetto
			// Blocco il mutex per evitare che il client possa inviare un nuovo messaggio prima che io abbia
			// finito di creare il precedente
			kvs.LockMutexMessage(idClient)
			kvs.SetReceiveTsFromClient(idClient, args.GetSendingFIFO()+1) // Incremento il timestamp di ricezione del client
			break
		}
	}
}
