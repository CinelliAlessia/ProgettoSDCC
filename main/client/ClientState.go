package main

import (
	"main/common"
	"sync"
)

// ClientsState Rappresenta lo stato dei clients, tutte le variabili sono vettori perché ciascun indice rappresenta il singolo client
type ClientsState struct {
	MutexSentMsgCounter []sync.Mutex
	SentMsgCounter      []int // SentMsgCounter rappresenta l'indice dei messaggi inviati al server, usato in syncCall e asyncCall

	MutexReceiveMsgCounter []sync.Mutex
	ReceiveMsgCounter      []int // ReceiveMsgCounter rappresenta l'ordinamento con cui i messaggi sono stati eseguiti dal server

	MutexFirstRequest sync.Mutex
	FirstRequest      bool // FirstRequest è una variabile booleana che indica se è la prima richiesta o meno,
	// cosi da resettare i campi al cambio di consistenza scelta

	ListArgs []common.Args // ListArgs è un array di argomenti ciascuna per un client

	MutexSendingTS []sync.Mutex
	SendingTS      []int // SendingTS associa a ogni messaggio un timestamp di invio,
	// allegato successivamente ad args. Per assunzione FIFO Ordering in andata
}

// NewClientState Stiamo assumendo che avrò tanti client quanti server common.Replicas
func NewClientState() *ClientsState {
	return &ClientsState{
		SentMsgCounter:      make([]int, common.Replicas), // Tanti quanti sono i server
		MutexSentMsgCounter: make([]sync.Mutex, common.Replicas),

		ReceiveMsgCounter:      make([]int, common.Replicas), // Tanti quanti sono i server
		MutexReceiveMsgCounter: make([]sync.Mutex, common.Replicas),

		FirstRequest: true,
		ListArgs:     make([]common.Args, common.Replicas), // Tanti quanti sono i client

		MutexSendingTS: make([]sync.Mutex, common.Replicas),
		SendingTS:      make([]int, common.Replicas), // Tanti quanti sono i client
	}
}

// ----- SentMsgCounter ----- // Indice dei messaggi inviati al server

// SetSentMsgCounter imposta l'indice dei messaggi inviati al server, protetto da mutexSentMsgCounter
func (clientState *ClientsState) SetSentMsgCounter(index int, value int) {
	clientState.LockMutexSentMsg(index)
	defer clientState.UnlockMutexSentMsg(index)
	clientState.SentMsgCounter[index] = value
}

// IncreaseSentMsg incrementa l'indice dei messaggi inviati al server, non protetto da mutexSentMsgCounter
func (clientState *ClientsState) IncreaseSentMsg(index int) {
	clientState.SentMsgCounter[index]++
}

// GetSentMsgCounter restituisce l'indice dei messaggi inviati al server, protetto da mutexSentMsgCounter
func (clientState *ClientsState) GetSentMsgCounter(index int) int {
	clientState.LockMutexSentMsg(index)
	defer clientState.UnlockMutexSentMsg(index)

	return clientState.SentMsgCounter[index]
}

func (clientState *ClientsState) LockMutexSentMsg(index int) {
	clientState.MutexSentMsgCounter[index].Lock()
}

func (clientState *ClientsState) UnlockMutexSentMsg(index int) {
	clientState.MutexSentMsgCounter[index].Unlock()
}

// ----- ReceiveMsgCounter ----- // Ordinamento con cui i messaggi sono stati eseguiti dal server, come il client
// Dovrebbe leggerli (e mostrarli a schermo)

// IncreaseReceiveMsg Incrementa l'indice dei messaggi ricevuti dal client, non protetto da mutexReceive
func (clientState *ClientsState) IncreaseReceiveMsg(index int) {
	clientState.ReceiveMsgCounter[index]++
}

// GetReceiveMsgCounter restituisce l'indice dei messaggi ricevuti dal client, protetto da mutexReceive
func (clientState *ClientsState) GetReceiveMsgCounter(index int) int {
	clientState.MutexReceiveLock(index)
	defer clientState.MutexReceiveUnlock(index)

	return clientState.ReceiveMsgCounter[index]
}

func (clientState *ClientsState) MutexReceiveLock(index int) {
	clientState.MutexReceiveMsgCounter[index].Lock()
}

func (clientState *ClientsState) MutexReceiveUnlock(index int) {
	clientState.MutexReceiveMsgCounter[index].Unlock()
}

// ----- FirstRequest ----- // Indica se è la prima richiesta o meno, cosi da resettare i campi

// SetFirstRequest imposta la variabile booleana FirstRequest, protetta da mutexFirstRequest
func (clientState *ClientsState) SetFirstRequest(firstRequest bool) {
	clientState.MutexFirstRequest.Lock()
	defer clientState.MutexFirstRequest.Unlock()

	clientState.FirstRequest = firstRequest
}

// GetFirstRequest restituisce la variabile booleana FirstRequest, protetta da mutexFirstRequest
func (clientState *ClientsState) GetFirstRequest() bool {
	clientState.MutexFirstRequest.Lock()
	defer clientState.MutexFirstRequest.Unlock()
	return clientState.FirstRequest
}

// ----- ListArgs ----- // Array di argomenti ciascuna per un client

func (clientState *ClientsState) SetListArgs(index int, args common.Args) {
	clientState.ListArgs[index] = args
}

func (clientState *ClientsState) GetListArgs(index int) common.Args {
	return clientState.ListArgs[index]
}

// ----- SendingTS ----- // Timestamp di invio associato dal client alla creazione dell'operazione,
// come i messaggi verranno inviati dal client al server

// IncreaseSendingTS incrementa di uno, il contatore dei messaggi da inviare al server.
//
//	SendingTS indica l'ordinamento con cui il client invia le operazioni al server.
//	Andrà allegato successivamente ad args, per assunzione FIFO Ordering in andata
//	Non protetto da mutex
func (clientState *ClientsState) IncreaseSendingTS(index int) {
	clientState.SendingTS[index]++
}

// GetSendingTS restituisce il contatore dei messaggi da inviare al server.
//
//	SendingTS indica l'ordinamento con cui il client invia le operazioni al server.
//	Non protetto da mutex
func (clientState *ClientsState) GetSendingTS(index int) int {
	return clientState.SendingTS[index]
}
