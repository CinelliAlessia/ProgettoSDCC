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