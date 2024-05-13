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
	args.ConfigureSafeBool() // Inizializzo il safeBool

	receiveMessage := make(chan bool, 1)

	// Eseguo la funzione reale solo se la condizione è true
	go func() {
		// Attendo che il canale sia true impostato da canExecute
		args.WaitCondition() // Aspetta che la condizione sia true
		receiveMessage <- true
	}()

	// Controllo se è possibile eseguire il messaggio a livello applicativo
	go kvc.canReceive(args)

	<-receiveMessage

	message := kvc.createMessage(args, operation)

	err := algorithms.SendToAllServer(common.Causal+".Update", *message, response)
	if err != nil {
		response.SetResult(false)
		return err
	}
	return nil
}

// RealFunction esegue l'operazione di put e di delete realmente
func (kvc *KeyValueStoreCausale) realFunction(message *commonMsg.MessageC, response *common.Response) error {

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
	if message.GetSenderID() == kvc.GetIdServer() {
		response.SetResult(result)

		// ----- FIFO ORDERED ----- //
		// Imposto il numero di risposte inviate al determinato client
		response.SetReceptionFIFO(kvc.GetResponseOrderingFIFO(message.GetClientID()))
	}

	if result {
		printGreen("ESEGUITO", *message, kvc)
	}

	// Controllo se è possibile gestire altre risposte
	go kvc.canHandleOtherResponse()
	return nil
}

func (kvc *KeyValueStoreCausale) createMessage(args common.Args, typeFunc string) *commonMsg.MessageC {
	kvc.LockMutexClock()
	defer kvc.UnlockMutexClock()

	// Incremento di uno l'indice del clock del server che ha ricevuto la richiesta
	kvc.SetVectorClock(kvc.GetIdServer(), kvc.GetClock()[kvc.GetIdServer()]+1)

	message := commonMsg.NewMessageC(kvc.GetIdServer(), typeFunc, args, kvc.GetClock())

	printDebugBlue("RICEVUTO da client", *message, kvc)

	// Questo mutex mi permette di evitare scheduling tra il lascia passare di canReceive e la creazione del messaggio
	kvc.UnlockMutexMessage(message.GetClientID()) // Mutex chiuso in canReceive

	return message
}

// In canReceive, si vuole realizzare una mappa che aiuti nell'assunzione di una rete FIFO Ordered
func (kvc *KeyValueStoreCausale) canReceive(args common.Args) bool {

	// Se il client non è nella mappa, lo aggiungo e imposto il timestamp di ricezione a zero
	if !kvc.ExistClient(args.GetClientID()) {
		// Non ho mai ricevuto un messaggio da questo client
		kvc.NewClientMap(args.GetClientID()) // Inserisco il client nella mappa
	}

	// Il client è sicuramente nella mappa
	requestTs := kvc.GetReceiveTsFromClient(args.GetClientID()) // Ottengo il timestamp di ricezione del client

	if args.GetSendingFIFO() == requestTs { // Se il messaggio che ricevo dal client è quello che mi aspetto
		// Blocco il mutex per evitare che il client possa inviare un nuovo messaggio prima che io abbia finito di creare il precedente
		kvc.LockMutexMessage(args.GetClientID())
		kvc.SetReceiveTsFromClient(args.GetClientID(), args.GetSendingFIFO()+1) // Incremento il timestamp di ricezione del client
		args.SetCondition(true)                                                 // Imposto la condizione a true, il messaggio può essere eseguito
		kvc.canHandleOtherRequest()
		return true // è possibile accettare il messaggio
	}
	kvc.AddBufferedArgs(args) // Aggiungo il messaggio in attesa di essere eseguito
	return false
}
