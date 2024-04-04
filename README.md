*DIPENDENZE*
- Aggiunta dipendenza esterna import "github.com/google/uuid" in KeyValueStoreSequential

*LOCALE e DOCKER*
- Per variare da utilizzo locale e utilizzo docker cambiare unicamente la variabile globale nel file start.sh

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
   2. La esegue se la richiesta è la prima nella coda. 

*Gestione degli eventi esterni: PUT e DELETE*
   1. Inoltra in broadcast la richiesta a tutti i server (incluso se stesso).
   2. Ciascun server che riceve la richiesta:
      1. La aggiunge a una coda locale ordinata per timestamp della richiesta.
      2. Invia un ack a tutti i server per indicare che lui ha letto quel messaggio.
      3. Esegue la richiesta ricevuta esclusivamente se: 
         - La richiesta è la prima nella sua coda (ha timestamp minore di tutte le altre in coda).
         - Sono stati ricevuti tutti gli ack relativi a quella richiesta.

- L'invio dell'ack a tutti i server avviene in maniera asincrona.

***CAUSALE:***

**Multicast Causalmente Ordinato**

