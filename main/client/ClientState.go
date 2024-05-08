package main

import (
	"main/common"
	"sync"
)

// ClientState Rappresenta lo stato del client, tutte le variabili sono vettori perché ciascun indice rappresenta il singolo client
type ClientState struct {
	MutexSend []sync.Mutex
	SendIndex []int // SendIndex rappresenta l'indice dei messaggi inviati al server,
	// usato in syncCall e asyncCall

	MutexReceive []sync.Mutex
	ReceiveIndex []int // ReceiveIndex rappresenta l'ordinamento con cui i messaggi sono stati eseguiti
	// dal server per un corretto print

	MutexFirstRequest sync.Mutex
	FirstRequest      bool // FirstRequest è una variabile booleana che indica se è la prima richiesta o meno,
	// cosi da resettare i campi

	ListArgs []common.Args // ListArgs è un array di argomenti ciascuna per un client

	MutexSendingTS []sync.Mutex
	SendingTS      []int // SendingTS associa a ogni messaggio un timestamp di invio,
	// allegato successivamente ad args. Per assunzione FIFO Ordering in andata
}

// NewClientState Stiamo assumendo che avrò tanti client quanti server common.Replicas
func NewClientState() *ClientState {
	return &ClientState{
		SendIndex: make([]int, common.Replicas), // Tanti quanti sono i server
		MutexSend: make([]sync.Mutex, common.Replicas),

		ReceiveIndex: make([]int, common.Replicas), // Tanti quanti sono i server
		MutexReceive: make([]sync.Mutex, common.Replicas),

		FirstRequest: true,
		ListArgs:     make([]common.Args, common.Replicas), // Tanti quanti sono i client
		SendingTS:    make([]int, common.Replicas),         // Tanti quanti sono i client
	}
}

// ----- SendIndex ----- // Indice dei messaggi inviati al server

// SetSendIndex imposta l'indice dei messaggi inviati al server, protetto da mutexSend
func (clientState *ClientState) SetSendIndex(index int, value int) {
	clientState.MutexSend[index].Lock()
	clientState.SendIndex[index] = value
	clientState.MutexSend[index].Unlock()
}

// IncreaseSendIndex incrementa l'indice dei messaggi inviati al server, non protetto da mutexSend
func (clientState *ClientState) IncreaseSendIndex(index int) {
	clientState.SendIndex[index]++
}

// GetSendIndex restituisce l'indice dei messaggi inviati al server, protetto da mutexSend
func (clientState *ClientState) GetSendIndex(index int) int {
	clientState.MutexSendLock(index)
	defer clientState.MutexSendUnlock(index)

	return clientState.SendIndex[index]
}

func (clientState *ClientState) MutexSendLock(index int) {
	clientState.MutexSend[index].Lock()
}

func (clientState *ClientState) MutexSendUnlock(index int) {
	clientState.MutexSend[index].Unlock()
}

// ----- ReceiveIndex ----- // Ordinamento con cui i messaggi sono stati eseguiti dal server

// IncreaseReceiveIndex Incrementa l'indice dei messaggi ricevuti dal client, non protetto da mutexReceive
func (clientState *ClientState) IncreaseReceiveIndex(index int) {
	clientState.ReceiveIndex[index]++
}

// GetReceiveIndex restituisce l'indice dei messaggi ricevuti dal client, protetto da mutexReceive
func (clientState *ClientState) GetReceiveIndex(index int) int {
	clientState.MutexReceiveLock(index)
	defer clientState.MutexReceiveUnlock(index)

	return clientState.ReceiveIndex[index]
}

func (clientState *ClientState) MutexReceiveLock(index int) {
	clientState.MutexReceive[index].Lock()
}

func (clientState *ClientState) MutexReceiveUnlock(index int) {
	clientState.MutexReceive[index].Unlock()
}

// ----- FirstRequest ----- // Indica se è la prima richiesta o meno, cosi da resettare i campi

// SetFirstRequest imposta la variabile booleana FirstRequest, protetta da mutexFirstRequest
func (clientState *ClientState) SetFirstRequest(firstRequest bool) {
	clientState.MutexFirstRequest.Lock()
	clientState.FirstRequest = firstRequest
	clientState.MutexFirstRequest.Unlock()
}

// GetFirstRequest restituisce la variabile booleana FirstRequest, protetta da mutexFirstRequest
func (clientState *ClientState) GetFirstRequest() bool {
	clientState.MutexFirstRequest.Lock()
	defer clientState.MutexFirstRequest.Unlock()
	return clientState.FirstRequest
}

// ----- ListArgs ----- // Array di argomenti ciascuna per un client

func (clientState *ClientState) SetListArgs(index int, args common.Args) {
	clientState.ListArgs[index] = args
}

func (clientState *ClientState) GetListArgs(index int) common.Args {
	return clientState.ListArgs[index]
}

// ----- SendingTS ----- // Timestamp di invio associato dal client alla creazione dell'operazione

// IncreaseSendingTS associa a ogni messaggio un timestamp di invio,
// ce andrà allegato successivamente ad args. Per assunzione FIFO Ordering in andata
// Non protetto da mutex
func (clientState *ClientState) IncreaseSendingTS(index int) {
	clientState.SendingTS[index]++
}

func (clientState *ClientState) GetSendingTS(index int) int {
	return clientState.SendingTS[index]
}
