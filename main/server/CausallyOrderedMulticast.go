package main

import (
	"fmt"
	"main/common"
	"reflect"
	"sort"
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

	kvc.addToSortQueue(message)

	// Aggiornamento del clock
	kvc.mutexClock.Lock()
	for i := 0; i < common.Replicas; i++ {
		kvc.vectorClock[i] = common.Max(message.VectorClock[i], kvc.vectorClock[i])
	}
	kvc.mutexClock.Unlock()

	// Ciclo finché controlSendToApplication non restituisce true
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
			break // Esci dal ciclo se controlSendToApplication restituisce true
		}

		// Altrimenti, attendi un breve periodo prima di riprovare
		time.Sleep(time.Millisecond * 100)
	}
	return nil
}

func (kvc *KeyValueStoreCausale) addToSortQueue(message MessageC) {
	kvc.mutexQueue.Lock()
	defer kvc.mutexQueue.Unlock()

	kvc.queue = append(kvc.queue, message)

	// Ordina la coda in base all'orologio vettoriale e altri criteri
	sort.Slice(kvc.queue, func(i, j int) bool {
		// Confronto basato sull'orologio vettoriale
		for idx := range kvc.queue[i].VectorClock {
			if kvc.queue[i].VectorClock[idx] != kvc.queue[j].VectorClock[idx] {
				return kvc.queue[i].VectorClock[idx] < kvc.queue[j].VectorClock[idx]
			}
		}
		// Se l'orologio vettoriale è lo stesso, ordina in base all'ID
		return kvc.queue[i].Id < kvc.queue[j].Id
	})
}

/*
2. Quando il processo `pj` riceve il messaggio `m` da `pi`, lo mette in una coda d'attesa e ritarda la consegna al
livello applicativo finché non si verificano entrambe le seguenti condizioni:
  - `t(m)[i] = Vj[i] + 1` (il messaggio `m` è il successivo che `pj` si aspetta da `pi`).
  - `t(m)[k] ≤ Vj[k]` per ogni processo `pk` diverso da `i` (ovvero `pj` ha visto almeno gli stessi messaggi di `pk` visti da `pi`).
*/
func (kvc *KeyValueStoreCausale) controlSendToApplication(message MessageC) bool {
	// Verifica se il messaggio è il prossimo atteso
	expectedTimestamp := make([]int, common.Replicas)
	//copy(expectedTimestamp, kvc.vectorClock)

	senderID := 0                 // Dovrai definire un modo per identificare l'ID del mittente
	expectedTimestamp[senderID]++ // Il messaggio successivo che pj si aspetta da pi

	if !reflect.DeepEqual(expectedTimestamp, message.VectorClock) {
		return false
	}

	// Verifica se il messaggio ha ricevuto tutti gli ack
	if message.NumberAck == common.Replicas {
		// Il messaggio ha ricevuto tutti gli ack, quindi può essere rimosso dalla coda
		kvc.removeMessage(message)
		return true
	}

	return false
}

// removeByID Rimuove un messaggio dalla coda basato sull'ID del messaggio
func (kvc *KeyValueStoreCausale) removeMessage(message MessageC) {
	for i := range kvc.queue {
		if kvc.queue[i].Id == message.Id {
			// Rimuovi l'elemento dalla slice
			kvc.queue = append(kvc.queue[:i], kvc.queue[i+1:]...)
			fmt.Println("removeByID: Messaggio con ID", message.Id, "rimosso dalla coda")
			return
		}
	}
	fmt.Println("removeByID: Messaggio con ID", message.Id, "non trovato nella coda")
}
