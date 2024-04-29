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

/* ARGS STRUCT */

func newArgs() Args {
	return Args{}
}

func (args *Args) SetKey(key string) {
	args.Key = key
}

func (args *Args) SetValue(value string) {
	args.Value = value
}

func (args *Args) SetTimestamp(timestamp int) {
	args.Timestamp = timestamp
}

func (args *Args) GetKey() string {
	return args.Key
}

func (args *Args) GetValue() string {
	return args.Value
}

func (args *Args) GetTimestamp() int {
	return args.Timestamp
}

/* RESPONDE STRUCT */

func (response *Response) SetValue(value string) {
	response.Value = value
}

func (response *Response) SetResult(result bool) {
	response.Result = result
}

func (response *Response) SetDone(done chan bool) {
	response.Done = done
}

func (response *Response) GetValue() string {
	return response.Value
}

func (response *Response) GetResult() bool {
	return response.Result
}

func (response *Response) GetDone() chan bool {
	return response.Done
}
