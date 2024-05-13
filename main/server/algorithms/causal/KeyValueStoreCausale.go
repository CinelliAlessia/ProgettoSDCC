package causal

import (
	"main/common"
	"main/server/algorithms"
	"main/server/message"
)

// Get restituisce il valore associato alla chiave specificata -> Lettura -> Evento interno
func (kvc *KeyValueStoreCausale) Get(args common.Args, response *common.Response) error {
	return kvc.executeRequest(args, response, common.Get)
}

// Put inserisce una nuova coppia chiave-valore, se la chiave è già presente, sovrascrive il valore associato
func (kvc *KeyValueStoreCausale) Put(args common.Args, response *common.Response) error {
	return kvc.executeRequest(args, response, common.Put)
}

// Delete elimina una coppia chiave-valore, è un operazione di scrittura
func (kvc *KeyValueStoreCausale) Delete(args common.Args, response *common.Response) error {
	return kvc.executeRequest(args, response, common.Del)
}

func (kvc *KeyValueStoreCausale) executeRequest(args common.Args, response *common.Response, operation string) error {

	// Controllo se è possibile processare la richiesta
	// La funzione è bloccante fin quando non potrò eseguirla
	args.ConfigureSafeBool() // Inizializzo il safeBool
	kvc.canReceive(args)

	message := kvc.createMessage(args, operation) // Generazione del messaggio

	go kvc.canHandleOtherRequest() // Controllo se posso gestire altre richieste

	err := algorithms.SendToAllServer(common.Causal+".Update", *message, response)
	if err != nil {
		response.SetResult(false)
		return err
	}
	return nil
}

// RealFunction esegue l'operazione di put e di delete realmente
func (kvc *KeyValueStoreCausale) realFunction(message *commonMsg.MessageC, response *common.Response) {

	result := true

	if message.GetTypeOfMessage() == common.Put { // Scrittura
		kvc.GetDatastore()[message.GetKey()] = message.GetValue()

	} else if message.GetTypeOfMessage() == common.Del { // Scrittura
		kvc.DeleteFromDatastore(message.GetKey())

	} else if message.GetTypeOfMessage() == common.Get { // Lettura

		val, ok := kvc.GetDatastore()[message.GetKey()]
		if !ok {
			printRed("NON ESEGUITO", *message, kvc)
			result = false
		}

		response.SetValue(val)
		message.SetValue(val) //Fatto solo per DEBUG per il print
	}

	// A prescindere da result, verrà inviata una risposta al client
	if message.GetSenderID() == kvc.GetServerID() {
		response.SetResult(result)

		// ----- FIFO ORDERED ----- //
		// Imposto il numero di risposte inviate al determinato client
		response.SetReceptionFIFO(kvc.GetResponseOrderingFIFO(message.GetClientID()))
	}

	if result {
		printGreen("ESEGUITO", *message, kvc)
	}
}

func (kvc *KeyValueStoreCausale) createMessage(args common.Args, typeFunc string) *commonMsg.MessageC {
	kvc.LockMutexClock()
	defer kvc.UnlockMutexClock()

	// Incremento di uno l'indice del clock del server che ha ricevuto la richiesta
	kvc.SetVectorClock(kvc.GetServerID(), kvc.GetClock()[kvc.GetServerID()]+1)

	message := commonMsg.NewMessageC(kvc.GetServerID(), typeFunc, args, kvc.GetClock())

	printDebugBlue("RICEVUTO da client", *message, kvc)

	// Questo mutex mi permette di evitare scheduling tra il lascia passare di canReceive e la creazione del messaggio
	kvc.UnlockMutexMessage(message.GetClientID()) // Mutex chiuso in canReceive

	return message
}

// In canReceive, si vuole realizzare una mappa che aiuti nell'assunzione di una rete FIFO Ordered
func (kvc *KeyValueStoreCausale) canReceive(args common.Args) {
	idClient := args.GetClientID()

	receiveMessage := make(chan bool, 1)
	// Eseguo la funzione reale solo se la condizione è true
	go func() {
		// Attendo che il canale sia true impostato da canExecute
		args.WaitCondition() // Aspetta che la condizione sia true
		receiveMessage <- true
	}()

	// Se il client non è nella mappa, lo aggiungo e imposto il timestamp di ricezione e di invio a zero
	if !kvc.ExistClient(idClient) {
		// Non ho mai ricevuto un messaggio da questo client
		kvc.NewClientMap(idClient) // Inserisco il client nella mappa
	}

	// Il client è sicuramente nella mappa
	requestTs := kvc.GetReceiveTsFromClient(idClient) // Ottengo il numero di messaggi ricevuti da questo client

	if args.GetSendingFIFO() == requestTs { // Se il messaggio che ricevo dal client è quello che mi aspetto
		// Blocco il mutex per evitare che il client possa inviare un nuovo messaggio prima che io abbia
		// finito di creare il precedente
		kvc.LockMutexMessage(idClient)
		kvc.SetReceiveTsFromClient(idClient, args.GetSendingFIFO()+1) // Incremento il timestamp di ricezione del client
		args.SetCondition(true)                                       // Imposto la condizione a true, il messaggio può essere eseguito
	} else {
		kvc.AddBufferedArgs(args) // Aggiungo il messaggio in attesa di essere eseguito
	}
	<-receiveMessage // Attendo che la condizione sia true prima di terminare la funzione
}
