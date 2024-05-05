package common

// KeyValueStoreRequest è un'interfaccia rappresentante che chiamate RPC esposte al client
type KeyValueStoreRequest interface {
	Get(args Args, reply *Response) error
	Put(args Args, reply *Response) error
	Delete(args Args, reply *Response) error
}
