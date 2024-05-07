package main

import (
	"main/common"
	"sync"
)

type ClientState struct {
	SendIndex    []int
	MutexSent    []sync.Mutex
	ReceiveIndex []int
	MutexReceive []sync.Mutex
}

func NewClientState() *ClientState {
	return &ClientState{
		SendIndex:    make([]int, common.Replicas),
		MutexSent:    make([]sync.Mutex, common.Replicas),
		ReceiveIndex: make([]int, common.Replicas),
		MutexReceive: make([]sync.Mutex, common.Replicas),
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
