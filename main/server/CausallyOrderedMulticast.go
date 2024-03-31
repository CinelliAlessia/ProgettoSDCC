package main

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

func (kvc *KeyValueStoreCausale) CausallyOrderedMulticast() {

}
