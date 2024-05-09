package causal

import (
	"main/common"
	"main/server/algorithms"
	"main/server/message"
)

// Get restituisce il valore associato alla chiave specificata -> Lettura -> Evento interno
func (kvc *KeyValueStoreCausale) Get(args common.Args, response *common.Response) error {
	for !kvc.canReceive(args) {
		// Aspetta di ricevere tutti i messaggi precedenti da parte del client
	}
	message := kvc.createMessage(args, common.Get)

	err := algorithms.SendToAllServer(common.Causal+".CausallyOrderedMulticast", *message, response)
	if err != nil {
		response.SetResult(false)
		return err
	}
	return nil
}

// Put inserisce una nuova coppia chiave-valore, se la chiave è già presente, sovrascrive il valore associato
func (kvc *KeyValueStoreCausale) Put(args common.Args, response *common.Response) error {
	for !kvc.canReceive(args) {
		// Aspetta di ricevere tutti i messaggi precedenti da parte del client
	}
	message := kvc.createMessage(args, common.Put)

	err := algorithms.SendToAllServer(common.Causal+".CausallyOrderedMulticast", *message, response)
	if err != nil {
		response.SetResult(false)
		return err
	}
	return nil
}

// Delete elimina una coppia chiave-valore, è un operazione di scrittura
func (kvc *KeyValueStoreCausale) Delete(args common.Args, response *common.Response) error {
	for !kvc.canReceive(args) {
		// Aspetta di ricevere tutti i messaggi precedenti da parte del client
	}

	message := kvc.createMessage(args, common.Del)

	err := algorithms.SendToAllServer(common.Causal+".CausallyOrderedMulticast", *message, response)
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
			if message.GetIdSender() == kvc.GetIdServer() {
				result = false
			}
		}

		response.SetValue(val)
		message.SetValue(val) //Fatto solo per DEBUG per il print
	}

	// A prescindere da result, verrà inviata una risposta al client
	if message.GetIdSender() == kvc.GetIdServer() {
		response.SetResult(result)
		response.SetReceptionFIFO(kvc.GetResponseOrderingFIFO(message.GetClientID())) // Setto il numero di risposte inviate al determinato client

		kvc.SetResponseOrderingFIFO(message.GetClientID(), kvc.GetResponseOrderingFIFO(message.GetClientID())+1) // Incremento il numero di risposte inviate al determinato client
	}

	if result {
		printGreen("ESEGUITO", *message, kvc)
	}

	return nil
}

func (kvc *KeyValueStoreCausale) createMessage(args common.Args, typeFunc string) *commonMsg.MessageC {
	kvc.mutexClock.Lock()
	defer kvc.mutexClock.Unlock()

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
	requestTs, err := kvc.GetReceiveTsFromClient(args.GetClientID()) // Ottengo il timestamp di ricezione del client

	if args.GetSendingFIFO() == requestTs && // Se il messaggio che ricevo dal client è quello che mi aspetto
		err == nil { // Se non ci sono stati errori

		// Blocco il mutex per evitare che il client possa inviare un nuovo messaggio prima che io abbia finito di creare il precedente
		kvc.LockMutexMessage(args.GetClientID())
		kvc.SetRequestClient(args.GetClientID(), args.GetSendingFIFO()+1) // Incremento il timestamp di ricezione del client
		return true                                                       // è possibile accettare il messaggio
	}
	return false
}
