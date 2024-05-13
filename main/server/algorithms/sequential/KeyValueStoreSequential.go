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
	kvs.receiveRequest(args)

	// Si crea un messaggio con 3 ack "ricevuti" così che, per inviarlo a livello applicativo si controllerà
	// solamente l'ordinamento del messaggio nella coda.
	message := kvs.createMessage(args, common.Get)
	message.ConfigureSafeBool()

	//Aggiunge alla coda ordinandolo per timestamp, cosi verrà eseguito esclusivamente se è in testa alla coda
	kvs.addToSortQueue(message)

	executeMessage := make(chan bool, 1)

	// Eseguo la funzione reale solo se la condizione è true
	go func() {
		// Attendo che il canale sia true impostato da canExecute
		message.WaitCondition() // Aspetta che la condizione sia true
		err := kvs.realFunction(message, response)
		if err != nil {
			return
		} // Invio a livello applicativo
		executeMessage <- true
	}()

	// Controllo se è possibile eseguire il messaggio a livello applicativo
	go kvs.canExecute(message)

	<-executeMessage
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
	kvs.receiveRequest(args)
	message := kvs.createMessage(args, operation)

	// Invio la richiesta a tutti i server per sincronizzare i datastore
	err := algorithms.SendToAllServer(common.Sequential+".Update", *message, response)
	if err != nil {
		response.SetResult(false)
		return err
	}
	return nil
}

func (kvs *KeyValueStoreSequential) receiveRequest(args common.Args) {
	/*args.ConfigureSafeBool() // Inizializzo il safeBool
	receiveMessage := make(chan bool, 1)

	// Eseguo la funzione reale solo se la condizione è true
	go func() {
		// Attendo che il canale sia true impostato da canExecute
		args.WaitCondition() // Aspetta che la condizione sia true
		receiveMessage <- true
	}()

	// Controllo se è possibile eseguire il messaggio a livello applicativo
	go kvs.canReceive(args)

	<-receiveMessage*/

	for !kvs.canReceive(args) {
		// Aspetto che il messaggio possa essere eseguito
	}
}

// realFunction esegue l'operazione di get, put e di delete realmente,
// inserendo la risposta adeguata nella struttura common.Response
// Se l'operazione è andata a buon fine, restituisce true, altrimenti restituisce false,
// sarà la risposta che leggerà il client
func (kvs *KeyValueStoreSequential) realFunction(message *commonMsg.MessageS, response *common.Response) error {

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
	// Controllo se è possibile gestire altre risposte
	go kvs.canHandleOtherResponse()
	return nil
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
func (kvs *KeyValueStoreSequential) canReceive(args common.Args) bool {

	// Se il client non è nella mappa, lo aggiungo e imposto il timestamp di ricezione a zero
	if !kvs.ExistClient(args.GetClientID()) {
		// Non ho mai ricevuto un messaggio da questo client
		kvs.NewClientMap(args.GetClientID()) // Inserisco il client nella mappa
	}

	// Il client è sicuramente nella mappa
	requestTs := kvs.GetReceiveTsFromClient(args.GetClientID()) // Ottengo il timestamp di ricezione del client

	if args.GetSendingFIFO() == requestTs { // Se il messaggio che ricevo dal client è quello che mi aspetto

		// Blocco il mutex per evitare che il client possa inviare un nuovo messaggio prima che io abbia finito di creare il precedente
		kvs.LockMutexMessage(args.GetClientID())
		kvs.SetReceiveTsFromClient(args.GetClientID(), args.GetSendingFIFO()+1) // Incremento il timestamp di ricezione del client

		//args.SetCondition(true) // Imposto la condizione a true, il messaggio può essere eseguito
		//go kvs.canHandleOtherRequest()
		return true // è possibile accettare il messaggio
	}
	//kvs.AddBufferedArgs(args) // Aggiungo il messaggio in attesa di essere eseguito
	return false
}
