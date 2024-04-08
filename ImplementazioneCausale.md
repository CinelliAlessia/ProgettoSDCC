**Multicast causalmente ordinato:**

Un messaggio viene consegnato solo se tutti i messaggi che lo precedono
causalmente (secondo la relazione di causa-effetto) sono stati già consegnati. Si tratta di un
indebolimento del multicast totalmente ordinato.
Mantenendo le assunzioni sulla comunicazione affidabile e FIFO ordered, introduciamo un algoritmo che fa
uso del clock logico vettoriale per risolvere il problema del multicast causalmente ordinato in modo
completamente distribuito:
1. pi invia un messaggio m con timestamp t(m) basato sul clock logico vettoriale Vi, dove stavolta Vj[i] conta
   il numero di messaggi che pi ha inviato a pj (per cui c’è una differenza con l’utilizzo standard dei clock logici
   vettoriali).
2. pj riceve m da pi e ne ritarda la consegna al livello applicativo (ponendo m in una coda d’attesa) finché
   non si verificano entrambe le seguenti condizioni:
1) t(m)[i] = Vj[i]+1 (m è il messaggio successivo che pj si aspetta da pi).
2) t(m)[k] <= Vj[k] per ogni k != i (per ogni processo pk, pj ha visto almeno gli stessi messaggi di pk visti da pi).

Ecco un esempio di implementazione in pseudocodice:

Quando il processo P_i desidera inviare un messaggio multicast M:
1. Incrementa il proprio vettore di clock logico V_i.
2. Aggiunge V_i come timestamp al messaggio M.
3. Invia M a tutti i processi.

Quando il processo P_j riceve un messaggio multicast M da P_i:
1. Mette M nella coda d'attesa.
2. Se M è il prossimo messaggio che P_j si aspetta da P_i e P_j ha visto almeno gli stessi messaggi di ogni altro processo P_k, consegna M all'applicazione.

### Pseudo-codice
// Definizione dei processi e delle strutture dati
processes = {p1, p2, ..., pn} Insieme di tutti i processi

sentMessages = array di array di interi di dimensione n x n  // Conta i messaggi inviati da ogni processo a ogni altro processo
receivedMessages = array di array di interi di dimensione n x n  // Conta i messaggi ricevuti da ogni processo da ogni altro processo

// Inizializzazione
per ogni i da 1 a n:
per ogni j da 1 a n:
sentMessages[i][j] = 0
receivedMessages[i][j] = 0

// Funzione per inviare un messaggio da un processo a un altro
funzione send(processo mittente, processo destinatario, messaggio):
sentMessages[mittente][destinatario] += 1  // Aggiorna il contatore dei messaggi inviati
invia(messaggio, destinatario)  // Invia il messaggio al destinatario

// Funzione per ricevere un messaggio da un processo
funzione receive(processo destinatario, messaggio):
receivedMessages[mittente][destinatario] += 1  // Aggiorna il contatore dei messaggi ricevuti

// Funzione per verificare se un messaggio può essere consegnato al livello applicativo
funzione canDeliverMessage(messaggio, processo destinatario):
per ogni processo p diverso da destinatario:
se sentMessages[messaggio.sender][p] > receivedMessages[p][destinatario]:
// Il processo destinatario non ha ricevuto tutti i messaggi inviati da p
ritorna falso

    se messaggio.sender è diverso da destinatario e sentMessages[messaggio.sender][destinatario] == receivedMessages[destinatario][messaggio.sender] + 1:
        // Tutti i messaggi causali sono stati ricevuti e il messaggio attuale è il successivo atteso
        ritorna vero
    altrimenti:
        ritorna falso

// Funzione per consegnare un messaggio al livello applicativo
funzione deliverMessage(messaggio):
consegnare(messaggio)  // Consegnare il messaggio al livello applicativo
receivedMessages[messaggio.sender][messaggio.receiver] += 1  // Aggiorna il contatore dei messaggi ricevuti

// Processo principale di ricezione dei messaggi
processo receiveLoop(processo destinatario):
finché vero:
messaggio = ricevi()  // Ricevi un messaggio
se canDeliverMessage(messaggio, destinatario):
deliverMessage(messaggio)  // Consegnare il messaggio al livello applicativo

// Processo principale di invio dei messaggi
processo sendLoop(processo mittente, processo destinatario):
finché vero:
messaggio = creaMessaggio()  // Creare un nuovo messaggio
send(mittente, destinatario, messaggio)  // Inviare il messaggio al destinatario
attendi(acknowledgment)  // Attendere l'ACK dal destinatario

Quando un processo P_i desidera inviare un messaggio multicast M:
1. Incrementa il proprio vettore di clock logico V_i.
2. Aggiunge V_i come timestamp al messaggio M.
3. Invia M a tutti i processi.

Quando un processo P_j riceve un messaggio multicast M da P_i:
1. Mette M nella coda d'attesa.
2. Finché M non è pronto per essere consegnato al livello applicativo:
   - Verifica se il timestamp di M è il successivo atteso dall'orologio vettoriale di P_j per P_i.
   - Verifica se P_j ha visto almeno gli stessi messaggi di ogni altro processo P_k diverso da P_i.
   - Se entrambe le condizioni sono soddisfatte:
   - Incrementa l'orologio vettoriale di P_j per riflettere il ricevimento di M.
   - Consegna M al livello applicativo.
   - Altrimenti, ritarda la consegna di M e attendi.

*PROBLEMA*
- Identificazione dei server