- Aggiunta dipendenza esterna import "github.com/google/uuid" in KeyValueStoreSequential

**Garanzie Consistenza**

Questo progetto consiste nel realizzare un server che dia garanzia, a scelta del client, di avere repliche di datatstore che rispettino la consistenza sequenziale o la consistenza causale. 

- Per garantire la consistenza sequenziale si utilizza l'algoritmo di multicast totalmente ordinato
- Per garantire la consistenza causale si utilizza l'algoritmo di multicast causalmente ordinato

***SEQUENZIALE:***
*Multicast Totalmente Ordinato*

1. Un client effettua una chiamata RPC con la propria richiesta ad un singolo server (scelto random).
2. Il server che riceve la richiesta:
   1. Se è un evento interno (GET) incrementa il suo clock logico scalare ed esegue la richiesta rispondendo al client semplicemente. 
   2. Se è un evento esterno (PUT o DELETE) incrementa il suo clock logico scalare e inoltra la richiesta del client a tutti gli altri server, entriamo più nel dettaglio:

**Gestione degli eventi di PUT e DELETE**
1. Il server ricevente la richiesta dal client inoltra la tutti i server (incluso se stesso).
2. Ciascun server che riceve la richiesta:
   1. La aggiunge a una coda locale ordinata per timestamp della richiesta.
   2. Invia un ack a tutti i server per indicare che lui ha letto quel messaggio.
   3. Esegue la richiesta ricevuta esclusivamente se, la richiesta è la prima nella sua coda (ha timestamp minore di tutte le altre) AND ha ricevuto tutti gli ack relativi a quella richiesta,