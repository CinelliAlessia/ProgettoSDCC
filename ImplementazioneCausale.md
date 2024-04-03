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

Questa implementazione garantisce che i messaggi vengano consegnati solo quando sono stati rispettati i vincoli causali definiti nel tuo scenario.


*PROBLEMA*
- Identificazione dei server