package main

// In questo file dovrebbero esserci i servizi esposti dal server per il clint, che sono "Fake" poiché il server in
// realtà deve eseguire l'algoritmo multicast totalmente ordinato.
// Ma è fake solo per le operazioni di lettura

type Args struct {
	key   string
	value string
}

type Response struct {
	reply string
}

type KeyValueStoreService interface {
	Get(args Args, reply *Response) error
	Put(args Args, reply *Response) error
	Delete(args Args, reply *Response) error
}
