package common

// Response è una struttura creata per memorizzare la risposta delle chiamate RPC
type Response struct {
	Value  string
	Result bool
	Done   chan bool
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
