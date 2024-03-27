package common

// In questo file dovrebbero esserci i servizi esposti dal server per il clint, che sono "Fake" poiché il server in
// realtà deve eseguire l'algoritmo multicast totalmente ordinato.
// Ma è fake solo per le operazioni di lettura

type Args struct {
	Key   string
	Value string
}

type Response struct {
	Reply string
}

// KeyValueStoreService è un'interfaccia che deve essere implementata da tutti i servizi che vogliono essere
type KeyValueStoreService interface {
	Get(args Args, reply *Response) error
	Put(args Args, reply *Response) error
	Delete(args Args, reply *Response) error
}
