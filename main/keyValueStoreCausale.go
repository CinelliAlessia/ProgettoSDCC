package main

import (
	"sync"
)

// KeyValueStoreCausale rappresenta il servizio di memorizzazione chiave-valore
type KeyValueStoreCausale struct {
	data        map[string]string // Mappa -> struttura dati che associa chiavi a valori
	vectorClock []int             // Orologio vettoriale
	mutex       sync.Mutex
}

// Get restituisce il valore associato alla chiave specificata -> Lettura -> Evento interno
func (kvs *KeyValueStoreCausale) Get(args Args, response *Response) error {
	return nil
}

// Put inserisce una nuova coppia chiave-valore, se la chiave è già presente, sovrascrive il valore associato
func (kvs *KeyValueStoreCausale) Put(args Args, response *Response) error {
	return nil
}

// Delete elimina una coppia chiave-valore, è un operazione di scrittura
func (kvs *KeyValueStoreCausale) Delete(args Args, response *Response) error {
	return nil
}
