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


*PROBLEMA*
- Mi arriva prima un ack e dopo la richiesta di append => Ignoro e rispondo "false" cosi che mi venga re-inviata.
- 

