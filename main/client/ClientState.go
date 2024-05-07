package main

import (
	"main/common"
	"sync"
)

// ClientState Rappresenta lo stato del client, tutte le variabili sono vettori perché ciascun indice rappresenta il singolo client
type ClientState struct {
	SendIndex []int // SendIndex rappresenta l'indice dei messaggi da inviare al server, usato in syncCall e asyncCall
	MutexSent []sync.Mutex

	ReceiveIndex []int // ReceiveIndex rappresenta l'ordinamento con cui i messaggi sono stati eseguiti dal server per un corretto print
	MutexReceive []sync.Mutex

	FirstRequest bool // FirstRequest è una variabile booleana che indica se è la prima richiesta o meno, cosi da resettare i campi
	// Sarebbe da usare in caso in cui venga cliccato Exit per passare da una consistenza all'altra

	ListArgs []common.Args // ListArgs è un array di argomenti ciascuna per un client

	SendingTS []int // SendingTS associa a ogni messaggio un timestamp di invio, allegato successivamente ad args. Per assunzione FIFO Ordering in andata
}

// NewClientState Stiamo assumendo che avrò tanti client quanti server common.Replicas
func NewClientState() *ClientState {
	return &ClientState{
		SendIndex: make([]int, common.Replicas), // Tanti quanti sono i server
		MutexSent: make([]sync.Mutex, common.Replicas),

		ReceiveIndex: make([]int, common.Replicas), // Tanti quanti sono i server
		MutexReceive: make([]sync.Mutex, common.Replicas),

		FirstRequest: true,
		ListArgs:     make([]common.Args, common.Replicas), // Tanti quanti sono i client
		SendingTS:    make([]int, common.Replicas),         // Tanti quanti sono i client
	}
}

func (clientState *ClientState) IncreaseSendIndex(index int) {
	clientState.MutexSent[index].Lock()
	clientState.SendIndex[index]++
	clientState.MutexSent[index].Unlock()
}

func (clientState *ClientState) GetSendIndex(index int) int {
	return clientState.SendIndex[index]
}

func (clientState *ClientState) IncreaseReceiveIndex(index int) {
	clientState.MutexReceive[index].Lock()
	clientState.ReceiveIndex[index]++
	clientState.MutexReceive[index].Unlock()
}

func (clientState *ClientState) GetReceiveIndex(index int) int {
	return clientState.ReceiveIndex[index]
}

func (clientState *ClientState) SetFirstRequest(firstRequest bool) {
	clientState.FirstRequest = firstRequest
}

func (clientState *ClientState) GetFirstRequest() bool {
	return clientState.FirstRequest
}

func (clientState *ClientState) SetListArgs(index int, args common.Args) {
	clientState.ListArgs[index] = args
}

func (clientState *ClientState) GetListArgs(index int) common.Args {
	return clientState.ListArgs[index]
}

// IncreaseSendingTS associa a ogni messaggio un timestamp di invio,
// ce andrà allegato successivamente ad args. Per assunzione FIFO Ordering in andata
func (clientState *ClientState) IncreaseSendingTS(index int) {
	clientState.SendingTS[index]++
}

func (clientState *ClientState) GetSendingTS(index int) int {
	return clientState.SendingTS[index]
}
