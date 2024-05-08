package causal

import (
	"main/common"
	"main/server/algorithms"
	"main/server/message"
)

// Get restituisce il valore associato alla chiave specificata -> Lettura -> Evento interno
func (kvc *KeyValueStoreCausale) Get(args common.Args, response *common.Response) error {

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
		//response.SetReceptionFIFO(kvc.GetResponseOrderingFIFO(message.GetClientID())) // Setto il numero di risposte inviate al determinato client
		//kvc.IncreaseResponseOrderingFIFO(message.GetClientID())                       // Incrementa il numero di risposte inviate al determinato server
	}

	/*if result && message.GetIdSender() == kvc.GetIdServer() {
		printGreen("ESEGUITO mio", *message, kvc, nil)
	} else */if result {
		printGreen("ESEGUITO", *message, kvc)
	}

	return nil
}

func (kvc *KeyValueStoreCausale) createMessage(args common.Args, typeFunc string) *commonMsg.MessageC {
	kvc.mutexClock.Lock()
	defer kvc.mutexClock.Unlock()

	kvc.SetVectorClock(kvc.GetIdServer(), kvc.GetClock()[kvc.GetIdServer()]+1)

	message := commonMsg.NewMessageC(kvc.GetIdServer(), typeFunc, args, kvc.GetClock())

	printDebugBlue("RICEVUTO da client", *message, kvc)
	return message
}