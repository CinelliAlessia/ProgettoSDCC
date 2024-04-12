**Dipendenze Esterne** Per garantire il corretto funzionamento del sistema, sono state aggiunte alcune dipendenze esterne che facilitano determinate funzionalità all'interno del codice. In particolare, abbiamo incluso le seguenti dipendenze:

*   **github.com/google/uuid**: Questa libreria è stata utilizzata per generare in modo univoco gli identificatori univoci (UUID) necessari per identificare in modo univoco le chiavi all'interno del nostro sistema di memorizzazione chiave-valore distribuito.
*   **github.com/fatih/color**: Prima di poter utilizzare questa libreria per la gestione dei colori, è stata eseguita una procedura di installazione tramite il comando "go get -u github.com/fatih/color". Questa libreria ci permette di migliorare la leggibilità del nostro output a schermo, fornendo colori distintivi per determinati messaggi o output.

**Configurazione Locale e Docker** 
Per agevolare l'utilizzo del sistema in diversi ambienti, ho incluso la possibilità di eseguire l'applicazione sia in un ambiente locale che tramite Docker. Questo è reso possibile da una semplice modifica di una variabile globale nel file "start.sh". Modificando questa variabile, è possibile selezionare se eseguire il sistema in un ambiente locale o all'interno di un container Docker, offrendo così una maggiore flessibilità nell'ambiente di esecuzione.

**Configurazione Debug Intensivo**
Similmente alla variabile d'ambiente per l'utilizzo del sistema in locale o su docker è stata aggiunta una variabile per indicare se si vorrebbero avere maggior print di debug per una spiegazione esplicita o meno, modificabile nel file "start.sh" 

**Garanzie Consistenza**
Questo progetto consiste nel realizzare un server che dia garanzia, a scelta del client, di avere repliche di datatstore che rispettino la consistenza sequenziale o la consistenza causale. 

- Per garantire la consistenza sequenziale si utilizza l'algoritmo di multicast totalmente ordinato.
- Per garantire la consistenza causale si utilizza l'algoritmo di multicast causalmente ordinato.

***SEQUENZIALE:***

**Multicast Totalmente Ordinato**

1. Un client effettua una chiamata RPC con la propria richiesta ad un singolo server (scelto random).
2. Il server che riceve la richiesta incrementa il suo clock logico scalare e lo allega alla richiesta.
3. La gestione avviene in maniera differente a seconda se è un evento interno o esterno.

*Gestione dell'evento interno: GET*
   1. Aggiunge la richiesta a una coda locale ordinata per timestamp.
   2. La esegue quando la richiesta sarà la prima nella coda. 

*Gestione degli eventi esterni: PUT e DELETE*
   1. Inoltra in broadcast la richiesta a tutti i server (incluso se stesso).
   2. Ciascun server che riceve la richiesta:
      1. La aggiunge a una coda locale ordinata per timestamp della richiesta.
      2. Modifica il suo clock locale
      3. Invia un ack a tutti i server per indicare che lui ha letto quel messaggio.
      4. Esegue la richiesta ricevuta esclusivamente se: 
         - La richiesta è la prima nella sua coda (ha timestamp minore di tutte le altre in coda).
         - Sono stati ricevuti tutti gli ack relativi a quella richiesta.

- L'invio dell'ack a tutti i server avviene in maniera asincrona.

***CAUSALE:***

**Multicast Causalmente Ordinato**

1. Un client effettua una chiamata RPC con la propria richiesta ad un singolo server (scelto random).
2. Il server che riceve la richiesta incrementa il suo valore del clock logico vettoriale e lo allega alla richiesta.
3. Il server invia in multicast il messaggio creato.
4. Ciascun server ricevente:
   1. Aggiunge la richiesta nella coda.
   2. Controlla attivamente se la richiesta ricevuta può essere inviata a livello applicativo, se si, verrà rimossa dalla coda.

Il controllo della richiesta avviene controllando due parametri, se pj è il server che riceve la richiesta fatta da pi:
1) t(m)[i] = Vj[i]+1 (m è il messaggio successivo che pj si aspetta da pi).
2) t(m)[k] <= Vj[k] per ogni k != i (per ogni processo pk, pj ha visto almeno gli stessi messaggi di pk visti da pi).
Se entrambe queste condizioni sono rispettate, il messaggio verrà inviato a livello applicativo.

