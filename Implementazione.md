PROGETTO:
Mantenere le repliche di un datastore coerenti garantendo le consistenze 
causali e sequenziali a seconda della scelta del "client".

1) Il server (i server replica):
   Mantengono una replica di un dataStore
   Ricevono le richieste da un client
2) Il servizio RPC sequenziale:
   
3) Il servizio RPC causale:
4) Il test:

   Chi utilizzerà i server stabilisce prima di iniziare il tipo di consistenza che si vuole utilizzare. 

Due implementazioni dell'interfaccia delle procedure RPC: 
1) Causale: Rispetta la causa effetto -> Clock logico vettoriale
2) Sequenziale: Tutte le repliche vedono la stessa sequenza. -> Multicast totalmente ordinato

Nel messaggio inviato oltre l'evento c'è il clock -> cio mi obbliga a fare due interfacce, perchè un clock è scalare
e l'altro è vettoriale.


import "github.com/google/uuid"

// Genera un ID univoco utilizzando UUID
func generateUniqueID() string {
id := uuid.New()
return id.String()
}


**Package:**

project/
│
├── main.go
├── key_value_store.go
├── rpc_server.go
├── rpc_client.go
└── config.go


**RPC**
La decisione di utilizzare un parametro per ricevere una risposta anziché ottenerla come valore di ritorno dipende dal fatto che stiamo utilizzando RPC (Remote Procedure Call).

Quando si utilizza RPC, le chiamate di funzione devono essere serializzate e deserializzate attraverso la rete. Questo significa che non possiamo restituire direttamente un valore di ritorno dalla funzione chiamata sul server al client come si farebbe in una chiamata di funzione locale.

Invece, i parametri che vengono passati per riferimento (come puntatori in Go) possono essere serializzati e deserializzati facilmente attraverso la rete. Questo è il motivo per cui utilizziamo un parametro per ricevere la risposta nel caso delle chiamate RPC.

Quando chiamiamo una funzione RPC, passiamo un parametro per contenere la risposta deserializzata. Questo parametro viene modificato dalla funzione RPC sul server e quindi contiene il risultato della chiamata RPC quando la funzione ritorna. Questo approccio ci consente di ottenere il risultato della chiamata RPC nel client.

Con repliche del datastore, è essenziale garantire che ogni scrittura sia immediatamente replicata su tutte le altre repliche per mantenere la consistenza dei dati tra di esse. Puoi fare ciò attraverso una strategia di replicazione sincrona o asincrona.

**REPLICAZIONE**

Replicazione sincrona: Con la replicazione sincrona, ogni scrittura viene confermata solo dopo che è stata replicata con successo su tutte le altre repliche. Questo garantisce che tutte le repliche contengano la stessa copia dei dati e che ogni scrittura sia visibile a tutte le repliche prima di essere confermata. Tuttavia, questa approccio potrebbe rallentare le operazioni di scrittura a causa della necessità di attendere che tutte le repliche confermino la scrittura.

Replicazione asincrona: Con la replicazione asincrona, ogni scrittura viene replicata su altre repliche in modo asincrono, senza dover attendere la conferma da parte di tutte le repliche prima di confermare la scrittura. Questo può migliorare le prestazioni delle operazioni di scrittura poiché non c'è bisogno di attendere la conferma da parte di tutte le repliche. Tuttavia, c'è il rischio di perdere dati se una replica non riesce a replicare correttamente una scrittura.

Nella consistenza sequenziale dovrebbe essere necessaria la replicazione sincrona.

**DOCKER COMPOSE:**

Passaggio della porta tramite variabile d'ambiente o parametro di configurazione: Puoi passare la porta del servizio
come variabile d'ambiente o parametro di configurazione al momento dell'avvio del contenitore. Ad esempio, puoi passare la porta come variabile d'ambiente:

docker run -e PORT=8080 nome_contenitore

Quindi, all'interno del contenitore, puoi accedere alla porta utilizzando la variabile d'ambiente PORT.


**ALGORITMO CLOCK LOGICO SCALARE**
1) Tutti gli N processi hanno il clock = 0
2) Per ogni evento interno il clock aumenta di uno
3) Se si riceve un messaggio da un altro processo:
   1) Si prende il max clock logico
   2) Lo si incrementa di uno
   3) Si esegue l'evento di receive(m)
4) Se si invia un messaggio ad un altro processo:
   1) Si incrementa il clock di uno
   2) Allega al messaggio il clock logico
   3) Esegue l'evento di send(m)

Nel mio progetto un evento di send e receive sono le "stabilizzazioni" delle repliche, la richiesta
di scrittura da parte degli altri server. 


- Nel tuo file di configurazione Docker Compose, hai definito tre repliche per il servizio "server". Ogni replica ha porte mappate sulla porta ${RPC_PORT} su host specifici, come segue:

- La prima replica ha le porte mappate come segue: "8081:${RPC_PORT}". Questo significa che la porta 8081 sull'host sarà mappata alla porta specificata da ${RPC_PORT} nel container della prima replica.

- La seconda replica ha le porte mappate come segue: "8082:${RPC_PORT}". Questo significa che la porta 8082 sull'host sarà mappata alla porta specificata da ${RPC_PORT} nel container della seconda replica.

- La terza replica ha le porte mappate come segue: "8083:${RPC_PORT}". Questo significa che la porta 8083 sull'host sarà mappata alla porta specificata da ${RPC_PORT} nel container della terza replica.


- Mi arriva prima un ack e dopo la richiesta di append



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

Quando il processo pi invia un messaggio m, crea un timestamp t(m) basato sul clock logico vettoriale Vi, dove Vj[i] conta il numero di messaggi che pi ha inviato a pj.
Quando il processo pj riceve il messaggio m da pi, lo mette in una coda d'attesa e ritarda la consegna al livello applicativo finché non si verificano entrambe le seguenti condizioni:

t(m)[i] = Vj[i] + 1 (il messaggio m è il successivo che pj si aspetta da pi).
t(m)[k] ≤ Vj[k] per ogni processo pk diverso da i (ovvero pj ha visto almeno gli stessi messaggi di pk visti da pi).
Ecco un esempio di implementazione in pseudocodice:

Quando il processo P_i desidera inviare un messaggio multicast M:
1. Incrementa il proprio vettore di clock logico V_i.
2. Aggiunge V_i come timestamp al messaggio M.
3. Invia M a tutti i processi.

Quando il processo P_j riceve un messaggio multicast M da P_i:
1. Mette M nella coda d'attesa.
2. Se M è il prossimo messaggio che P_j si aspetta da P_i e P_j ha visto almeno gli stessi messaggi di ogni altro processo P_k, consegna M all'applicazione.

Questa implementazione garantisce che i messaggi vengano consegnati solo quando sono stati rispettati i vincoli causali definiti nel tuo scenario.