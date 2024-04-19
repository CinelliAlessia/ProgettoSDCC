package main

import (
	"fmt"
	"main/common"
	"time"
)

// CausallyOrderedMulticast esegue l'algoritmo multicast causalmente ordinato sul messaggio ricevuto.
// Aggiunge il messaggio alla coda dei messaggi in attesa di essere eseguiti e cicla finché il controlSendToApplication
// non restituisce true, indicando che la richiesta può essere eseguita a livello applicativo. Quando ciò accade,
// la funzione esegue effettivamente l'operazione a livello applicativo tramite la chiamata a RealFunction e rimuove
// il messaggio dalla coda. Restituisce un booleano tramite reply per indicare se l'operazione è stata eseguita con successo.
func (kvc *KeyValueStoreCausale) CausallyOrderedMulticast(message MessageC, response *common.Response) error {

	kvc.addToQueue(message)

	// Solo per DEBUG
	if kvc.Id != message.IdSender {
		kvc.printDebugBlue("RICEVUTO da server", message)
	}

	// Ciclo finché controlSendToApplication restituisce true
	// Controllo quando la richiesta può essere eseguita a livello applicativo
	response.Result = false
	// TODO: vale la pena aggiungere un mutex?
	for {
		canSend := kvc.controlSendToApplication(message)
		if canSend {
			// Invio a livello applicativo
			err := kvc.realFunction(message, response)
			if err != nil {
				return err
			}
			break
			/*if message.TypeOfMessage == "Get" && !(response.Result) {
				// Se un client chiede di leggere una chiave che non esiste,
				// aspetto che la chiave verrà inserita.
				time.Sleep(time.Millisecond * 100)
				continue
			} else {
				kvc.removeMessageToQueue(message)
				break
			}*/
		}
		// La richiesta non può essere ancora eseguita, si attende un breve periodo prima di riprovare
		time.Sleep(time.Millisecond * 1000)
	}
	return nil
}

// addToQueue aggiunge il messaggio passato come argomento alla coda.
func (kvc *KeyValueStoreCausale) addToQueue(message MessageC) {
	kvc.mutexQueue.Lock()
	defer kvc.mutexQueue.Unlock()
	kvc.Queue = append(kvc.Queue, message)
}

// controlSendToApplication realizza questo controllo:
// Quando il processo `pj` riceve il messaggio `m` da `pi`, lo mette in una coda d'attesa e ritarda la consegna al
// livello applicativo finché non si verificano entrambe le seguenti condizioni:
//   - `t(m)[i] = Vj[i] + 1` (il messaggio `m` è il successivo che `pj` si aspetta da `pi`).
//   - `t(m)[k] ≤ Vj[k]` per ogni processo `pk` diverso da `i` (ovvero `pj` ha visto almeno gli stessi messaggi di `pk` visti da `pi`).
func (kvc *KeyValueStoreCausale) controlSendToApplication(message MessageC) bool {
	result := false

	// Verifica se il messaggio m è il successivo che pj si aspetta da pi
	if (message.IdSender != kvc.Id) && (message.VectorClock[message.IdSender] == (kvc.VectorClock[message.IdSender] + 1)) {

		// Verifica se pj ha visto almeno gli stessi messaggi di pk visti da pi per ogni processo pk diverso da i
		for index := range message.VectorClock {
			if (index != message.IdSender) && (index != kvc.Id) && (message.VectorClock[index] > kvc.VectorClock[index]) {
				// pj non ha visto almeno gli stessi messaggi di pk visti da pi
				//fmt.Println("non ho ancora visto dei messaggi", "msg", message.VectorClock, "mio", kvc.VectorClock)
				result = false
			}
		}
		result = true
	} else if message.IdSender == kvc.Id { //Ho "ricevuto" una mia richiesta -> è possibile processarla
		result = true
	}

	if result {
		// Se è un evento di lettura, controllo se la chiave è presente nel mio datastore
		// Se non c'è aspetterò fin quando non verrà inserita
		if message.TypeOfMessage == "Get" { // Lettura
			_, result = kvc.Datastore[message.Args.Key]
		}
	}

	if result {
		// Entrambe le condizioni soddisfatte, il messaggio può essere consegnato al livello applicativo
		// Aggiorno il mio orologio vettoriale
		if message.IdSender != kvc.Id {
			kvc.mutexClock.Lock()
			kvc.VectorClock[message.IdSender] += 1
			kvc.mutexClock.Unlock()
		}

		kvc.removeMessageToQueue(message)
		return true
	}

	//fmt.Println("Messaggio", message.TypeOfMessage, message.Args.Key+":"+message.Args.Value, "non può essere ancora eseguito da", kvc.Id)
	//fmt.Println("VectorClock messaggio", message.VectorClock, "VectorClock mio", kvc.VectorClock)
	return false
}

// removeMessageToQueue Rimuove un messaggio dalla coda basato sull'ID del messaggio passato come argomento
func (kvc *KeyValueStoreCausale) removeMessageToQueue(message MessageC) {
	kvc.mutexQueue.Lock()
	defer kvc.mutexQueue.Unlock()

	for i, m := range kvc.Queue {
		if m.Id == message.Id {
			kvc.Queue = append(kvc.Queue[:i], kvc.Queue[i+1:]...)
			return
		}
	}
	fmt.Println("removeMessageToQueue: Messaggio con ID", message.Id, "non trovato nella coda")
}
