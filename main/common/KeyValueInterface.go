package common

// Args rappresenta gli argomenti delle chiamate RPC
type Args struct {
	Key       string
	Value     string
	Timestamp int // Chiedere ulteriori spiegazioni
}

// Response è una struttura creata per memorizzare la risposta delle chiamate RPC
type Response struct {
	Value  string
	Result bool
	Done   chan bool
}

// KeyValueStoreService è un'interfaccia rappresentante che chiamate RPC esposte al client
type KeyValueStoreService interface {
	Get(args Args, reply *Response) error
	Put(args Args, reply *Response) error
	Delete(args Args, reply *Response) error
}
