**Dipendenze Esterne** Per garantire il corretto funzionamento del sistema, sono state aggiunte alcune dipendenze 
esterne che facilitano determinate funzionalità all'interno del codice. In particolare, abbiamo incluso le seguenti 
dipendenze:

*   **github.com/google/uuid**: Questa libreria è stata utilizzata per generare in modo univoco gli identificatori 
* univoci (UUID) necessari per identificare in modo univoco le chiavi all'interno del nostro sistema di memorizzazione 
* chiave-valore distribuito.
*   **github.com/fatih/color**: Prima di poter utilizzare questa libreria per la gestione dei colori, è stata eseguita 
* una procedura di installazione tramite il comando "go get -u github.com/fatih/color". Questa libreria ci permette di 
* migliorare la leggibilità del nostro output a schermo, fornendo colori distintivi per determinati messaggi o output.

**Configurazione Locale e Docker** 
Per agevolare l'utilizzo del sistema in diversi ambienti, ho incluso la possibilità di eseguire l'applicazione sia in 
un ambiente locale che tramite Docker. Questo è reso possibile da una semplice modifica di una variabile globale nel 
file "start.sh". Modificando questa variabile, è possibile selezionare se eseguire il sistema in un ambiente locale 
o all'interno di un container Docker, offrendo così una maggiore flessibilità nell'ambiente di esecuzione.

**Configurazione Debug Intensivo**
Similmente alla variabile d'ambiente per l'utilizzo del sistema in locale o su docker è stata aggiunta una variabile 
per indicare se si vorrebbero avere maggior print di debug per una spiegazione esplicita o meno, modificabile nel file 
"start.sh" 

**Garanzie Consistenza**
Questo progetto consiste nel realizzare un server che dia garanzia, a scelta del client, di avere repliche di datastore
che rispettino la consistenza sequenziale o la consistenza causale. 

- Per garantire la consistenza sequenziale si utilizza l'algoritmo di multicast totalmente ordinato.
- Per garantire la consistenza causale si utilizza l'algoritmo di multicast causalmente ordinato.

**Multicast Causalmente Ordinato**

Il client effettua le richieste RPC chiamando le `KeyValueStoreCausale.Put`, `KeyValueStoreCausale.Get`, 
`KeyValueStoreCausale.Delete`.

Ciascuna di queste funzioni RPC, lato server, una volta ricevuta la richiesta incrementano il clock vettoriale 
associato al server (protetto da un mutex), lo associa alla richiesta ricevuta e genera un messaggio da inviare 
in multicast, tramite `sendToAllServer()`.

Ciascun server che riceve un messaggio di multicast, in `CausallyOrderedMulticast()`, aggiunge la richiesta alla coda e
controlla (ogni 100 Millisecondi) se può essere eseguita a livello applicativo.

Il controllo avviene in `controlSendToApplication()`, quando il processo `pj` riceve il messaggio `m` da `pi`, dove per
`t(m)` si intende il clock vettoriale associato al messaggio inviato:
- `t(m)[i] = Vj[i] + 1` (il messaggio `m` è il successivo che `pj` si aspetta da `pi`).
- `t(m)[k] ≤ Vj[k]` per ogni processo `pk` diverso da `i` (ovvero `pj` ha visto almeno gli stessi messaggi di `pk` visti
da `pi`).
- Viene ulteriormente controllato se il messaggio ricevuto è stato inviato da se stesso, in quel caso è sicuramente un 
"messaggio che si aspetta di ricevere".

In caso il controllo vada a buon fine:
- Se non si è il sender del messaggio, si incrementa il clock relativo al server che ha inviato il messaggio.
- Viene rimosso il messaggio dalla coda.
- è possibile inviare il messaggio a livello applicativo eseguendo `realFunction()` che esegue l'operazione richiesta
nel datastore del server. 

**Multicast Totalmente Ordinato**

Il client effettua le richieste RPC chiamando le `KeyValueStoreSequential.Put`, `KeyValueStoreSequential.Get`, 
`KeyValueStoreSequential.Delete`.

Ciascuna di queste funzioni RPC, lato server, una volta ricevuta la richiesta incrementano il suo clock logico scalare 
(protetto da un mutex):
- Per le richieste di tipo `GET` viene eseguita direttamente la richiesta e restituito il risultato al client.
- Per le richieste di tipo `PUT` e `DELETE` viene generato un messaggio da inviare in multicast, tramite `sendToAllServer()`.

Ciascun server che riceve un messaggio di multicast (da una richiesta di `PUT` o `DELETE`), in `TotalOrderedMulticast()`:
- Aggiunge la richiesta alla coda, ordinata in base al timestamp locale.
- Invia un ack a tutti i server per indicare che lui ha letto quel messaggio, tramite `sendAck()`.
- Controlla se può essere eseguita a livello applicativo.

Il controllo avviene in `controlSendToApplication()`: verificando se il messaggio è in testa alla coda e ha ricevuto 
tutti gli ack relativi a quella richiesta. 
In caso il controllo vada a buon fine è possibile inviare il messaggio a livello applicativo:
- Viene rimosso il messaggio dalla coda 
- Calcola il max tra il suo clock scalare e il clock scalare del messaggio ricevuto.
- Incrementa di uno il clock scalare, per conteggiare l'evento 
- Viene eseguita `realFunction()` che esegue l'operazione richiesta nel datastore del server. 

`sendAck()` invia un messaggio di ack a tutti i server, in modo asincrono.
Viene effettuata una chiamata RPC `ReceiveAck()`, che gestisce la ricezione dell'ack, se fa riferimento ad un messaggio 
che il server ha già ricevuto, viene incrementato il contatore degli ack ricevuti, 
se non è un messaggio che il server ha già ricevuto, viene impostata la risposta a false.
In sendAck() viene controllato il valore della risposta, se è false, viene inviato nuovamente l'ack.

**Test**
L'output del client mostra in blu, con il termine `RUN operation` le richieste ai server replica secondo l'ordine in cui
vengono eseguite. 