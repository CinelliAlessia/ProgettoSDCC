package keyvaluestore

import (
	"main/common"
	"main/server/message"
)

// Get restituisce il valore associato alla chiave specificata -> Lettura -> Evento interno
func (kvc *KeyValueStoreCausale) Get(args common.Args, response *common.Response) error {

	message := kvc.createMessage(args, common.Get)

	err := sendToAllServer(common.Causal+".CausallyOrderedMulticast", *message, response)
	if err != nil {
		response.SetResult(false)
		return err
	}
	return nil
}

// Put inserisce una nuova coppia chiave-valore, se la chiave è già presente, sovrascrive il valore associato
func (kvc *KeyValueStoreCausale) Put(args common.Args, response *common.Response) error {

	message := kvc.createMessage(args, common.Put)

	err := sendToAllServer(common.Causal+".CausallyOrderedMulticast", *message, response)
	if err != nil {
		response.SetResult(false)
		return err
	}
	return nil
}

// Delete elimina una coppia chiave-valore, è un operazione di scrittura
func (kvc *KeyValueStoreCausale) Delete(args common.Args, response *common.Response) error {

	message := kvc.createMessage(args, common.Del)

	err := sendToAllServer(common.Causal+".CausallyOrderedMulticast", *message, response)
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
			printRed("NON ESEGUITO", *message, kvc, nil)
			//response.SetResult(false)
			//return nil
			if message.GetIdSender() == kvc.GetIdServer() {
				result = false
			}
		}

		response.SetValue(val)
		message.SetValue(val) //Fatto solo per DEBUG per il print
	}

	if message.GetIdSender() == kvc.GetIdServer() {
		response.SetResult(result)
		//response.SetReceptionFIFO(kvc.GetResponseOrderingFIFO())
		//kvc.IncreaseResponseOrderingFIFO()
	}

	if result {
		printGreen("ESEGUITO", *message, kvc, nil)
	}

	return nil
}

func (kvc *KeyValueStoreCausale) createMessage(args common.Args, typeFunc string) *commonMsg.MessageC {
	kvc.mutexClock.Lock()
	defer kvc.mutexClock.Unlock()

	kvc.SetVectorClock(kvc.GetIdServer(), kvc.GetClock()[kvc.GetIdServer()]+1)

	message := commonMsg.NewMessageC(kvc.GetIdServer(), typeFunc, args, kvc.GetClock())

	printDebugBlue("RICEVUTO da client", *message, kvc, nil)
	return message
}
