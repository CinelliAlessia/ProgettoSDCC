package causal

import (
	"fmt"
	"main/common"
	"main/server/message"
)

// Update esegue l'algoritmo multicast causalmente ordinato sul messaggio ricevuto.
// Aggiunge il messaggio alla coda dei messaggi in attesa di essere eseguiti e controlla finché il controlSendToApplication
// non restituisce true, indicando che la richiesta può essere eseguita a livello applicativo. Quando ciò accade,
// la funzione esegue effettivamente l'operazione a livello applicativo tramite la chiamata a realFunction e rimuove
// il messaggio dalla coda.
func (kvc *KeyValueStoreCausale) Update2(message commonMsg.MessageC, response *common.Response) error {
	(&message).ConfigureSafeBool()

	kvc.addToQueue(&message) // Aggiungi il messaggio alla coda

	// Solo per DEBUG
	if kvc.GetServerID() != (&message).GetSenderID() {
		printDebugBlue("RICEVUTO da server", message, kvc)
	}

	executeMessage := make(chan bool, 1)
	go func() {
		// Attendo che il canale sia true impostato da canExecute
		(&message).WaitCondition() // Aspetta che la condizione sia true, verrà impostato in canExecute se è possibile eseguire il messaggio a livello applicativo
		kvc.realFunction(&message, response)
		executeMessage <- true
	}()

	// Controllo se è possibile eseguire il messaggio a livello applicativo
	go kvc.canExecute(&message)

	<-executeMessage // Attendo che la condizione sia true
	// è stata eseguita real function nella goroutine, la risposta è stata popolata.
	return nil
}

func (kvc *KeyValueStoreCausale) Update(message commonMsg.MessageC, response *common.Response) error {

	kvc.addToQueue(&message) // Aggiungi il messaggio alla coda

	// Solo per DEBUG
	if kvc.GetServerID() != (&message).GetSenderID() {
		printDebugBlue("RICEVUTO da server", message, kvc)
	}

	kvc.canExecute(&message)             // Posso eseguire il mio messaggio
	kvc.realFunction(&message, response) // Eseguo la funzione reale

	go kvc.canHandleOtherResponse() // Controllo se posso gestire altri messaggi

	return nil
}

// addToQueue aggiunge il messaggio passato come argomento alla coda.
func (kvc *KeyValueStoreCausale) addToQueue(message *commonMsg.MessageC) {
	kvc.LockMutexQueue()
	defer kvc.UnlockMutexQueue()

	kvc.SetQueue(append(kvc.GetQueue(), *message))
}

// canExecute controlla se è possibile eseguire il messaggio a livello applicativo, se è possibile lo esegue
// e restituisce true, altrimenti restituisce false
func (kvc *KeyValueStoreCausale) canExecute2(message *commonMsg.MessageC) {
	// Valuto un messaggio alla volta
	kvc.executeFunctionMutex.Lock()
	defer kvc.executeFunctionMutex.Unlock()
	canSend := kvc.controlSendToApplication(message) // Controllo se le due condizioni del M.C.O sono soddisfatte

	if canSend {
		message.SetCondition(true) // Imposto la condizione a true
	} else {
		// Bufferizzo il messaggio
		kvc.AddBufferedMessage(*message)
	}
}

func (kvc *KeyValueStoreCausale) canExecute(message *commonMsg.MessageC) {
	message.ConfigureSafeBool()

	executeMessage := make(chan bool, 1)
	go func() {
		// Attendo che il canale sia true impostato da canExecute
		message.WaitCondition() // Aspetta che la condizione sia true, verrà impostato in canExecute se è possibile eseguire il messaggio a livello applicativo
		executeMessage <- true
	}()

	canSend := kvc.controlSendToApplication(message) // Controllo se le due condizioni del M.C.O sono soddisfatte

	if canSend {
		message.SetCondition(true) // Imposto la condizione a true
	} else {
		// Bufferizzo il messaggio
		kvc.AddBufferedMessage(*message)
	}
	<-executeMessage // Attendo che la condizione sia true
	// è stata eseguita real function nella goroutine, la risposta è stata popolata.

}

// controlSendToApplication :
//
//	Quando il processo corrente `pj` riceve il messaggio `m` da `pi`, lo mette in una coda d'attesa e ritarda la consegna
//	a livello applicativo finché non si verificano entrambe le seguenti condizioni:
//	  - `t(m)[i] = Vj[i] + 1` (il messaggio `m` è il successivo che `pj` si aspetta da `pi`).
//	  - `t(m)[k] ≤ Vj[k]` per ogni processo `pk` diverso da `i` (ovvero `pj` ha visto almeno gli stessi messaggi di `pk` visti da `pi`).
func (kvc *KeyValueStoreCausale) controlSendToApplication(message *commonMsg.MessageC) bool {
	kvc.LockMutexClock()
	defer kvc.UnlockMutexClock()

	result := false

	// Verifica se il messaggio m è il successivo che pj si aspetta da pi
	if (message.GetSenderID() != kvc.GetServerID()) && // Se il mittente del messaggio non è il server stesso che sta processando il messaggio
		(message.GetClock()[message.GetSenderID()] == (kvc.GetClock()[message.GetSenderID()] + 1)) { // e il messaggio m è il successivo che pj (Io) si aspetta da pi

		result = true

		// Verifica se pj ha visto almeno gli stessi messaggi di pk visti da pi per ogni processo pk diverso da i
		for index := range message.GetClock() { // Per ogni indice del vettore dei clock logici
			if (index != message.GetSenderID()) && // Se l'indice non è quello del mittente del messaggio
				(index != kvc.GetServerID()) && // e non è quello del server stesso che sta processando il messaggio
				(message.GetClock()[index] > kvc.GetClock()[index]) { // e pj non ha visto almeno gli stessi messaggi di pk visti da pi
				result = false
			}
		}

	} else if message.GetSenderID() == kvc.GetServerID() { //Ho ricevuto una mia richiesta -> è possibile processarla
		result = true
	}

	// Se è un evento di lettura, controllo se la chiave è presente nel mio datastore
	// Se non c'è aspetterò fin quando non verrà inserita
	if message.GetTypeOfMessage() == common.Get && result {
		_, result = kvc.GetDatastore()[message.GetKey()]
		fmt.Println("Get: variabile non presente nel datastore")
	}

	if result {
		// Entrambe le condizioni soddisfatte, il messaggio può essere consegnato al livello applicativo
		// Aggiorno il mio orologio vettoriale
		if message.GetSenderID() != kvc.GetServerID() {
			kvc.SetVectorClock(message.GetSenderID(), kvc.GetClock()[message.GetSenderID()]+1)
		} else {
			// Incremento il numero di risposte inviate al determinato client
			kvc.SetResponseOrderingFIFO(message.GetClientID(), 1)
		}

		kvc.removeMessageToQueue(message) // Rimuovo il messaggio dalla coda
		return true
	}

	fmt.Println("Non è possibile eseguire", message.GetTypeOfMessage(), message.GetKey()+":"+message.GetValue())
	return false
}

// removeMessageToQueue Rimuove un messaggio dalla coda basato sull'ID del messaggio passato come argomento
func (kvc *KeyValueStoreCausale) removeMessageToQueue(message *commonMsg.MessageC) {
	kvc.LockMutexQueue()
	defer kvc.UnlockMutexQueue()

	for i, m := range kvc.GetQueue() {
		if m.GetIdMessage() == message.GetIdMessage() {
			kvc.SetQueue(append(kvc.GetQueue()[:i], kvc.GetQueue()[i+1:]...))
			return
		}
	}
}
