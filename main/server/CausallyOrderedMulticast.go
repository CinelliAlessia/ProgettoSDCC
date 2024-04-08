package main

import (
	"fmt"
	"main/common"
	"time"
)

/* Multicast causalmente ordinato utilizzano il clock logico vettoriale.

1. Quando il processo `pi` invia un messaggio `m`, crea un timestamp `t(m)` basato sul clock logico vettoriale `Vi`,
dove `Vj[i]` conta il numero di messaggi che `pi` ha inviato a `pj`.

2. Quando il processo `pj` riceve il messaggio `m` da `pi`, lo mette in una coda d'attesa e ritarda la consegna al
livello applicativo finché non si verificano entrambe le seguenti condizioni:
   - `t(m)[i] = Vj[i] + 1` (il messaggio `m` è il successivo che `pj` si aspetta da `pi`).
   - `t(m)[k] ≤ Vj[k]` per ogni processo `pk` diverso da `i` (ovvero `pj` ha visto almeno gli stessi messaggi di `pk` visti da `pi`).

Ecco un esempio di implementazione in pseudocodice:

Quando il processo P_i desidera inviare un messaggio multicast M:
    1. Incrementa il proprio vettore di clock logico V_i.
    2. Aggiunge V_i come timestamp al messaggio M.
    3. Invia M a tutti i processi.

Quando il processo P_j riceve un messaggio multicast M da P_i:
    1. Mette M nella coda d'attesa.
    2. Se M è il prossimo messaggio che P_j si aspetta da P_i e P_j ha visto almeno gli stessi messaggi di ogni altro
	processo P_k, consegna M all'applicazione.

Questa implementazione garantisce che i messaggi vengano consegnati solo quando sono stati rispettati i vincoli causali
	definiti nel tuo scenario.*/

func (kvc *KeyValueStoreCausale) CausallyOrderedMulticast(message MessageC, reply *bool) error {
	fmt.Println("CausallyOrderedMulticast: Ho ricevuto la richiesta che mi è stata inoltrata da un server")

	kvc.addToQueue(message)

	// Ciclo finché controlSendToApplication non restituisce true
	// Controllo quando la richiesta può essere eseguita a livello applicativo
	*reply = false
	for {
		canSend := kvc.controlSendToApplication(message)
		if canSend {

			// Invio a livello applicativo
			var replySaveInDatastore common.Response
			err := kvc.RealFunction(message, &replySaveInDatastore)
			if err != nil {
				return err
			}

			*reply = true
			break
		}

		// Altrimenti, attendi un breve periodo prima di riprovare
		time.Sleep(time.Millisecond * 1000)
	}
	return nil
}

func (kvc *KeyValueStoreCausale) addToQueue(message MessageC) {
	kvc.mutexQueue.Lock()
	defer kvc.mutexQueue.Unlock()

	kvc.queue = append(kvc.queue, message)

	// Ordina la coda in base all'orologio vettoriale e altri criteri
	/* sort.Slice(kvc.queue, func(i, j int) bool {
		// Confronto basato sull'orologio vettoriale
		for idx := range kvc.queue[i].VectorClock {
			if kvc.queue[i].VectorClock[idx] != kvc.queue[j].VectorClock[idx] {
				return kvc.queue[i].VectorClock[idx] < kvc.queue[j].VectorClock[idx]
			}
		}
		// Se l'orologio vettoriale è lo stesso, ordina in base all'ID
		return kvc.queue[i].Id < kvc.queue[j].Id
	}) */
}

/*
2. Quando il processo `pj` riceve il messaggio `m` da `pi`, lo mette in una coda d'attesa e ritarda la consegna al
livello applicativo finché non si verificano entrambe le seguenti condizioni:
  - `t(m)[i] = Vj[i] + 1` (il messaggio `m` è il successivo che `pj` si aspetta da `pi`).
  - `t(m)[k] ≤ Vj[k]` per ogni processo `pk` diverso da `i` (ovvero `pj` ha visto almeno gli stessi messaggi di `pk` visti da `pi`).
*/
func (kvc *KeyValueStoreCausale) controlSendToApplication(message MessageC) bool {
	// Verifica se il messaggio m è il successivo che pj si aspetta da pi
	if message.IdSender != kvc.id && message.VectorClock[message.IdSender] == kvc.vectorClock[message.IdSender]+1 {

		// Verifica se pj ha visto almeno gli stessi messaggi di pk visti da pi per ogni processo pk diverso da i
		for index := range message.VectorClock {
			if index != message.IdSender && message.VectorClock[index] > kvc.vectorClock[index] {
				// pj non ha visto almeno gli stessi messaggi di pk visti da pi
				fmt.Println("Index", index, "message.Id", message.IdSender, "", message.VectorClock[index], "", kvc.vectorClock[index])
				return false
			}
		}
		// Entrambe le condizioni soddisfatte, il messaggio può essere consegnato al livello applicativo
		fmt.Println("La condizione è soddisfatta", message.TypeOfMessage, message.Args.Key, message.Args.Value, message.VectorClock, "atteso (no +1):", kvc.vectorClock, "id", message.IdSender)
		kvc.mutexClock.Lock()
		kvc.vectorClock[message.IdSender] = message.VectorClock[message.IdSender]
		kvc.mutexClock.Unlock()
		return true
	} else if message.IdSender == kvc.id {
		fmt.Println("La condizione è soddisfatta", message.TypeOfMessage, message.Args.Key, message.Args.Value, message.VectorClock, "atteso:", kvc.vectorClock, "id", message.IdSender)
		return true
	}
	// Una delle condizioni non è soddisfatta, il messaggio non può essere consegnato al livello applicativo
	fmt.Println("La condizione NON è soddisfatta", message.TypeOfMessage, message.Args.Key, message.Args.Value, message.VectorClock, "atteso (no +1):", kvc.vectorClock, "id", message.IdSender)
	return false
}

// removeByID Rimuove un messaggio dalla coda basato sull'ID del messaggio
func (kvc *KeyValueStoreCausale) removeMessageToQueue(message MessageC) {
	kvc.mutexQueue.Lock()
	defer kvc.mutexQueue.Unlock()
	if kvc.queue[0].Id == message.Id {
		// Rimuovi l'elemento dalla slice
		kvc.queue = kvc.queue[1:]
		return
	}
	fmt.Println("AAA removeByID: Messaggio con ID", message.Id, "non trovato nella coda")
}
