# Sistema di memorizzazione chiave-valore con garanzie di consistenza Sequenziale e Causale
## Start

### *Esecuzione in locale o con Docker-Compose:*

Per eseguire il sistema in locale o tramite Docker-Compose, è sufficiente eseguire il comando:

`./start.sh` in /ProgettoSDCC/main 

La scelta tra un ambiente locale o Docker-Compose è possibile modificando la variabile
d'ambiente `CONFIG` all'interno del file "start.sh".

### *Esecuzione con istanze EC2:*

Per connettersi ad un istanza EC2 via SSH, è necessario utilizzare il comando:

`ssh -i "chiave-privata.pem" ec2-user@address`

*Installare docker*

sudo yum update -y

sudo yum install -y docker

*Installare docker-compose*

sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose

sudo chmod +x /usr/local/bin/docker-compose

*Run docker demon*

sudo service docker start

*Installa git e clona repository*

sudo yum install git -y

git clone https://github.com/CinelliAlessia/ProgettoSDCC.git

*Run docker compose:*

cd ProgettoSDCC/main/

sudo docker-compose up --build

*Per vedere i container attivi*

sudo docker ps

*Per entrare in un container*

sudo docker exec -it main_client_1 /bin/bash

*Per uscire da un container Docker senza interromperlo:*

Utilizzare la combinazione di tasti Ctrl + P seguita da Ctrl + Q

*Per arrestare i container:*

sudo docker stop <nome-container>

## **Garanzie Consistenza**
Questo progetto consiste nel realizzare un server che dia garanzia, a scelta del client, di avere repliche di datastore
che rispettino la consistenza sequenziale o la consistenza causale.

- Per garantire la consistenza causale si utilizza l'algoritmo di Multicast Causalmente Ordinato.
- Per garantire la consistenza sequenziale si utilizza l'algoritmo di Multicast Totalmente Ordinato.

### **Multicast Causalmente Ordinato**

Il client effettua le richieste RPC chiamando le `KeyValueStoreCausale.Put`, `KeyValueStoreCausale.Get`,
`KeyValueStoreCausale.Delete`.

Per ciascuna di queste funzioni RPC, lato server, una volta ricevuta la richiesta controlla se può essere accettata
o se verrà bufferizzata, per rispettare l'assunzione di comunicazione FIFO Ordered.
Una volta che la richiesta può essere ricevuta, il server incrementa il suo indice nel clock vettoriale per 
conteggiare l'evento di receive, lo associa alla richiesta e genera un messaggio da inviare in Multicast, tramite `sendToAllServer()`.

Ciascun server che riceve un messaggio di multicast, in `Update()`, aggiunge la richiesta alla sua coda locale e
controllerà, tramite la funzione canExecute() se può essere eseguita a livello applicativo, altrimenti il messaggio viene
bufferizzato e la goroutine si metterà in attesa con una variabile condizionale. Verrà sbloccata univocamente quando il 
messaggio potrà essere eseguito a livello applicativo.

Il controllo avviene in `controlSendToApplication()`, quando il processo `pj` riceve il messaggio `m` da `pi`, dove per
`t(m)` si intende il clock vettoriale associato al messaggio inviato:
- `t(m)[i] = Vj[i] + 1` (il messaggio `m` è il successivo che `pj` si aspetta da `pi`).
- `t(m)[k] ≤ Vj[k]` per ogni processo `pk` diverso da `i` (ovvero `pj` ha visto almeno gli stessi messaggi di `pk` visti
  da `pi`).
- Si controlla se il messaggio ricevuto è stato inviato da se stesso, in quel caso è sicuramente un
  messaggio che "si aspetta di ricevere".
- Un ulteriore controllo si effettua nel caso sia una richiesta di `Get`, nel caso un client ha effettuato una richiesta
  di lettura di una key non presente nel datastore, questa richiesta verrà eseguita solo quando il server avrà ricevuto
  almeno una richiesta di `Put` per quella Key, cioè la chiave dovrà essere presente nel datastore.

In caso il controllo vada a buon fine:
- Se non si è il sender del messaggio, si incrementa il clock relativo al server che ha inviato il messaggio, per
  conteggiare l'evento di receive.
- Viene rimosso il messaggio dalla coda.
- è possibile inviare il messaggio a livello applicativo eseguendo `realFunction()` che esegue l'operazione richiesta
  nel datastore del server.

### **Multicast Totalmente Ordinato**

Il client effettua le richieste RPC chiamando le `KeyValueStoreSequential.Put`, `KeyValueStoreSequential.Get`,
`KeyValueStoreSequential.Delete`.

Il Multicast Totalmente Ordinato, per essere realizzato ha bisogno di una assunzione:
- I messaggi vengono consegnati al server nello stesso ordine in cui il client lo invia. Assunzione FIFO Ordering
- Comunicazione affidabile, no perdita di messaggi.

Il server, alla ricezione di qualsiasi delle tre chiamate RPC, per rispettare l'assunzione di FIFO Ordering, esegue la
funzione `canReceive()`, controllando se ha ricevuto tutti i messaggi precedenti a quello attuale da parte del client,
continuerà l'esecuzione solo se ha ricevuto tutti i messaggi precedenti.

Ciascuna di queste funzioni RPC, lato server, una volta ricevuta la richiesta incrementano il suo clock logico scalare
(protetto da un mutex) e aggiungono il messaggio creato alla coda:
- Per le richieste di tipo `GET`, se il messaggio relativo è in testa alla coda, verrà eseguita direttamente la
  richiesta e restituito il risultato al client, poiché le operazioni di lettura sono considerate eventi interni,
  e non c'è bisogno che tutti i datastore replica ne siano a conoscenza.
- Per le richieste di tipo `PUT` e `DELETE` viene generato un messaggio da inviare in multicast, tramite
  `sendToAllServer()`, essendo essi eventi esterni, che vanno a modificare tutti i datastore replica.

Ciascun server che riceve un messaggio di multicast (da una richiesta di `PUT` o `DELETE`), in `Update()`:
- Invia un ack a tutti i server per indicare che lui ha letto quel messaggio, tramite `sendAck()`.
- Controlla se può essere eseguita a livello applicativo, tramite `canExecute()`.

Il controllo avviene in `controlSendToApplication()`: verificando se il messaggio è in testa alla coda e ha ricevuto
tutti gli ack da parte dei server.
In caso il controllo vada a buon fine è possibile inviare il messaggio a livello applicativo, l'esecuzione è protetta da
un mutex per garantirne un esecuzione atomica:
- Viene rimosso il messaggio dalla coda
- Calcola il max tra il suo clock scalare e il clock scalare del messaggio ricevuto.
- Incrementa di uno il clock scalare, per conteggiare l'evento
- Viene eseguita `realFunction()` che esegue l'operazione richiesta nel datastore del server.

`realFunction()` esegue l'operazione richiesta nel datastore del server, in caso di `PUT` e `DELETE` viene aggiornato il
datastore replica, in caso di `GET` viene restituito il valore associato alla chiave richiesta.
Gestisce l'assunzione FIFO Ordering per i messaggi di risposta, impostando un timestamp per l'ordinamento dei messaggi.

`sendAck()` invia un messaggio di ack a tutti i server, in modo asincrono.
Viene effettuata una chiamata RPC `ReceiveAck()`, che gestisce la ricezione dell'ack, se fa riferimento ad un messaggio
che il server ha già ricevuto, viene incrementato il contatore degli ack ricevuti,
se non è un messaggio che il server ha già ricevuto, viene impostata la risposta a false.
In sendAck() viene controllato il valore della risposta, se è false, viene inviato nuovamente l'ack.

### **Dipendenze Esterne** 
Per garantire il corretto funzionamento del sistema, sono state aggiunte alcune dipendenze 
esterne che facilitano determinate funzionalità all'interno del codice. 
In particolare, abbiamo incluso le seguenti dipendenze:

*   **github.com/google/uuid**: Questa libreria è stata utilizzata per generare in modo univoco gli identificatori univoci (UUID) necessari per identificare in modo univoco le chiavi all'interno del nostro sistema di memorizzazione chiave-valore distribuito.
*   **github.com/fatih/color**: Prima di poter utilizzare questa libreria per la gestione dei colori, è stata eseguita una procedura di installazione tramite il comando "go get -u github.com/fatih/color". Questa libreria ci permette di migliorare la leggibilità del nostro output a schermo, fornendo colori distintivi per determinati messaggi o output.

### Variabili d'ambiente nel file start.sh
- `CONFIG`: Variabile d'ambiente per la scelta tra un ambiente locale o Docker-Compose.
- `DEBUG`: Variabile d'ambiente per la scelta tra una spiegazione esplicita o meno.
- `USEENDKEY`: Variabile d'ambiente per la scelta tra l'utilizzo di una chiave finale o meno nella consistenza sequenziale.